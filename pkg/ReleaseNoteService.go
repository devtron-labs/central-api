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
		// tag is mandatory while drafting a new release
		if release.TagName == releaseInfo.TagName {
			release.ReleaseName = releaseInfo.ReleaseName
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
				if item == nil {
					impl.logger.Warnw("error while getting release from repository", "err", err)
					continue
				}
				var tagName, releaseName, body, tagLink string
				var createdAt, publishedAt time.Time
				if item.TagName != nil {
					tagName = *item.TagName
				}
				if item.Name != nil {
					releaseName = *item.Name
				}
				if item.Body != nil {
					body = *item.Body
				}
				if item.TagName != nil {
					tagLink = fmt.Sprintf("%s/%s", TagLink, *item.TagName)
				}
				if item.CreatedAt != nil {
					createdAt = item.CreatedAt.Time
				}
				if item.PublishedAt != nil {
					publishedAt = item.PublishedAt.Time
				}
				dto := &common.Release{
					TagName:     tagName,
					ReleaseName: releaseName,
					CreatedAt:   createdAt,
					PublishedAt: publishedAt,
					Body:        body,
					TagLink:     tagLink,
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
		BaseMinVersionSupported:       "v0.6.0",
		IsIncludedInLegacyFullPackage: true,
		Description:                   "<div class=\"module-details__feature-info fs-14 fw-4\"><p>GitOps is an operational framework that takes DevOps best practices used for application development such as version control, collaboration, compliance and applies them to infrastructure automation. Similar to how teams use application source code, operations teams that adopt GitOps use configuration files stored as code (infrastructure as code).</p><p>Devtron uses GitOps to automate the process of provisioning infrastructure. GitOps configuration files generate the same infrastructure environment every time it’s deployed, just as application source code generates the same application binaries every time it’s built.</p><h3 class=\"module-details__features-list-heading fs-14 fw-6\">Features:</h3><ul class=\"module-details__features-list pl-22 mb-24\"><li>Implements GitOps to manage the state of Kubernetes applications.</li><li>Simplified and abstracted integration with ArgoCD for GitOps operation.</li><li>No prior knowledge of ArgoCD is required.</li></ul></div>",
		Title:                         "GitOps (Argo CD)",
		Icon:                          "https://cdn.devtron.ai/images/ic-integration-gitops-argocd.png",
		Info:                          "Declarative GitOps CD for Kubernetes powered by Argo CD",
		Assets:                        []string{"https://cdn.devtron.ai/images/img-gitops-1.png"},
		DependentModules:              []int{1},
		ResourceFilter: &common.ResourceFilter{
			GlobalFilter: &common.ResourceIdentifier{
				Labels: map[string]string{
					"app.kubernetes.io/part-of": "argocd",
				},
			},
		},
	})

	modules = append(modules, &common.Module{
		Id:                            3,
		Name:                          "security.clair",
		BaseMinVersionSupported:       "v0.6.0",
		IsIncludedInLegacyFullPackage: true,
		Description:                   "<div class=\"module-details__feature-info fs-14 fw-4\"><p>When you work with containers (Docker) you are not only packaging your application but also part of the OS. It is crucial to know what kind of libraries might be vulnerable in your container. One way to find this information is to look at the Docker registry [Hub or Quay.io] security scan. This means your vulnerable image is already on the Docker registry.</p><p>What you want is a scan as a part of CI/CD pipeline that stops the Docker image push on vulnerabilities:</p><ul class=\"module-details__features-list pl-22 mb-24\" style=\"\n    list-style: decimal;\n\"><li>Build and test your application\n</li><li>Build the container\n</li><li>Test the container for vulnerabilities\n</li><li>Check the vulnerabilities against allowed ones, if everything is allowed then pass otherwise fail\n</li></ul><p>This straightforward process is not that easy to achieve when using the services like Docker Hub or Quay.io. This is because they work asynchronously which makes it harder to do straightforward CI/CD pipeline.</p><h3 class=\"module-details__features-list-heading fs-14 fw-6\">Features:</h3><ul class=\"module-details__features-list pl-22 mb-24\"><li>Scans an image against Clair server</li><li>Compares the vulnerabilities against a whitelist</li><li>Blocks images from deployment if blacklisted / blocked vulnerabilities are detected</li><li>Ability to define hierarchical security policy (Global / Cluster / Environment / Application) to allow / block vulnerabilities based on criticality (High / Moderate / Low)</li><li>Shows security vulnerabilities detected in deployed applications</li></ul></div>",
		Title:                         "Vulnerability Scanning (Clair)",
		Icon:                          "https://cdn.devtron.ai/images/ic-integration-security-clair.png",
		Info:                          "Seamless integration with Clair for vulnerability scanning of images.",
		Assets:                        []string{"https://cdn.devtron.ai/images/img-security-clair-1.png", "https://cdn.devtron.ai/images/img-security-clair-2.png", "https://cdn.devtron.ai/images/img-security-clair-3.png", "https://cdn.devtron.ai/images/img-security-clair-4.png"},
		DependentModules:              []int{1},
	})

	modules = append(modules, &common.Module{
		Id:                            4,
		Name:                          "notifier",
		BaseMinVersionSupported:       "v0.6.0",
		IsIncludedInLegacyFullPackage: true,
		Description:                   "<div class=\"module-details__feature-info fs-14 fw-4\"><p>Receive alerts for build and deployment pipelines on trigger, success, and failure events. An alert will be sent to desired slack channel and Email(supports SES and SMTP configurations) with the required information to take be able to quick actions whenever required.</p><h3 class=\"module-details__features-list-heading fs-14 fw-6\">Features:</h3><ul class=\"module-details__features-list pl-22 mb-24\"><li>Receive alerts for start, success, and failure events on desired build pipelines</li><li>Receive alerts for start, success, and failure events on desired deployment pipelines</li><li>Receive alerts on desired Slack channels via webhook</li><li>Receive alerts on your email address (supports SES and SMTP)</li></ul><h3 class=\"module-details__features-list-heading fs-14 fw-6\">How to use the Integration?</h3><span>After you install the integration, you can configure notifications from Global configurations &gt; Notifications section. For more details on how to configure notifications, refer\n<a href=\"https://docs.devtron.ai/getting-started/global-configurations/manage-notification\" target=\"_blank\">here</a>.\n</span></div>",
		Title:                         "Notifications",
		Icon:                          "https://cdn.devtron.ai/images/ic-integration-notifications.png",
		Info:                          "Get notified when build and deployment pipelines start, fail or succeed.",
		Assets:                        []string{"https://cdn.devtron.ai/images/img-notification-1.png", "https://cdn.devtron.ai/images/img-notification-2.png", "https://cdn.devtron.ai/images/img-notification-3.png"},
		DependentModules:              []int{1},
	})

	modules = append(modules, &common.Module{
		Id:                            5,
		Name:                          "monitoring.grafana",
		BaseMinVersionSupported:       "v0.6.0",
		IsIncludedInLegacyFullPackage: true,
		Description:                   "<div class=\"module-details__feature-info fs-14 fw-4\"><p>Devtron leverages the power of Grafana to show application metrics like CPU, Memory utilization, Status 4xx/ 5xx/ 2xx, Throughput, and Latency.</p><h3 class=\"module-details__features-list-heading fs-14 fw-6\">Features:</h3><ul class=\"module-details__features-list pl-22 mb-24\"><li>CPU usage: Displays the overall utilization of CPU by an application. It is available as aggregated or per pod.</li><li>Memory usage: Displays the overall utilization of memory by an application. It is available as aggregated or per pod.</li><li>Throughput: Indicates the number of requests processed by an application per minute.</li><li>Status codes: Indicates the application’s response to the client’s request with a specific status code as shown below:<ul class=\"module-details__features-list pl-22 mb-24\"><li>1xx: Communicates transfer protocol level information</li><li>2xx: Client’s request is processed successfully</li><li>3xx: Client must take some additional action to complete their request</li><li>4xx: There is an error on the client side</li><li>5xx: There is an error on the server side</li></ul></li></ul><h3 class=\"module-details__features-list-heading fs-14 fw-6\">How to use the Integration?</h3><span>After you install the integration, you can enable application metrics for all or specific environments in an application. For more details on how to enable application metrics, refer \n<a href=\"https://docs.devtron.ai/v/v0.5/usage/applications/app-details/app-metrics\" target=\"_blank\">here</a>\n</span></div>",
		Title:                         "Monitoring (Grafana)",
		Icon:                          "https://cdn.devtron.ai/images/ic-integration-grafana.png",
		Info:                          "Enables metrics like CPU, memory, status codes, throughput, and latency for applications.",
		Assets:                        []string{"https://cdn.devtron.ai/images/img-grafana-1.png", "https://cdn.devtron.ai/images/img-grafana-2.png"},
		DependentModules:              []int{1},
		ResourceFilter: &common.ResourceFilter{
			GlobalFilter: &common.ResourceIdentifier{
				Labels: map[string]string{
					"app.kubernetes.io/name": "grafana",
				},
			},
		},
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
