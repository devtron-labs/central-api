package pkg

import (
	"github.com/devtron-labs/central-api/common"
	"go.uber.org/zap"
)

type CiBuildMetadataService interface {
	GetDockerfileTemplateMetadata() *common.DockerfileTemplateMetadata
	GetBuildpackMetadata() *common.BuildPackMetadata
}

type CiBuildMetadataServiceImpl struct {
	Logger                     *zap.SugaredLogger
	BuildPackMetadata          *common.BuildPackMetadata
	DockerfileTemplateMetadata *common.DockerfileTemplateMetadata
}

func NewCiBuildMetadataServiceImpl(logger *zap.SugaredLogger) *CiBuildMetadataServiceImpl {
	buildpackMetadata := setupBuildpackMetadata()
	templateMetadata := setupDockerfileTemplateMetadata()
	metadataServiceImpl := &CiBuildMetadataServiceImpl{
		Logger:                     logger,
		BuildPackMetadata:          buildpackMetadata,
		DockerfileTemplateMetadata: templateMetadata,
	}
	return metadataServiceImpl
}

func setupDockerfileTemplateMetadata() *common.DockerfileTemplateMetadata {

	var languageFrameworks []*common.LanguageFramework
	languageFrameworks = append(languageFrameworks, &common.LanguageFramework{
		Language:     common.JAVA,
		Framework:    common.MAVEN,
		LanguageIcon: "https://cdn.devtron.ai/images/ic-Java.png",
		TemplateUrl:  "https://raw.githubusercontent.com/devtron-labs/devtron/main/sample-docker-templates/java/Maven_Dockerfile",
	})
	languageFrameworks = append(languageFrameworks, &common.LanguageFramework{
		Language:     common.JAVA,
		Framework:    common.GRADLE,
		LanguageIcon: "https://cdn.devtron.ai/images/ic-Java.png",
		TemplateUrl:  "https://raw.githubusercontent.com/devtron-labs/devtron/main/sample-docker-templates/java/Gradle_Dockerfile",
	})
	languageFrameworks = append(languageFrameworks, &common.LanguageFramework{
		Language:     common.GO,
		LanguageIcon: "https://cdn.devtron.ai/images/ic-go.png",
		TemplateUrl:  "https://raw.githubusercontent.com/devtron-labs/devtron/main/sample-docker-templates/go/Dockerfile",
	})
	languageFrameworks = append(languageFrameworks, &common.LanguageFramework{
		Language:     common.PYTHON,
		Framework:    common.DJANGO,
		LanguageIcon: "https://cdn.devtron.ai/images/ic-python.png",
		TemplateUrl:  "https://raw.githubusercontent.com/devtron-labs/devtron/main/sample-docker-templates/django/Dockerfile",
	})
	languageFrameworks = append(languageFrameworks, &common.LanguageFramework{
		Language:     common.PYTHON,
		Framework:    common.FLASK,
		LanguageIcon: "https://cdn.devtron.ai/images/ic-python.png",
		TemplateUrl:  "https://raw.githubusercontent.com/devtron-labs/devtron/main/sample-docker-templates/flask/Dockerfile",
	})
	languageFrameworks = append(languageFrameworks, &common.LanguageFramework{
		Language:     common.NODE,
		LanguageIcon: "https://cdn.devtron.ai/images/ic-nodejs.png",
		TemplateUrl:  "https://raw.githubusercontent.com/devtron-labs/devtron/main/sample-docker-templates/node/Dockerfile",
	})
	return &common.DockerfileTemplateMetadata{
		LanguageFrameworks: languageFrameworks,
	}
}

func setupBuildpackMetadata() *common.BuildPackMetadata {
	buildpackMetadata := &common.BuildPackMetadata{
		LanguageBuilder: CreateLanguageBuilderMetadata(),
	}
	return buildpackMetadata
}

