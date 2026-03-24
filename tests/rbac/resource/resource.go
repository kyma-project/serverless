package resource

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"strings"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	k8syaml "k8s.io/apimachinery/pkg/util/yaml"
)

// TestResource describes a Kubernetes resource that the RBAC test will exercise.
type TestResource interface {
	// Name returns a human-readable identifier for log output (e.g. "functions.serverless.kyma-project.io").
	Name() string
	// Object returns a fresh sample object ready to be created in the given namespace.
	Object(namespace string) *unstructured.Unstructured
	// ObjectList returns an empty list object whose GVK matches Object, used for client.List.
	ObjectList() *unstructured.UnstructuredList
}

// yamlResource implements TestResource backed by raw YAML.
type yamlResource struct {
	raw map[string]interface{}
}

func (r *yamlResource) Name() string {
	kind, _, _ := unstructured.NestedString(r.raw, "kind")
	group := ""
	if api, ok := r.raw["apiVersion"].(string); ok {
		if parts := strings.SplitN(api, "/", 2); len(parts) == 2 {
			group = parts[0]
		}
	}
	if group != "" {
		return strings.ToLower(kind) + "s." + group
	}
	return strings.ToLower(kind) + "s"
}

func (r *yamlResource) Object(namespace string) *unstructured.Unstructured {
	// deep-copy by re-serialising
	obj := &unstructured.Unstructured{Object: deepCopyMap(r.raw)}
	obj.SetNamespace(namespace)
	// clear any server-set fields from a previous run
	obj.SetResourceVersion("")
	obj.SetUID("")
	return obj
}

func (r *yamlResource) ObjectList() *unstructured.UnstructuredList {
	list := &unstructured.UnstructuredList{}
	apiVersion, _ := r.raw["apiVersion"].(string)
	kind, _ := r.raw["kind"].(string)
	list.SetAPIVersion(apiVersion)
	list.SetKind(kind + "List")
	return list
}

// LoadFromFile reads a YAML file and returns one TestResource per.
func LoadFromFile(path string) ([]TestResource, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading resources file: %w", err)
	}
	return parse(data)
}

// parse splits YAML and returns TestResources.
func parse(data []byte) ([]TestResource, error) {
	var resources []TestResource
	decoder := k8syaml.NewYAMLToJSONDecoder(bytes.NewReader(data))
	for {
		var raw map[string]interface{}
		if err := decoder.Decode(&raw); err != nil {
			if err == io.EOF {
				break
			}
			return nil, fmt.Errorf("decoding YAML document: %w", err)
		}
		if len(raw) == 0 {
			continue // skip empty
		}
		resources = append(resources, &yamlResource{raw: raw})
	}
	if len(resources) == 0 {
		return nil, fmt.Errorf("no resources found in YAML")
	}
	return resources, nil
}

// deepCopyMap recursively copies a map[string]interface{}.
func deepCopyMap(m map[string]interface{}) map[string]interface{} {
	out := make(map[string]interface{}, len(m))
	for k, v := range m {
		switch val := v.(type) {
		case map[string]interface{}:
			out[k] = deepCopyMap(val)
		case []interface{}:
			out[k] = deepCopySlice(val)
		default:
			out[k] = v
		}
	}
	return out
}

func deepCopySlice(s []interface{}) []interface{} {
	out := make([]interface{}, len(s))
	for i, v := range s {
		switch val := v.(type) {
		case map[string]interface{}:
			out[i] = deepCopyMap(val)
		case []interface{}:
			out[i] = deepCopySlice(val)
		default:
			out[i] = v
		}
	}
	return out
}
