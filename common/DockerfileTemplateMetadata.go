package common

type DockerfileTemplateMetadata struct {
	LanguageFrameworks []*LanguageFramework
}

type LanguageFramework struct {
	Language     Language
	Framework    Framework
	LanguageIcon string
	TemplateUrl  string
}

type Framework string

const (
	MAVEN  Framework = "Maven"
	GRADLE Framework = "Gradle"
	DJANGO Framework = "Django"
	FLASK  Framework = "Flask"
	RAILS  Framework = "Rails"
)
