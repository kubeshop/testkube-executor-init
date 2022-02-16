package runner

import (
	"os"

	"github.com/kubeshop/testkube/pkg/api/v1/testkube"
	"github.com/kubeshop/testkube/pkg/executor/content"
)

// NewRunner creates init runner
func NewRunner() *InitRunner {
	return &InitRunner{
		Fetcher: content.NewFetcher(os.Getenv("RUNNER_DATADIR")),
	}
}

// InitRunner prepares data for executor
type InitRunner struct {
	Fetcher content.ContentFetcher
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

	path, err := r.Fetcher.Fetch(execution.Content)
	if err != nil {
		return result, err
	}

	return testkube.ExecutionResult{
		Status: testkube.StatusPtr(testkube.SUCCESS_ExecutionStatus),
		Output: "created content path: " + path,
	}, nil
}
