package common

import "time"

type Response struct {
	Code   int         `json:"code,omitempty"`
	Status string      `json:"status,omitempty"`
	Result interface{} `json:"result,omitempty"`
	Errors []*ApiError `json:"errors,omitempty"`
}
type ApiError struct {
	HttpStatusCode    int         `json:"-"`
	Code              string      `json:"code,omitempty"`
	InternalMessage   string      `json:"internalMessage,omitempty"`
	UserMessage       interface{} `json:"userMessage,omitempty"`
	UserDetailMessage string      `json:"userDetailMessage,omitempty"`
}

type ReleaseList struct {
	Releases []*Release `json:"releases"`
}

type Release struct {
	TagName      string    `json:"tagName"`
	ReleaseName  string    `json:"releaseName"`
	CreatedAt    time.Time `json:"createdAt"`
	PublishedAt  time.Time `json:"publishedAt"`
	Body         string    `json:"body"`
	Prerequisite bool      `json:"prerequisite"`
	TagLink      string    `json:"tagLink"`
}

const MODULE_CICD = "cicd"
const MODULE_Security = "security"

type Module struct {
	Id                            int      `json:"id"`
	Name                          string   `json:"name"`
	BaseMinVersionSupported       string   `json:"baseMinVersionSupported"`
	IsIncludedInLegacyFullPackage bool     `json:"isIncludedInLegacyFullPackage"`
	Assets                        []string `json:"assets"`
	Description                   string   `json:"description"`
	Title                         string   `json:"title"`
	Icon                          string   `json:"icon"`
	Info                          string   `json:"info"`
}
