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
