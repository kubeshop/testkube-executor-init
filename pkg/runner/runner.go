package runner

import (
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/kubeshop/testkube/pkg/api/v1/testkube"
	"github.com/kubeshop/testkube/pkg/executor/content"
)

// NewRunner creates init runner
func NewRunner() *InitRunner {
	dir := os.Getenv("RUNNER_DATADIR")
	return &InitRunner{
		Fetcher: content.NewFetcher(dir),
		dir:     dir,
	}
}

// InitRunner prepares data for executor
type InitRunner struct {
	Fetcher content.ContentFetcher
	dir     string
}

// Run prepares data for executor
func (r *InitRunner) Run(execution testkube.Execution) (result testkube.ExecutionResult, err error) {
	gitUsername := os.Getenv("RUNNER_GITUSERNAME")
	gitToken := os.Getenv("RUNNER_GITTOKEN")
	if gitUsername != "" && gitToken != "" {
		if execution.Content != nil && execution.Content.Repository != nil {
			execution.Content.Repository.Username = gitUsername
			execution.Content.Repository.Token = gitToken
		}
	}

	if execution.ParamsFile != "" {
		filename := "params-file"
		if err = ioutil.WriteFile(filepath.Join(r.dir, filename), []byte(execution.ParamsFile), 0666); err != nil {
			return result, err
		}
	}

	path, err := r.Fetcher.Fetch(execution.Content)
	if err != nil {
		return result, err
	}

	return testkube.ExecutionResult{
		Status: testkube.StatusPtr(testkube.SUCCESS_ExecutionStatus),
		Output: "created content path: " + path,
	}, nil
}
