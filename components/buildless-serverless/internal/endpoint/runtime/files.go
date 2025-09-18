package runtime

import (
	"encoding/base64"
	"fmt"
	"os"

	"github.com/kyma-project/serverless/components/buildless-serverless/api/v1alpha2"
	"github.com/kyma-project/serverless/components/buildless-serverless/internal/endpoint/packagejson"
	"github.com/kyma-project/serverless/components/buildless-serverless/internal/endpoint/types"
	"github.com/pkg/errors"
)

func ReadFiles(f *v1alpha2.Function) ([]types.FileResponse, error) {
	runtimeDir := fmt.Sprintf("runtimes/%s", f.Spec.Runtime)

	if f.HasPythonRuntime() {
		// TODO: support non-nodejs runtimes
		return nil, errors.New("ejecting functions with non-nodejs runtimes is not supported")
	}

	return readNodejsFiles(f, runtimeDir)
}

func readNodejsFiles(f *v1alpha2.Function, runtimeDir string) ([]types.FileResponse, error) {
	commonFiles, err := readCommonFiles(runtimeDir)
	if err != nil {
		return nil, err
	}

	// read package.json and merge function dependencies
	packagejsonFile, err := os.ReadFile(runtimeDir + "/package.json")
	if err != nil {
		return nil, errors.Wrap(err, "failed to read package.json")
	}

	packagejsonFile, err = packagejson.Merge([]byte(f.Spec.Source.Inline.Dependencies), packagejsonFile)
	if err != nil {
		return nil, errors.Wrap(err, "failed to merge package.json")
	}

	// read server.mjs
	serverFile, err := os.ReadFile(runtimeDir + "/server.mjs")
	if err != nil {
		return nil, errors.Wrap(err, "failed to read server.mjs")
	}

	return append(commonFiles, []types.FileResponse{
		{Name: "package.json", Data: base64.StdEncoding.EncodeToString(packagejsonFile)},
		{Name: "server.mjs", Data: base64.StdEncoding.EncodeToString(serverFile)},
		{Name: "handler.js", Data: base64.StdEncoding.EncodeToString([]byte(f.Spec.Source.Inline.Source))},
	}...), nil
}

func readCommonFiles(runtimeDir string) ([]types.FileResponse, error) {
	// read lib files
	libFilesInfo, dirErr := os.ReadDir(runtimeDir + "/lib")
	if dirErr != nil {
		return nil, errors.Wrap(dirErr, "failed to read lib directory")
	}

	libFiles := make([]types.FileResponse, 0, len(libFilesInfo))
	for _, f := range libFilesInfo {
		if f.IsDir() {
			continue
		}

		data, err := os.ReadFile(runtimeDir + "/lib/" + f.Name())
		if err != nil {
			return nil, errors.Wrapf(err, "failed to read lib file '%s'", f.Name())
		}
		libFiles = append(libFiles, types.FileResponse{Name: fmt.Sprintf("/lib/%s", f.Name()), Data: base64.StdEncoding.EncodeToString(data)})
	}

	// read Makefile
	makefileFile, err := os.ReadFile(runtimeDir + "/cli/Makefile")
	if err != nil {
		return nil, errors.Wrap(err, "failed to read Makefile")
	}

	// read Dockerfile
	dockerfileFile, err := os.ReadFile(runtimeDir + "/cli/Dockerfile")
	if err != nil {
		return nil, errors.Wrap(err, "failed to read Dockerfile")
	}

	return append(libFiles, []types.FileResponse{
		{Name: "Dockerfile", Data: base64.StdEncoding.EncodeToString(dockerfileFile)},
		{Name: "Makefile", Data: base64.StdEncoding.EncodeToString(makefileFile)},
	}...), nil
}
