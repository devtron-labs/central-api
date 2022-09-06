package pkg

import (
	"context"
	"encoding/json"
	"fmt"
	util "github.com/devtron-labs/central-api/client"
	"github.com/devtron-labs/central-api/common"
	"github.com/google/go-github/github"
	"github.com/patrickmn/go-cache"
	"go.uber.org/zap"
	"strings"
	"sync"
	"time"
)

type ReleaseNoteService interface {
	GetModules() ([]*common.Module, error)
	GetReleases() ([]*common.Release, error)
	UpdateReleases(requestBodyBytes []byte) (bool, error)
	GetModulesV2() ([]*common.Module, error)
	GetModuleByName(name string) (*common.Module, error)
}

type ReleaseNoteServiceImpl struct {
	logger       *zap.SugaredLogger
	client       *util.GitHubClient
	releaseCache *util.ReleaseCache
	mutex        sync.Mutex
	moduleConfig *util.ModuleConfig
}

func NewReleaseNoteServiceImpl(logger *zap.SugaredLogger, client *util.GitHubClient, releaseCache *util.ReleaseCache,
	moduleConfig *util.ModuleConfig) *ReleaseNoteServiceImpl {
	serviceImpl := &ReleaseNoteServiceImpl{
		logger:       logger,
		client:       client,
		releaseCache: releaseCache,
		moduleConfig: moduleConfig,
	}
	_, err := serviceImpl.GetReleases()
	if err != nil {
		serviceImpl.logger.Errorw("error on app init call for releases", "err", err)
		//ignore error for starting application
	}
	return serviceImpl
}

const ActionPublished = "published"
const ActionEdited = "edited"
const EventTypeRelease = "release"
const TimeFormatLayout = "2006-01-02T15:04:05Z"
const TagLink = "https://github.com/devtron-labs/devtron/releases/tag"
const PrerequisitesMatcher = "<!--upgrade-prerequisites-required-->"

func (impl *ReleaseNoteServiceImpl) UpdateReleases(requestBodyBytes []byte) (bool, error) {
	data := make(map[string]interface{})
	err := json.Unmarshal(requestBodyBytes, &data)
	if err != nil {
		impl.logger.Errorw("unmarshal error", "err", err)
		return false, err
	}
	action := data["action"].(string)
	if action != ActionPublished && action != ActionEdited {
		impl.logger.Warnw("handling only published and edited action, ignored other actions", "action", action)
		return false, nil
	}
	releaseData := data["release"].(map[string]interface{})
	releaseName := releaseData["name"].(string)
	tagName := releaseData["tag_name"].(string)
	createdAtString := releaseData["created_at"].(string)
	createdAt, error := time.Parse(TimeFormatLayout, createdAtString)
	if error != nil {
		impl.logger.Errorw("error on time parsing, ignored this key", "err", error)
		//return false, nil
	}
	publishedAtString := releaseData["published_at"].(string)
	publishedAt, error := time.Parse(TimeFormatLayout, publishedAtString)
	if error != nil {
		impl.logger.Errorw("error on time parsing, ignored this key", "err", error)
		//return false, nil
	}
	body := releaseData["body"].(string)
	releaseInfo := &common.Release{
		TagName:     tagName,
		ReleaseName: releaseName,
		Body:        body,
		CreatedAt:   createdAt,
		PublishedAt: publishedAt,
		TagLink:     fmt.Sprintf("%s/%s", TagLink, tagName),
	}
	impl.getPrerequisiteContent(releaseInfo)

	//updating cache, fetch existing object and append new item
	var releaseList []*common.Release
	//releaseList = append(releaseList, releaseInfo)
	cachedReleases := impl.releaseCache.GetReleaseCache()
	if cachedReleases != nil {
		itemMap, ok := cachedReleases.(map[string]cache.Item)
		if !ok {
			// Can't assert, handle error.
			impl.logger.Error("Can't assert, handle err")
			return false, nil
		}
		impl.logger.Info(itemMap)
		if itemMap != nil {
			items := itemMap["releases"]
			if items.Object != nil {
				releases := items.Object.([]*common.Release)
				releaseList = append(releaseList, releases...)
			}
		}
	}

	isNew := true
	for _, release := range releaseList {
		if release.ReleaseName == releaseInfo.ReleaseName {
			release.Body = releaseInfo.Body
			isNew = false
		}
	}
	if isNew {
		releaseList = append([]*common.Release{releaseInfo}, releaseList...)
	}
	impl.mutex.Lock()
	defer impl.mutex.Unlock()
	impl.releaseCache.UpdateReleaseCache(releaseList)
	return true, nil
}

