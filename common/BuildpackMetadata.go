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

type Language string

const (
	NODE   Language = "Node"
	JAVA   Language = "Java"
	PYTHON Language = "Python"
	PHP    Language = "PHP"
	RUBY   Language = "Ruby"
	GO     Language = "Go"
	DOTNET Language = ".NET"
)

type BuildPackMetadata struct {
	LanguageBuilder []*LanguageBuilder
}

type BuilderLanguageMetadata struct {
	Id                  string
	BuilderLangEnvParam string
}

type LanguageBuilder struct {
	Language                Language
	LanguageIcon            string
	Versions                []string
	BuilderLanguageMetadata []*BuilderLanguageMetadata
}
