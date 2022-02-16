package runner

import (
	"github.com/kubeshop/testkube/pkg/api/v1/testkube"
	"github.com/kubeshop/testkube/pkg/executor/content"
)

// NewRunner creates init runner
func NewRunner() *InitRunner {
	return &InitRunner{
		Fetcher: content.NewFetcher(),
	}
}

// InitRunner prepares data for executor
type InitRunner struct {
	Fetcher content.ContentFetcher
}

// Run prepares data for executor
func (r *InitRunner) Run(execution testkube.Execution) (result testkube.ExecutionResult, err error) {
	path, err := r.Fetcher.Fetch(execution.Content)
	if err != nil {
		return result, err
	}

	return testkube.ExecutionResult{
		Status: testkube.StatusPtr(testkube.SUCCESS_ExecutionStatus),
		Output: "created content path: " + path,
	}, nil
}
