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
	}

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
		Name:                          impl.moduleConfig.ModuleConfig.Name,
		BaseMinVersionSupported:       impl.moduleConfig.ModuleConfig.BaseMinVersionSupported,
		IsIncludedInLegacyFullPackage: true,
		Description:                   impl.moduleConfig.ModuleConfig.Description,
		Title:                         impl.moduleConfig.ModuleConfig.Title,
		Icon:                          impl.moduleConfig.ModuleConfig.Icon,
		Info:                          impl.moduleConfig.ModuleConfig.Info,
		Assets:                        impl.moduleConfig.ModuleConfig.Assets,
	})
	return modules, nil
}
