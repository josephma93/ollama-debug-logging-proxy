package redact

import (
	"encoding/json"
	"strings"
)

const (
	imagesFieldName = "images"
	RedactedValue   = "[redacted]"
)

// RedactImagesJSON replaces JSON object fields named "images" (case-insensitive)
// with the string value "[redacted]" at any nesting depth.
//
// If the input is malformed or non-JSON, the original bytes are returned with
// redacted=false.
func RedactImagesJSON(input []byte) ([]byte, bool) {
	var root any
	if err := json.Unmarshal(input, &root); err != nil {
		return input, false
	}

	updated, changed := redactValue(root)
	if !changed {
		return input, false
	}

	output, err := json.Marshal(updated)
	if err != nil {
		return input, false
	}

	return output, true
}

// RedactImagesString is a string convenience wrapper around RedactImagesJSON.
func RedactImagesString(input string) (string, bool) {
	output, redacted := RedactImagesJSON([]byte(input))
	return string(output), redacted
}

func redactValue(value any) (any, bool) {
	switch typed := value.(type) {
	case map[string]any:
		changed := false
		for key, child := range typed {
			if strings.EqualFold(key, imagesFieldName) {
				typed[key] = RedactedValue
				changed = true
				continue
			}

			updatedChild, childChanged := redactValue(child)
			if childChanged {
				typed[key] = updatedChild
				changed = true
			}
		}
		return typed, changed
	case []any:
		changed := false
		for i, child := range typed {
			updatedChild, childChanged := redactValue(child)
			if childChanged {
				typed[i] = updatedChild
				changed = true
			}
		}
		return typed, changed
	default:
		return value, false
	}
}
