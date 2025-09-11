package packagejson

import (
	"encoding/json"

	"github.com/pkg/errors"
)

func Merge(from, target []byte) ([]byte, error) {
	fromObj := map[string]any{}
	err := json.Unmarshal(from, &fromObj)
	if err != nil {
		return nil, errors.Wrap(err, "failed to unmarshal source package.json")
	}

	targetObj := map[string]any{}
	err = json.Unmarshal(target, &targetObj)
	if err != nil {
		return nil, errors.Wrap(err, "failed to unmarshal target package.json")
	}

	merged := merge(fromObj, targetObj)
	return json.MarshalIndent(merged, "", "  ")
}

func merge(from, target map[string]any) map[string]any {
	for k, v := range from {
		if _, exists := target[k]; !exists {
			target[k] = v
			continue
		}

		// both are maps, merge them recursively
		fromMap, okFrom := v.(map[string]any)
		targetMap, okTarget := target[k].(map[string]any)
		if okFrom && okTarget {
			target[k] = merge(fromMap, targetMap)
			continue
		}

		// both are arrays, merge them uniquely
		fromArr, okFrom := v.([]any)
		targetArr, okTarget := target[k].([]any)
		if okFrom && okTarget {
			target[k] = mergeArrays(fromArr, targetArr)
			continue
		}

		// otherwise, override the value
		target[k] = v
	}
	return target
}

func mergeArrays(from, target []any) []any {
	existing := map[any]bool{}
	for _, v := range target {
		existing[v] = true
	}

	for _, v := range from {
		if _, exists := existing[v]; !exists {
			target = append(target, v)
		}
	}
	return target
}