func CreateLanguageBuilderMetadata() []*common.LanguageBuilder {
	var languageBuilders []*common.LanguageBuilder
	languageBuilders = append(languageBuilders, &common.LanguageBuilder{
		Language:     common.JAVA,
		LanguageIcon: "https://cdn.devtron.ai/images/ic-Java.png",
		Versions:     []string{"8", "11"},
		BuilderLanguageMetadata: []*common.BuilderLanguageMetadata{
			{Id: "gcr.io/buildpacks/builder:v1", BuilderLangEnvParam: "GOOGLE_RUNTIME_VERSION"},
			{Id: "paketobuildpacks/builder:full", BuilderLangEnvParam: "BP_JVM_VERSION"},
			{Id: "paketobuildpacks/builder:base", BuilderLangEnvParam: "BP_JVM_VERSION"},
			{Id: "paketobuildpacks/builder:tiny", BuilderLangEnvParam: "BP_JVM_VERSION"},
			{Id: "heroku/buildpacks:20", BuilderLangEnvParam: "DEVTRON_LANG_VERSION"},
		},
	})
	languageBuilders = append(languageBuilders, &common.LanguageBuilder{
		Language:     common.PYTHON,
		LanguageIcon: "https://cdn.devtron.ai/images/ic-python.png",
		Versions:     []string{"3.7.*"},
		BuilderLanguageMetadata: []*common.BuilderLanguageMetadata{
			{Id: "gcr.io/buildpacks/builder:v1", BuilderLangEnvParam: "GOOGLE_RUNTIME_VERSION"},
			{Id: "paketobuildpacks/builder:full", BuilderLangEnvParam: "BP_CPYTHON_VERSION"},
			{Id: "paketobuildpacks/builder:base", BuilderLangEnvParam: "BP_CPYTHON_VERSION"},
			{Id: "heroku/buildpacks:20", BuilderLangEnvParam: "DEVTRON_LANG_VERSION"},
		},
	})
	languageBuilders = append(languageBuilders, &common.LanguageBuilder{
		Language:     common.PHP,
		LanguageIcon: "https://cdn.devtron.ai/images/ic-php.png",
		Versions:     []string{"7.4"},
		BuilderLanguageMetadata: []*common.BuilderLanguageMetadata{
			{Id: "gcr.io/buildpacks/builder:v1", BuilderLangEnvParam: "GOOGLE_RUNTIME_VERSION"},
			{Id: "paketobuildpacks/builder:full", BuilderLangEnvParam: "BP_PHP_VERSION"},
			{Id: "paketobuildpacks/builder:base", BuilderLangEnvParam: "BP_PHP_VERSION"},
			{Id: "heroku/buildpacks:20", BuilderLangEnvParam: ""},
		},
	})
	languageBuilders = append(languageBuilders, &common.LanguageBuilder{
		Language:     common.GO,
		LanguageIcon: "https://cdn.devtron.ai/images/ic-go.png",
		Versions:     []string{"1.18", "1.19"},
		BuilderLanguageMetadata: []*common.BuilderLanguageMetadata{
			{Id: "gcr.io/buildpacks/builder:v1", BuilderLangEnvParam: "GOOGLE_RUNTIME_VERSION"},
			{Id: "paketobuildpacks/builder:full", BuilderLangEnvParam: "BP_GO_VERSION"},
			{Id: "paketobuildpacks/builder:base", BuilderLangEnvParam: "BP_GO_VERSION"},
			{Id: "paketobuildpacks/builder:tiny", BuilderLangEnvParam: "BP_GO_VERSION"},
			{Id: "heroku/buildpacks:20", BuilderLangEnvParam: "GOVERSION"},
		},
	})
	languageBuilders = append(languageBuilders, &common.LanguageBuilder{
		Language:     common.RUBY,
		LanguageIcon: "https://cdn.devtron.ai/images/ic-ruby.png",
		Versions:     []string{"2.7"},
		BuilderLanguageMetadata: []*common.BuilderLanguageMetadata{
			{Id: "gcr.io/buildpacks/builder:v1", BuilderLangEnvParam: "GOOGLE_RUNTIME_VERSION"},
			{Id: "paketobuildpacks/builder:full", BuilderLangEnvParam: "BP_MRI_VERSION"},
			{Id: "paketobuildpacks/builder:base", BuilderLangEnvParam: "BP_MRI_VERSION"},
			{Id: "heroku/buildpacks:20", BuilderLangEnvParam: ""},
		},
	})
	//languageBuilders = append(languageBuilders, &common.LanguageBuilder{
	//	Language: common.DOTNET,
	//	LanguageIcon: "https://cdn.devtron.ai/images/ic-Dotnet.png",
	//	Versions: []string{"5.0.4"},
	//	BuilderLanguageMetadata: []*common.BuilderLanguageMetadata{
	//		{Id: "gcr.io/buildpacks/builder:v1", BuilderLangEnvParam: "GOOGLE_RUNTIME_VERSION"},
	//		{Id: "paketobuildpacks/builder:full", BuilderLangEnvParam: "BP_DOTNET_FRAMEWORK_VERSION"},
	//		{Id: "paketobuildpacks/builder:base", BuilderLangEnvParam: "BP_DOTNET_FRAMEWORK_VERSION"},
	//		//{Id: "heroku/buildpacks:18", BuilderLangEnvParam: ""},
	//		{Id: "heroku/buildpacks:20", BuilderLangEnvParam: ""},
	//	},
	//})
	languageBuilders = append(languageBuilders, &common.LanguageBuilder{
		Language:     common.NODE,
		LanguageIcon: "https://cdn.devtron.ai/images/ic-nodejs.png",
		Versions:     []string{"16.x", "18.x"},
		BuilderLanguageMetadata: []*common.BuilderLanguageMetadata{
			{Id: "gcr.io/buildpacks/builder:v1", BuilderLangEnvParam: "GOOGLE_RUNTIME_VERSION"},
			{Id: "paketobuildpacks/builder:full", BuilderLangEnvParam: "BP_NODE_VERSION"},
			{Id: "paketobuildpacks/builder:base", BuilderLangEnvParam: "BP_NODE_VERSION"},
			{Id: "heroku/buildpacks:20", BuilderLangEnvParam: "DEVTRON_LANG_VERSION"},
		},
	})

	return languageBuilders
}

func (impl CiBuildMetadataServiceImpl) GetDockerfileTemplateMetadata() *common.DockerfileTemplateMetadata {
	return impl.DockerfileTemplateMetadata
}

func (impl CiBuildMetadataServiceImpl) GetBuildpackMetadata() *common.BuildPackMetadata {
	return impl.BuildPackMetadata
}
