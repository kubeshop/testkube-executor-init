package runner

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/kubeshop/testkube/pkg/api/v1/testkube"
	"github.com/kubeshop/testkube/pkg/envs"
	"github.com/kubeshop/testkube/pkg/executor"
	"github.com/kubeshop/testkube/pkg/executor/content"
	"github.com/kubeshop/testkube/pkg/executor/output"
	"github.com/kubeshop/testkube/pkg/executor/runner"
	"github.com/kubeshop/testkube/pkg/ui"
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
	output.PrintLog(fmt.Sprintf("%s Initializing...", ui.IconTruck))
	params, err := envs.LoadTestkubeVariables()
	if err != nil {
		output.PrintLog(fmt.Sprintf("%s Environment variables read unsuccessfully", ui.IconCross))
		return result, fmt.Errorf("%s Could not read environment variables: %w", ui.IconCross, err)
	}

	gitUsername := params.GitUsername
	gitToken := params.GitToken

	if gitUsername != "" && gitToken != "" {
		if execution.Content != nil && execution.Content.Repository != nil {
			execution.Content.Repository.Username = gitUsername
			execution.Content.Repository.Token = gitToken
		}
	}

	if execution.VariablesFile != "" {
		output.PrintLog(fmt.Sprintf("%s Creating variables file...", ui.IconWorld))
		file := filepath.Join(r.dir, "params-file")
		if err = os.WriteFile(file, []byte(execution.VariablesFile), 0666); err != nil {
			output.PrintLog(fmt.Sprintf("%s Could not create variables file %s: %s", ui.IconCross, file, err.Error()))
			return result, fmt.Errorf("could not create variables file %s: %w", file, err)
		}
		output.PrintLog(fmt.Sprintf("%s Variables file created", ui.IconCheckMark))
	}

	_, err = r.Fetcher.Fetch(execution.Content)
	if err != nil {
		output.PrintLog(fmt.Sprintf("%s Could not fetch test content: %s", ui.IconCross, err.Error()))
		return result, fmt.Errorf("could not fetch test content: %w", err)
	}

	// add copy files in case object storage is set
	if params.Endpoint != "" {
		output.PrintLog(fmt.Sprintf("%s Fetching uploads from object store %s...", ui.IconFile, params.Endpoint))
		fp := content.NewCopyFilesPlacer(params.Endpoint, params.AccessKeyID, params.SecretAccessKey, params.Location, params.Token, params.Bucket, params.Ssl)
		fp.PlaceFiles(execution.TestName, execution.BucketName)
	}

	output.PrintLog(fmt.Sprintf("%s Setting up access to files in %s", ui.IconFile, r.dir))
	_, err = executor.Run(r.dir, "chmod", nil, []string{"-R", "777", "."}...)
	if err != nil {
		output.PrintLog(fmt.Sprintf("%s Could not chmod for data dir: %s", ui.IconCross, err.Error()))
	}

	if execution.ArtifactRequest != nil {
		_, err = executor.Run(execution.ArtifactRequest.VolumeMountPath, "chmod", nil, []string{"-R", "777", "."}...)
		if err != nil {
			output.PrintLog(fmt.Sprintf("%s Could not chmod for artifacts dir: %s", ui.IconCross, err.Error()))
		}
	}
	output.PrintLog(fmt.Sprintf("%s Access to files enabled", ui.IconCheckMark))

	output.PrintLog(fmt.Sprintf("%s Initialization successful", ui.IconCheckMark))
	return testkube.NewPendingExecutionResult(), nil
}

// GetType returns runner type
func (r *InitRunner) GetType() runner.Type {
	return runner.TypeInit
}
