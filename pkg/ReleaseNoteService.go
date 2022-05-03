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
	"sync"
	"time"
)

type ReleaseNoteService interface {
	GetModules() ([]*common.Module, error)
	GetReleases() ([]*common.Release, error)
	UpdateReleases(requestBodyBytes []byte) (bool, error)
}

type ReleaseNoteServiceImpl struct {
	logger       *zap.SugaredLogger
	client       *util.GitHubClient
	releaseCache *util.ReleaseCache
	mutex        sync.Mutex
}

func NewReleaseNoteServiceImpl(logger *zap.SugaredLogger, client *util.GitHubClient, releaseCache *util.ReleaseCache) *ReleaseNoteServiceImpl {
	serviceImpl := &ReleaseNoteServiceImpl{
		logger:       logger,
		client:       client,
		releaseCache: releaseCache,
	}
	_, err := serviceImpl.GetReleases()
	if err != nil {
		serviceImpl.logger.Errorw("error on app init call for releases", "err", err)
		//ignore error for starting application
	}
	return serviceImpl
}

const ActionPublished = "published"
const EventTypeRelease = "release"

func (impl *ReleaseNoteServiceImpl) UpdateReleases(requestBodyBytes []byte) (bool, error) {
	data := make(map[string]interface{})
	err := json.Unmarshal(requestBodyBytes, &data)
	if err != nil {
		impl.logger.Errorw("unmarshal error", "err", err)
		return false, err
	}
	action := data["action"].(string)
	if action != ActionPublished {
		return false, nil
	}
	releaseData := data["release"].(map[string]interface{})
	releaseName := releaseData["name"].(string)
	tagName := releaseData["tag_name"].(string)
	createdAtString := releaseData["created_at"].(string)
	createdAt, error := time.Parse("2006-01-02T15:04:05.000Z", createdAtString)
	if error != nil {
		impl.logger.Error(error)
		//return false, nil
	}
	body := releaseData["body"].(string)
	releaseInfo := &common.Release{
		TagName:     tagName,
		ReleaseName: releaseName,
		Body:        body,
		CreatedAt:   createdAt,
	}

	//updating cache, fetch existing object and append new item
	var releaseList []*common.Release
	releaseList = append(releaseList, releaseInfo)
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
				}
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

func (impl *ReleaseNoteServiceImpl) GetModules() ([]*common.Module, error) {
	var modules []*common.Module
	modules = append(modules, &common.Module{
		Id:                            1,
		Name:                          common.MODULE_CICD,
		BaseMinVersionSupported:       "v0.0.1",
		IsIncludedInLegacyFullPackage: true,
		Assets:                        []string{"", ""},
		Description:                   "<div class=\"module-details__feature-info fs-13 fw-4\"><p>Continuous integration (CI) and continuous delivery (CD) embody a culture, set of operating principles, and collection of practices that enable application development teams to deliver code changes more frequently and reliably. The implementation is also known as the CI/CD pipeline.</p><p>CI/CD is one of the best practices for devops teams to implement. It is also an agile methodology best practice, as it enables software development teams to focus on meeting business requirements, code quality, and security because deployment steps are automated.</p><h3 class=\"module-details__features-list-heading fs-13 fw-6\">Features</h3><ul class=\"module-details__features-list pl-22 mb-24\"><li>Discovery: What would the users be searching for when they're looking for a CI/CD offering?</li><li>Detail: The CI/CD offering should be given sufficient importance (on Website, Readme). (Eg. Expand capability with CI/CD module [Discover more modules])</li><li>Installation: Ability to install CI/CD module with the basic installation.</li><li>In-Product discovery: How easy it is to discover the CI/CD offering primarily once the user is in the product. (Should we talk about modules on the login page?)</li></ul></div>",
		Title:                         "Build and Deploy (CI/CD)",
	})
	return modules, nil
}
