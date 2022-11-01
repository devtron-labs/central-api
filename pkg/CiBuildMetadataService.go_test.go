package pkg

import (
	"encoding/json"
	"fmt"
	"github.com/devtron-labs/central-api/internal/logger"
	"os"
	"testing"
)

func TestCiBuildMetadataService(t *testing.T) {
	t.Run("buildpackMetadata", func(t *testing.T) {
		metadataServiceImpl := NewCiBuildMetadataServiceImpl(logger.NewSugardLogger())
		jsonBytes, _ := json.Marshal(metadataServiceImpl.BuildPackMetadata)
		file, err := os.Create("BuildpackMetadata.json")
		fmt.Println(err)
		_, err = file.WriteString(string(jsonBytes))
		fmt.Println(err)
	})
}