func (impl *ReleaseNoteServiceImpl) GetReleases() ([]*common.Release, error) {
	var releaseList []*common.Release
	cachedReleases := impl.releaseCache.GetReleaseCache()
	if cachedReleases != nil {
		itemMap, ok := cachedReleases.(map[string]cache.Item)
		if !ok {
			impl.logger.Error("Can't assert, handle err")
			return releaseList, nil
		}
		impl.logger.Info(itemMap)
		if itemMap != nil {
			items := itemMap["releases"]
			if items.Object != nil {
				releases := items.Object.([]*common.Release)
				releaseList = append(releaseList, releases...)
			}
		}
	}

	if releaseList == nil {
		operationComplete := false
		retryCount := 0
		for !operationComplete && retryCount < 3 {
			retryCount = retryCount + 1
			releases, _, err := impl.client.GitHubClient.Repositories.ListReleases(context.Background(), impl.client.GitHubConfig.GitHubOrg, impl.client.GitHubConfig.GitHubRepo, &github.ListOptions{})
			if err != nil {
				responseErr, ok := err.(*github.ErrorResponse)
				if !ok || responseErr.Response.StatusCode != 404 {
					impl.logger.Errorw("error in fetching releases from github", "err", err, "config", "config")
					//todo - any specific message
					continue
				} else {
					impl.logger.Errorw("error in fetching releases from github", "err", err)
					continue
				}
			}
			if err == nil {
				operationComplete = true
			}
			result := &common.ReleaseList{}
			var releasesDto []*common.Release
			for _, item := range releases {
				dto := &common.Release{
					TagName:     *item.TagName,
					ReleaseName: *item.Name,
					CreatedAt:   item.CreatedAt.Time,
					PublishedAt: item.PublishedAt.Time,
					Body:        *item.Body,
					TagLink:     fmt.Sprintf("%s/%s", TagLink, *item.TagName),
				}
				impl.getPrerequisiteContent(dto)
				releasesDto = append(releasesDto, dto)
			}
			result.Releases = releasesDto
			releaseList = releasesDto
			impl.mutex.Lock()
			defer impl.mutex.Unlock()
			impl.releaseCache.UpdateReleaseCache(releaseList)
		}
		if !operationComplete {
			return releaseList, fmt.Errorf("failed operation on fetching releases from github, attempted 3 times")
		}
	}
	return releaseList, nil
}

func (impl *ReleaseNoteServiceImpl) getPrerequisiteContent(releaseInfo *common.Release) {
	if strings.Contains(releaseInfo.Body, PrerequisitesMatcher) {
		releaseInfo.Prerequisite = true
		start := strings.Index(releaseInfo.Body, PrerequisitesMatcher)
		end := strings.LastIndex(releaseInfo.Body, PrerequisitesMatcher)
		if end == 0 {
			return
		}
		prerequisiteMessage := strings.ReplaceAll(releaseInfo.Body[start:end], PrerequisitesMatcher, "")
		releaseInfo.PrerequisiteMessage = prerequisiteMessage
	}
}

func (impl *ReleaseNoteServiceImpl) GetModules() ([]*common.Module, error) {
	var modules []*common.Module
	modules = append(modules, &common.Module{
		Id:                            1,
		Name:                          "cicd",
		BaseMinVersionSupported:       impl.moduleConfig.ModuleConfig.BaseMinVersionSupported,
		IsIncludedInLegacyFullPackage: true,
		Description:                   impl.moduleConfig.ModuleConfig.Description,
		Title:                         impl.moduleConfig.ModuleConfig.Title,
		Icon:                          impl.moduleConfig.ModuleConfig.Icon,
		Info:                          impl.moduleConfig.ModuleConfig.Info,
		Assets:                        impl.moduleConfig.ModuleConfig.Assets,
		DependentModules:              []int{},
	})
	return modules, nil
}

