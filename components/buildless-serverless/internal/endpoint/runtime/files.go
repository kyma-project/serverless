package runtime

import (
	"encoding/base64"
	"fmt"
	"os"
	"path"

	"github.com/kyma-project/serverless/components/buildless-serverless/api/v1alpha2"
	"github.com/kyma-project/serverless/components/buildless-serverless/internal/endpoint/packagejson"
	"github.com/kyma-project/serverless/components/buildless-serverless/internal/endpoint/types"
	"github.com/pkg/errors"
)

func ReadFiles(f *v1alpha2.Function) ([]types.FileResponse, error) {
	runtimeDir := fmt.Sprintf("runtimes/%s", f.Spec.Runtime)

	if f.HasPythonRuntime() {
		return readPythonFiles(f.Spec.Source.Inline, runtimeDir)
	}

	return readNodejsFiles(f.Spec.Source.Inline, runtimeDir)
}

func readNodejsFiles(inline *v1alpha2.InlineSource, runtimeDir string) ([]types.FileResponse, error) {
	commonFiles, err := readCommonFiles(runtimeDir)
	if err != nil {
		return nil, err
	}

	// read package.json and merge function dependencies
	packagejsonFile, err := os.ReadFile(runtimeDir + "/package.json")
	if err != nil {
		return nil, errors.Wrap(err, "failed to read package.json")
	}

	if inline.Dependencies != "" {
		packagejsonFile, err = packagejson.Merge([]byte(inline.Dependencies), packagejsonFile)
		if err != nil {
			return nil, errors.Wrap(err, "failed to merge package.json")
		}
	}

	// read server.mjs
	serverFile, err := os.ReadFile(runtimeDir + "/server.mjs")
	if err != nil {
		return nil, errors.Wrap(err, "failed to read server.mjs")
	}

	// read sdk/ directory (nodejs26 only — legacy runtimes have no sdk/)
	sdkFiles, err := readDirFiles(runtimeDir, "/sdk")
	if err != nil && !os.IsNotExist(err) {
		return nil, errors.Wrap(err, "failed to read sdk directory")
	}

	files := append(commonFiles, []types.FileResponse{
		{Name: "package.json", Data: base64.StdEncoding.EncodeToString(packagejsonFile)},
		{Name: "server.mjs", Data: base64.StdEncoding.EncodeToString(serverFile)},
		{Name: "handler.js", Data: base64.StdEncoding.EncodeToString([]byte(inline.Source))},
	}...)
	return append(files, sdkFiles...), nil
}

func readPythonFiles(inline *v1alpha2.InlineSource, runtimeDir string) ([]types.FileResponse, error) {
	commonFiles, err := readCommonFiles(runtimeDir)
	if err != nil {
		return nil, err
	}

	// read requirements.txt and append function dependencies
	requirementsFile, err := os.ReadFile(runtimeDir + "/requirements.txt")
	if err != nil {
		return nil, errors.Wrap(err, "failed to read requirements.txt")
	}

	if inline.Dependencies != "" {
		requirementsFile = []byte(fmt.Sprintf("%s\n%s", string(requirementsFile), inline.Dependencies))
	}

	// read server.py
	serverFile, err := os.ReadFile(runtimeDir + "/server.py")
	if err != nil {
		return nil, errors.Wrap(err, "failed to read server.py")
	}

	return append(commonFiles, []types.FileResponse{
		{Name: "requirements.txt", Data: base64.StdEncoding.EncodeToString(requirementsFile)},
		{Name: "server.py", Data: base64.StdEncoding.EncodeToString(serverFile)},
		{Name: "handler.py", Data: base64.StdEncoding.EncodeToString([]byte(inline.Source))},
	}...), nil
}

func readCommonFiles(runtimeDir string) ([]types.FileResponse, error) {
	// read lib files
	libFiles, err := readDirFiles(runtimeDir, "/lib")
	if err != nil {
		return nil, errors.Wrap(err, "failed to read lib directory")
	}

	// read .gitignore
	gitignoreFile, err := os.ReadFile(runtimeDir + "/.gitignore")
	if err != nil {
		return nil, errors.Wrap(err, "failed to read .gitignore")
	}

	// read .dockerignore
	dockerignoreFile, err := os.ReadFile(runtimeDir + "/.dockerignore")
	if err != nil {
		return nil, errors.Wrap(err, "failed to read .dockerignore")
	}

	// read README.md
	readmeFile, err := os.ReadFile(runtimeDir + "/README_template.md")
	if err != nil {
		return nil, errors.Wrap(err, "failed to read README.md")
	}

	// read Makefile
	makefileFile, err := os.ReadFile(runtimeDir + "/Makefile")
	if err != nil {
		return nil, errors.Wrap(err, "failed to read Makefile")
	}

	// read Dockerfile
	dockerfileFile, err := os.ReadFile(runtimeDir + "/Dockerfile")
	if err != nil {
		return nil, errors.Wrap(err, "failed to read Dockerfile")
	}

	return append(libFiles, []types.FileResponse{
		{Name: ".gitignore", Data: base64.StdEncoding.EncodeToString(gitignoreFile)},
		{Name: ".dockerignore", Data: base64.StdEncoding.EncodeToString(dockerignoreFile)},
		{Name: "README.md", Data: base64.StdEncoding.EncodeToString(readmeFile)},
		{Name: "Dockerfile", Data: base64.StdEncoding.EncodeToString(dockerfileFile)},
		{Name: "Makefile", Data: base64.StdEncoding.EncodeToString(makefileFile)},
	}...), nil
}

// readDirFiles returns the non-directory entries in runtimeDir/folder as FileResponses,
// each named "<folder>/<basename>".
func readDirFiles(runtimeDir, folder string) ([]types.FileResponse, error) {
	entries, err := os.ReadDir(path.Join(runtimeDir, folder))
	if err != nil {
		return nil, err
	}

	files := make([]types.FileResponse, 0, len(entries))
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		data, err := os.ReadFile(path.Join(runtimeDir, folder, e.Name()))
		if err != nil {
			return nil, errors.Wrapf(err, "failed to read file '%s'", e.Name())
		}
		files = append(files, types.FileResponse{
			Name: fmt.Sprintf("%s/%s", folder, e.Name()),
			Data: base64.StdEncoding.EncodeToString(data),
		})
	}
	return files, nil
}
