package runtime

import (
	"fmt"
	"testing"

	"github.com/kyma-project/serverless/components/buildless-serverless/api/v1alpha2"
	"github.com/kyma-project/serverless/components/buildless-serverless/internal/endpoint/types"
	"github.com/stretchr/testify/require"
)

const (
	runtimesDir       = "../../../../runtimes/"
	handlerData       = "test handler"
	handlerBase64Data = "dGVzdCBoYW5kbGVy"
)

func Test_readNodejsFiles(t *testing.T) {
	t.Run("read true nodejs22 runtime files", func(t *testing.T) {
		inline := &v1alpha2.InlineSource{
			Source:       handlerData,
			Dependencies: "{}",
		}
		runtimeDir := fmt.Sprintf("%s/%s", runtimesDir, "nodejs22")

		gotList, gotErr := readNodejsFiles(inline, runtimeDir)
		require.NoError(t, gotErr)
		require.Len(t, gotList, 11)
		requireFileWithName(t, gotList, "package.json")
		require.Contains(t, gotList, types.FileResponse{Name: "handler.js", Data: handlerBase64Data})
	})

	t.Run("read true nodejs20 runtime files", func(t *testing.T) {
		inline := &v1alpha2.InlineSource{
			Source:       handlerData,
			Dependencies: "{}",
		}
		runtimeDir := fmt.Sprintf("%s/%s", runtimesDir, "nodejs20")

		gotList, gotErr := readNodejsFiles(inline, runtimeDir)
		require.NoError(t, gotErr)
		require.Len(t, gotList, 11)
		requireFileWithName(t, gotList, "package.json")
		require.Contains(t, gotList, types.FileResponse{Name: "handler.js", Data: handlerBase64Data})
	})

	t.Run("runtime dir does not exist", func(t *testing.T) {
		inline := &v1alpha2.InlineSource{
			Source:       handlerData,
			Dependencies: "{}",
		}
		runtimeDir := fmt.Sprintf("%s/%s", runtimesDir, "nodejs")

		gotList, gotErr := readNodejsFiles(inline, runtimeDir)
		require.Error(t, gotErr)
		require.Nil(t, gotList)
	})
}

func Test_readPythonFiles(t *testing.T) {
	t.Run("read true python312 runtime files", func(t *testing.T) {
		inline := &v1alpha2.InlineSource{
			Source:       handlerData,
			Dependencies: "",
		}
		runtimeDir := fmt.Sprintf("%s/%s", runtimesDir, "python312")

		gotList, gotErr := readPythonFiles(inline, runtimeDir)
		require.NoError(t, gotErr)
		require.Len(t, gotList, 9)
		requireFileWithName(t, gotList, "requirements.txt")
		require.Contains(t, gotList, types.FileResponse{Name: "handler.py", Data: handlerBase64Data})
	})

	t.Run("runtime dir does not exist", func(t *testing.T) {
		inline := &v1alpha2.InlineSource{
			Source:       handlerData,
			Dependencies: "",
		}
		runtimeDir := fmt.Sprintf("%s/%s", runtimesDir, "python")

		gotList, gotErr := readPythonFiles(inline, runtimeDir)
		require.Error(t, gotErr)
		require.Nil(t, gotList)
	})
}

func requireFileWithName(t *testing.T, files []types.FileResponse, name string) {
	for _, f := range files {
		if f.Name == name {
			return
		}
	}
	require.Fail(t, fmt.Sprintf("file %s not found", name))
}
