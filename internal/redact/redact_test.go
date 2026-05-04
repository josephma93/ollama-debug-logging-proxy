package redact

import (
	"encoding/json"
	"testing"
)

func TestRedactImagesJSONRecursiveCaseInsensitive(t *testing.T) {
	input := []byte(`{
		"model":"llava",
		"images":["a"],
		"nested":{
			"Images":["b"],
			"level2":[{"IMAGES":["c"]},{"keep":"value"}]
		},
		"items":[
			{"prompt":"x"},
			{"images":["d"]}
		]
	}`)

	output, redacted := RedactImagesJSON(input)
	if !redacted {
		t.Fatalf("expected redacted=true")
	}

	var got map[string]any
	if err := json.Unmarshal(output, &got); err != nil {
		t.Fatalf("unmarshal output: %v", err)
	}

	if got["images"] != RedactedValue {
		t.Fatalf("top-level images not redacted: %#v", got["images"])
	}

	nested, ok := got["nested"].(map[string]any)
	if !ok {
		t.Fatalf("nested should be object")
	}
	if nested["Images"] != RedactedValue {
		t.Fatalf("nested Images not redacted: %#v", nested["Images"])
	}

	level2, ok := nested["level2"].([]any)
	if !ok || len(level2) == 0 {
		t.Fatalf("nested.level2 should be non-empty array")
	}
	firstObj, ok := level2[0].(map[string]any)
	if !ok {
		t.Fatalf("nested.level2[0] should be object")
	}
	if firstObj["IMAGES"] != RedactedValue {
		t.Fatalf("nested.level2[0].IMAGES not redacted: %#v", firstObj["IMAGES"])
	}

	items, ok := got["items"].([]any)
	if !ok || len(items) < 2 {
		t.Fatalf("items should contain at least two elements")
	}
	secondObj, ok := items[1].(map[string]any)
	if !ok {
		t.Fatalf("items[1] should be object")
	}
	if secondObj["images"] != RedactedValue {
		t.Fatalf("items[1].images not redacted: %#v", secondObj["images"])
	}
}

func TestRedactImagesJSONMalformedPassthrough(t *testing.T) {
	tests := []struct {
		name  string
		input []byte
	}{
		{name: "truncated json", input: []byte(`{"images":[`)},
		{name: "plain string", input: []byte(`not-json`)},
		{name: "empty", input: []byte{}},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			output, redacted := RedactImagesJSON(tt.input)
			if redacted {
				t.Fatalf("expected redacted=false")
			}
			if string(output) != string(tt.input) {
				t.Fatalf("expected passthrough output, got=%q want=%q", string(output), string(tt.input))
			}
		})
	}
}

func TestRedactImagesJSONValidNoImages(t *testing.T) {
	input := []byte(`{"model":"llama3","prompt":"hello"}`)
	output, redacted := RedactImagesJSON(input)

	if redacted {
		t.Fatalf("expected redacted=false")
	}
	if string(output) != string(input) {
		t.Fatalf("expected unchanged output, got=%q want=%q", string(output), string(input))
	}
}

func TestRedactImagesStringMalformedPassthrough(t *testing.T) {
	input := `{"images":[`
	output, redacted := RedactImagesString(input)

	if redacted {
		t.Fatalf("expected redacted=false")
	}
	if output != input {
		t.Fatalf("expected unchanged output, got=%q want=%q", output, input)
	}
}
