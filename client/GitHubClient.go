/*
 * Copyright (c) 2020-2024. Devtron Inc.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package util

import (
	"context"
	"github.com/caarlos0/env"
	"github.com/google/go-github/github"
	"go.uber.org/zap"
	"golang.org/x/oauth2"
	http2 "net/http"
	"net/url"
	"path"
)

const (
	GIT_WORKING_DIR       = "/tmp/gitops/"
	GetRepoUrlStage       = "Get Repo Url"
	CreateRepoStage       = "Create Repo"
	CloneHttpStage        = "Clone Http"
	CreateReadmeStage     = "Create Readme"
	CloneSshStage         = "Clone Ssh"
	GITLAB_PROVIDER       = "GITLAB"
	GITHUB_PROVIDER       = "GITHUB"
	AZURE_DEVOPS_PROVIDER = "AZURE_DEVOPS"
	BITBUCKET_PROVIDER    = "BITBUCKET_CLOUD"
	GITHUB_API_V3         = "api/v3"
	GITHUB_HOST           = "github.com"
)

type GitConfig struct {
	GitlabGroupId        string //local
	GitlabGroupPath      string //local
	GitToken             string //not null  // public
	GitUserName          string //not null  // public
	GitWorkingDir        string //working directory for git. might use pvc
	GithubOrganization   string
	GitProvider          string // SUPPORTED VALUES  GITHUB, GITLAB
	GitHost              string
	AzureToken           string
	AzureProject         string
	BitbucketWorkspaceId string
	BitbucketProjectKey  string
}

type GitHubConfig struct {
	GitHubHost  string `env:"GITHUB_HOST" envDefault:"https://github.com"`
	GitHubOrg   string `env:"GITHUB_ORG" envDefault:""`
	GitHubToken string `env:"GITHUB_TOKEN" envDefault:""`
	GitHubRepo  string `env:"GITHUB_REPO" envDefault:"devtron"`

	GitHubWebhookSecret   string `env:"GITHUB_WEBHOOK_SECRET" envDefault:""`
	GitHubEventTypeHeader string `env:"GITHUB_EVENT_TYPE_HEADER" envDefault:"X-GitHub-Event"`
	GitHubSecretHeader    string `env:"GITHUB_SECRET_HEADER" envDefault:"X-Hub-Signature"`
	GitHubSecretValidator string `env:"GITHUB_SECRET_VALIDATOR" envDefault:"SHA-1"`
}

type GitHubClient struct {
	GitHubClient *github.Client
	GitHubConfig *GitHubConfig
}

/* #nosec */
func NewGitHubClient(logger *zap.SugaredLogger) (*GitHubClient, error) {
	cfg := &GitHubConfig{}
	err := env.Parse(cfg)
	if err != nil {
		logger.Error("err", err)
		return &GitHubClient{}, err
	}
	ctx := context.Background()
	httpTransport := &http2.Transport{}
	httpClient := &http2.Client{Transport: httpTransport}
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: cfg.GitHubToken},
	)
	ctx = context.WithValue(ctx, oauth2.HTTPClient, httpClient)
	tc := oauth2.NewClient(ctx, ts)
	var client *github.Client
	hostUrl, err := url.Parse(cfg.GitHubHost)
	if err != nil {
		logger.Errorw("error in creating git client ", "host", hostUrl, "err", err)
		return nil, err
	}
	if hostUrl.Host == GITHUB_HOST {
		client = github.NewClient(tc)
	} else {
		logger.Infow("creating github EnterpriseClient with org", "host", cfg.GitHubHost, "org", cfg.GitHubOrg)
		hostUrl.Path = path.Join(hostUrl.Path, GITHUB_API_V3)
		client, err = github.NewEnterpriseClient(hostUrl.String(), hostUrl.String(), tc)
	}
	gitHubClient := &GitHubClient{
		GitHubClient: client,
		GitHubConfig: cfg,
	}
	return gitHubClient, err
}
