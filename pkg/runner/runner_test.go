package runner

import (
	"os"
	"testing"

	"github.com/kubeshop/testkube/pkg/api/v1/testkube"
	"github.com/stretchr/testify/assert"
)

func TestRun(t *testing.T) {

	t.Run("runner should run test based on execution data", func(t *testing.T) {
		os.Setenv("RUNNER_DATADIR", "./testdir")

		// given
		runner := NewRunner()
		execution := testkube.NewQueuedExecution()
		execution.Content = testkube.NewStringTestContent("hello I'm  test content")

		// when
		result, err := runner.Run(*execution)

		// then
		assert.NoError(t, err)
		assert.Equal(t, result.Status, testkube.ExecutionStatusRunning)
	})

	t.Run("runner should place files when copyFiles is set", func(t *testing.T) {
		// given
		runner := NewRunner()
		execution := testkube.NewQueuedExecution()
		execution.Content = testkube.NewStringTestContent("hello I'm  test content")
		filePath := "/tmp/file1"
		fileContent := "file-content1"
		execution.CopyFiles = map[string]string{
			filePath: fileContent,
		}

		// when
		result, err := runner.Run(*execution)

		// then
		assert.NoError(t, err)
		assert.Equal(t, result.Status, testkube.ExecutionStatusRunning)

		gotContent, err := os.ReadFile(filePath)
		assert.NoError(t, err)
		assert.Equal(t, fileContent, string(gotContent))

		err = os.Remove(filePath)
		// if there's an error, the file needs to be removed manually
		assert.NoError(t, err)
	})

}
