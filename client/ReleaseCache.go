package util

import (
	"github.com/patrickmn/go-cache"
	"go.uber.org/zap"
	"time"
)

type ReleaseCache struct {
	cache  *cache.Cache
	logger *zap.SugaredLogger
}

func NewReleaseCache(logger *zap.SugaredLogger) *ReleaseCache {
	tokenCache := &ReleaseCache{
		cache:  cache.New(cache.NoExpiration, 5*time.Minute),
		logger: logger,
	}
	return tokenCache
}

func (impl *ReleaseCache) GetReleaseCache() interface{} {
	_, found := impl.cache.Get("releases")
	if !found {
		impl.cache.Add("releases", nil, cache.NoExpiration)
		return impl.cache.Items()
	}
	return impl.cache.Items()
}

func (impl *ReleaseCache) UpdateReleaseCache(value interface{}) {
	impl.cache.Set("releases", value, cache.NoExpiration)
}