func (impl *ReleaseNoteServiceImpl) GetModulesV2() ([]*common.Module, error) {
	var modules []*common.Module
	modules = append(modules, &common.Module{
		Id:                            1,
		Name:                          "cicd",
		BaseMinVersionSupported:       impl.moduleConfig.ModuleConfig.BaseMinVersionSupported,
		IsIncludedInLegacyFullPackage: true,
		Description:                   impl.moduleConfig.ModuleConfig.Description,
		Title:                         impl.moduleConfig.ModuleConfig.Title,
		Icon:                          impl.moduleConfig.ModuleConfig.Icon,
		Info:                          impl.moduleConfig.ModuleConfig.Info,
		Assets:                        impl.moduleConfig.ModuleConfig.Assets,
		DependentModules:              []int{},
	})
	modules = append(modules, &common.Module{
		Id:                            2,
		Name:                          "argo-cd",
		BaseMinVersionSupported:       "v0.5.3",
		IsIncludedInLegacyFullPackage: true,
		Description:                   "<div class=\"module-details__feature-info fs-14 fw-4\"><p>GitOps is an operational framework that takes DevOps best practices used for application development such as version control, collaboration, compliance and applies them to infrastructure automation. Similar to how teams use application source code, operations teams that adopt GitOps use configuration files stored as code (infrastructure as code).</p><p>Devtron uses GitOps to automate the process of provisioning infrastructure. GitOps configuration files generate the same infrastructure environment every time it’s deployed, just as application source code generates the same application binaries every time it’s built.</p><h3 class=\"module-details__features-list-heading fs-14 fw-6\">Features:</h3><ul class=\"module-details__features-list pl-22 mb-24\"><li>Implements GitOps to manage the state of Kubernetes applications.</li><li>Simplified and abstracted integration with ArgoCD for GitOps operation.</li><li>No prior knowledge of ArgoCD is required.</li></ul></div>",
		Title:                         "GitOps (by Argo CD)",
		Icon:                          "https://cdn.devtron.ai/images/ic-integration-gitops-argocd.png",
		Info:                          "Declarative GitOps CD for Kubernetes powered by Argo CD",
		Assets:                        []string{"https://cdn.devtron.ai/images/img-gitops-1.png"},
		DependentModules:              []int{1},
	})

	modules = append(modules, &common.Module{
		Id:                            3,
		Name:                          "security-clair",
		BaseMinVersionSupported:       "v0.5.4",
		IsIncludedInLegacyFullPackage: true,
		Description:                   "<div class=\"module-details__feature-info fs-14 fw-4\"><p>GitOps is an operational framework that takes DevOps best practices used for application development such as version control, collaboration, compliance and applies them to infrastructure automation. Similar to how teams use application source code, operations teams that adopt GitOps use configuration files stored as code (infrastructure as code).</p><p>Devtron uses GitOps to automate the process of provisioning infrastructure. GitOps configuration files generate the same infrastructure environment every time it’s deployed, just as application source code generates the same application binaries every time it’s built.</p><h3 class=\"module-details__features-list-heading fs-14 fw-6\">Features:</h3><ul class=\"module-details__features-list pl-22 mb-24\"><li>Implements GitOps to manage the state of Kubernetes applications.</li><li>Simplified and abstracted integration with ArgoCD for GitOps operation.</li><li>No prior knowledge of ArgoCD is required.</li></ul></div>",
		Title:                         "Vulnerability scanning (Clair)",
		Icon:                          "https://cdn.devtron.ai/images/ic-integration-security-clair.png",
		Info:                          "Seamless integration with Clair for vulnerability scanning of images.",
		Assets:                        []string{"https://cdn.devtron.ai/images/img-gitops-1.png"},
		DependentModules:              []int{1},
	})

	return modules, nil
}

func (impl *ReleaseNoteServiceImpl) GetModuleByName(name string) (*common.Module, error) {
	module := &common.Module{}
	modules, err := impl.GetModulesV2()
	if err != nil {
		impl.logger.Errorw("error on fetching modules", "err", err)
		return module, err
	}
	for _, item := range modules {
		if item.Name == name {
			module = item
		}
	}
	return module, nil
}
