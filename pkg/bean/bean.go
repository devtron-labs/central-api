package bean

import "fmt"

type Repository string

const (
	Oss Repository = "devtron"
)

func (i Repository) String() string {
	return string(i)
}

const ActionPublished = "published"
const ActionEdited = "edited"
const EventTypeRelease = "release"
const TimeFormatLayout = "2006-01-02T15:04:05Z"
const PrerequisitesMatcher = "<!--upgrade-prerequisites-required-->"

const (
	CACHE_KEY    = "latest"
	TempLocation = "/tmp/"
)

func GetCacheKeyBasedOnRepo(repo Repository) string {
	cacheKey := fmt.Sprintf("%s-%s", CACHE_KEY, repo)
	return cacheKey
}
