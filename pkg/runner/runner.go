package runner

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/kelseyhightower/envconfig"
	"github.com/kubeshop/testkube/pkg/api/v1/testkube"
	"github.com/kubeshop/testkube/pkg/executor"
	"github.com/kubeshop/testkube/pkg/executor/content"
	"github.com/kubeshop/testkube/pkg/executor/output"
)

type Params struct {
	Endpoint        string // RUNNER_ENDPOINT
	AccessKeyID     string // RUNNER_ACCESSKEYID
	SecretAccessKey string // RUNNER_SECRETACCESSKEY
	Location        string // RUNNER_LOCATION
	Token           string // RUNNER_TOKEN
	Ssl             bool   // RUNNER_SSL
}

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
	var params Params
	err = envconfig.Process("runner", &params)
	if err != nil {
		return result, fmt.Errorf("could not read environment variables: %w", err)
	}

	gitUsername := os.Getenv("RUNNER_GITUSERNAME")
	gitToken := os.Getenv("RUNNER_GITTOKEN")
	if gitUsername != "" && gitToken != "" {
		if execution.Content != nil && execution.Content.Repository != nil {
			execution.Content.Repository.Username = gitUsername
			execution.Content.Repository.Token = gitToken
		}
	}

	if execution.VariablesFile != "" {
		filename := "params-file"
		if err = os.WriteFile(filepath.Join(r.dir, filename), []byte(execution.VariablesFile), 0666); err != nil {
			return result, err
		}
	}

	path, err := r.Fetcher.Fetch(execution.Content)
	if err != nil {
		return result, err
	}

	// add copy files in case object storage is set
	if params.Endpoint != "" {
		fp := content.NewCopyFilesPlacer(params.Endpoint, params.AccessKeyID, params.SecretAccessKey, params.Location, params.Token, params.Ssl)
		err = fp.PlaceFiles(execution.TestName, execution.BucketName)
		if err != nil {
			return result, err
		}
	}

	_, err = executor.Run(r.dir, "chmod", nil, []string{"-R", "777", "."}...)
	if err != nil {
		return result, err
	}

	output.PrintLog("created content path: " + path)

	return testkube.NewPendingExecutionResult(), nil
}
