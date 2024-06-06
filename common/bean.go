/*
 * Copyright (c) 2024. Devtron Inc.
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
	TagName             string    `json:"tagName"`
	ReleaseName         string    `json:"releaseName"`
	CreatedAt           time.Time `json:"createdAt"`
	PublishedAt         time.Time `json:"publishedAt"`
	Body                string    `json:"body"`
	Prerequisite        bool      `json:"prerequisite"`
	PrerequisiteMessage string    `json:"prerequisiteMessage"`
	TagLink             string    `json:"tagLink"`
}

const MODULE_CICD = "cicd"
const MODULE_Security = "security"

type Module struct {
	Id                            int             `json:"id"`
	Name                          string          `json:"name"`
	BaseMinVersionSupported       string          `json:"baseMinVersionSupported"`
	IsIncludedInLegacyFullPackage bool            `json:"isIncludedInLegacyFullPackage"`
	Assets                        []string        `json:"assets"`
	Description                   string          `json:"description"`
	Title                         string          `json:"title"`
	Icon                          string          `json:"icon"`
	Info                          string          `json:"info"`
	DependentModules              []int           `json:"dependentModules"`
	ResourceFilter                *ResourceFilter `json:"resourceFilter,omitempty"`
	ModuleType                    string          `json:"moduleType"`
}

type ResourceFilter struct {
	GlobalFilter    *ResourceIdentifier `json:"globalFilter,omitempty"`
	GvkLevelFilters []*GvkLevelFilter   `json:"gvkLevelFilters,omitempty"`
}

type GvkLevelFilter struct {
	Gvk                *GroupVersionKind   `json:"gvk"`
	ResourceIdentifier *ResourceIdentifier `json:"filter"`
}

type GroupVersionKind struct {
	Group   string `json:"group"`
	Version string `json:"version"`
	Kind    string `json:"kind"`
}

type ResourceIdentifier struct {
	Labels map[string]string `json:"labels"`
}
