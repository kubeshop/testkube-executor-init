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
	"github.com/kubeshop/testkube/pkg/executor/runner"
	"github.com/kubeshop/testkube/pkg/ui"
)

type Params struct {
	Endpoint        string // RUNNER_ENDPOINT
	AccessKeyID     string // RUNNER_ACCESSKEYID
	SecretAccessKey string // RUNNER_SECRETACCESSKEY
	Location        string // RUNNER_LOCATION
	Token           string // RUNNER_TOKEN
	Ssl             bool   // RUNNER_SSL
	GitUsername     string // RUNNER_GITUSERNAME
	GitToken        string // RUNNER_GITTOKEN
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
	output.PrintEvent(fmt.Sprintf("%s Initializing...", ui.IconTruck))
	var params Params

	output.PrintEvent(fmt.Sprintf("%s Reading environment variables for runner setup...", ui.IconWorld))
	err = envconfig.Process("runner", &params)
	if err != nil {
		return result, fmt.Errorf("%s Could not read environment variables: %w", ui.IconCross, err)
	}

	gitUsername := params.GitUsername
	gitToken := params.GitToken
	output.PrintEvent(fmt.Sprintf("%s Environment variables read successfully", ui.IconCheckMark))
	printParams(params)

	if gitUsername != "" && gitToken != "" {
		if execution.Content != nil && execution.Content.Repository != nil {
			execution.Content.Repository.Username = gitUsername
			execution.Content.Repository.Token = gitToken
		}
	}

	if execution.VariablesFile != "" {
		output.PrintEvent(fmt.Sprintf("%s Creating variables file...", ui.IconWorld))
		file := filepath.Join(r.dir, "params-file")
		if err = os.WriteFile(file, []byte(execution.VariablesFile), 0666); err != nil {
			return result, fmt.Errorf("%s Could not create variables file %s: %w", ui.IconCross, file, err)
		}
		output.PrintEvent(fmt.Sprintf("%s Variables file created", ui.IconCheckMark))
	}

	output.PrintEvent(fmt.Sprintf("%s Fetching test content from %s...", ui.IconBox, execution.Content.Type_))
	path, err := r.Fetcher.Fetch(execution.Content)
	if err != nil {
		return result, fmt.Errorf("could not fetch test content: %w", err)
	}
	output.PrintEvent(fmt.Sprintf("%s Test content fetched to path %s", ui.IconCheckMark, path))

	// add copy files in case object storage is set
	if params.Endpoint != "" {
		output.PrintEvent(fmt.Sprintf("%s Fetching uploads from object store %s...", ui.IconFile, params.Endpoint))
		fp := content.NewCopyFilesPlacer(params.Endpoint, params.AccessKeyID, params.SecretAccessKey, params.Location, params.Token, params.Ssl)
		err = fp.PlaceFiles(execution.TestName, execution.BucketName)
		if err != nil {
			output.PrintLog(fmt.Sprintf("could not place bucket: %s", err.Error()))
		}
		output.PrintEvent(fmt.Sprintf("%s Placing uploads succeeded.", ui.IconCheckMark))
	}

	output.PrintEvent(fmt.Sprintf("%s Setting up access to files in %s", ui.IconFile, r.dir))
	_, err = executor.Run(r.dir, "chmod", nil, []string{"-R", "777", "."}...)
	if err != nil {
		output.PrintLog(fmt.Sprintf("could not chmod for data dir: %s", err.Error()))
	}

	if execution.ArtifactRequest != nil {
		_, err = executor.Run(execution.ArtifactRequest.VolumeMountPath, "chmod", nil, []string{"-R", "777", "."}...)
		if err != nil {
			output.PrintLog(fmt.Sprintf("could not chmod for artifacts dir: %s", err.Error()))
		}
	}
	output.PrintEvent(fmt.Sprintf("%s Access to files enabled", ui.IconCheckMark))

	output.PrintEvent(fmt.Sprintf("%s Initialization successful", ui.IconCheckMark))
	return testkube.NewPendingExecutionResult(), nil
}

// GetType returns runner type
func (r *InitRunner) GetType() runner.Type {
	return runner.TypeInit
}

// printParams shows the read parameters in logs
func printParams(params Params) {
	output.PrintLog(fmt.Sprintf("RUNNER_ENDPOINT=\"%s\"", params.Endpoint))
	printSensitiveParam("RUNNER_ACCESSKEYID", params.AccessKeyID)
	printSensitiveParam("RUNNER_SECRETACCESSKEY", params.SecretAccessKey)
	output.PrintLog(fmt.Sprintf("RUNNER_LOCATION=\"%s\"", params.Location))
	printSensitiveParam("RUNNER_TOKEN", params.Token)
	output.PrintLog(fmt.Sprintf("RUNNER_SSL=%t", params.Ssl))
	output.PrintLog(fmt.Sprintf("RUNNER_GITUSERNAME=\"%s\"", params.GitUsername))
	printSensitiveParam("RUNNER_GITTOKEN", params.GitToken)
}

// printSensitiveParam shows in logs if a parameter is set or not
func printSensitiveParam(name string, value string) {
	if len(value) == 0 {
		output.PrintLog(fmt.Sprintf("%s=\"\"", name))
	} else {
		output.PrintLog(fmt.Sprintf("%s=\"********\"", name))
	}
}
