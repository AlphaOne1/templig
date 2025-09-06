package templig_test

// These tests use Go's standard "testing" package and yaml/v4 for encoding,
// consistent with existing tests in merge_test.go.

import (
	"bytes"
	"testing"

	"go.yaml.in/yaml/v4"

	"github.com/AlphaOne1/templig"
)

func TestMerge_EmptyCollectionsAndSequences(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		a       string
		b       string
		want    string
		wantErr bool
	}{
		{
			name: "empty sequence with non-empty sequence",
			a:    `[]`,
			b:    `["x", 1]`,
			want: `["x", 1]`,
		},
		{
			name: "non-empty sequence with empty sequence",
			a:    `["x", 1]`,
			b:    `[]`,
			want: `["x", 1]`,
		},
		{
			name: "both empty sequences",
			a:    `[]`,
			b:    `[]`,
			want: `[]`,
		},
		{
			name: "empty map with non-empty map",
			a:    `{}`,
			b:    `{"a": 1, "b": 2}`,
			want: `{"a": 1, "b": 2}`,
		},
		{
			name: "non-empty map with empty map",
			a:    `{"a": 1, "b": 2}`,
			b:    `{}`,
			want: `{"a": 1, "b": 2}`,
		},
		{
			name: "both empty maps",
			a:    `{}`,
			b:    `{}`,
			want: `{}`,
		},
		{
			name: "sequence of mappings concatenation",
			a:    `[{"k": 1}]`,
			b:    `[{"k": 2}]`,
			want: `[{"k": 1}, {"k": 2}]`,
		},
	}

	for i, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			tt.want += "\n"

			var nodeA, nodeB yaml.Node
			if err := yaml.Unmarshal([]byte(tt.a), &nodeA); err != nil {
				t.Fatalf("%d - %s: unmarshal a: %v", i, tt.name, err)
			}
			if err := yaml.Unmarshal([]byte(tt.b), &nodeB); err != nil {
				t.Fatalf("%d - %s: unmarshal b: %v", i, tt.name, err)
			}

			result, err := templig.MergeYAMLNodes(&nodeA, &nodeB)
			if tt.wantErr {
				if err == nil {
					t.Fatalf("%d - %s: expected error but got nil", i, tt.name)
				}
				return
			}
			if err != nil {
				t.Fatalf("%d - %s: merge error: %v", i, tt.name, err)
			}

			var buf bytes.Buffer
			if encErr := yaml.NewEncoder(&buf).Encode(result.Content[0]); encErr != nil {
				t.Fatalf("%d - %s: encode error: %v", i, tt.name, encErr)
			}

			if buf.String() != tt.want {
				t.Errorf("%d - %s: mismatch\nwant:\n%s\ngot:\n%s", i, tt.name, tt.want, buf.String())
			}
		})
	}
}

func TestMerge_TopLevelTypeMismatch(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		a    string
		b    string
	}{
		{
			name: "map vs sequence",
			a:    `{"a": 1}`,
			b:    `["x"]`,
		},
		{
			name: "sequence vs map",
			a:    `["x"]`,
			b:    `{"a": 1}`,
		},
		{
			name: "scalar vs map",
			a:    `"foo"`,
			b:    `{"a": 1}`,
		},
		{
			name: "map vs scalar",
			a:    `{"a": 1}`,
			b:    `"bar"`,
		},
	}

	for i, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			var nodeA, nodeB yaml.Node
			if err := yaml.Unmarshal([]byte(tt.a), &nodeA); err != nil {
				t.Fatalf("%d - %s: unmarshal a: %v", i, tt.name, err)
			}
			if err := yaml.Unmarshal([]byte(tt.b), &nodeB); err != nil {
				t.Fatalf("%d - %s: unmarshal b: %v", i, tt.name, err)
			}

			res, err := templig.MergeYAMLNodes(&nodeA, &nodeB)
			if err == nil {
				t.Fatalf("%d - %s: expected error for top-level type mismatch; got result: %#v", i, tt.name, res)
			}
			if res != nil {
				t.Fatalf("%d - %s: expected nil result on error", i, tt.name)
			}
		})
	}
}

func TestMerge_NestedSequenceConcatInMap(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		a    string
		b    string
		want string
	}{
		{
			name: "concat nested sequences while preserving other keys",
			a:    `{"arr": [1, 2], "k": "v"}`,
			b:    `{"arr": [3], "k": "v"}`,
			want: `{"arr": [1, 2, 3], "k": "v"}`,
		},
		{
			name: "concat nested sequences existing only in a",
			a:    `{"arr": [1, 2]}`,
			b:    `{}`,
			want: `{"arr": [1, 2]}`,
		},
		{
			name: "concat nested sequences existing only in b",
			a:    `{}`,
			b:    `{"arr": [3, 4]}`,
			want: `{"arr": [3, 4]}`,
		},
	}

	for i, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			tt.want += "\n"

			var nodeA, nodeB yaml.Node
			if err := yaml.Unmarshal([]byte(tt.a), &nodeA); err != nil {
				t.Fatalf("%d - %s: unmarshal a: %v", i, tt.name, err)
			}
			if err := yaml.Unmarshal([]byte(tt.b), &nodeB); err != nil {
				t.Fatalf("%d - %s: unmarshal b: %v", i, tt.name, err)
			}

			result, err := templig.MergeYAMLNodes(&nodeA, &nodeB)
			if err != nil {
				t.Fatalf("%d - %s: merge error: %v", i, tt.name, err)
			}

			var buf bytes.Buffer
			if encErr := yaml.NewEncoder(&buf).Encode(result.Content[0]); encErr != nil {
				t.Fatalf("%d - %s: encode error: %v", i, tt.name, encErr)
			}

			if buf.String() != tt.want {
				t.Errorf("%d - %s: mismatch\nwant:\n%s\ngot:\n%s", i, tt.name, tt.want, buf.String())
			}
		})
	}
}

func TestMerge_ScalarOverridesAndNestedTypeMismatch(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		a       string
		b       string
		want    string
		wantErr bool
	}{
		{
			name: "boolean override",
			a:    `{"flag": false}`,
			b:    `{"flag": true}`,
			want: `{"flag": true}`,
		},
		{
			name: "number override",
			a:    `{"n": 1}`,
			b:    `{"n": 2}`,
			want: `{"n": 2}`,
		},
		{
			name: "string override",
			a:    `{"s": "a"}`,
			b:    `{"s": "b"}`,
			want: `{"s": "b"}`,
		},
		{
			name:    "nested type mismatch (map vs scalar) should error",
			a:       `{"o": {"k": 1}}`,
			b:       `{"o": 4}`,
			wantErr: true,
		},
	}

	for i, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			if !tt.wantErr {
				tt.want += "\n"
			}

			var nodeA, nodeB yaml.Node
			if err := yaml.Unmarshal([]byte(tt.a), &nodeA); err != nil {
				t.Fatalf("%d - %s: unmarshal a: %v", i, tt.name, err)
			}
			if err := yaml.Unmarshal([]byte(tt.b), &nodeB); err != nil {
				t.Fatalf("%d - %s: unmarshal b: %v", i, tt.name, err)
			}

			result, err := templig.MergeYAMLNodes(&nodeA, &nodeB)
			if tt.wantErr {
				if err == nil {
					t.Fatalf("%d - %s: expected error but got nil", i, tt.name)
				}
				if result != nil {
					t.Fatalf("%d - %s: expected nil result on error", i, tt.name)
				}
				return
			}
			if err != nil {
				t.Fatalf("%d - %s: merge error: %v", i, tt.name, err)
			}

			var buf bytes.Buffer
			if encErr := yaml.NewEncoder(&buf).Encode(result.Content[0]); encErr != nil {
				t.Fatalf("%d - %s: encode error: %v", i, tt.name, encErr)
			}

			if buf.String() != tt.want {
				t.Errorf("%d - %s: mismatch\nwant:\n%s\ngot:\n%s", i, tt.name, tt.want, buf.String())
			}
		})
	}
}

func TestMerge_AliasDefinedInAUsedInB(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		a       string
		b       string
		want    string
		wantErr bool
	}{
		{
			name: "anchor defined in a and alias used in b",
			a: `
x: &ref
    a: 3`,
			b: `
y: *ref`,
			want: `x: &ref
    a: 3
y: *ref`,
			wantErr: false,
		},
	}

	for i, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			tt.want += "\n"

			var nodeA, nodeB yaml.Node
			if err := yaml.Unmarshal([]byte(tt.a), &nodeA); err != nil {
				t.Fatalf("%d - %s: unmarshal a: %v", i, tt.name, err)
			}
			if err := yaml.Unmarshal([]byte(tt.b), &nodeB); err != nil {
				t.Fatalf("%d - %s: unmarshal b: %v", i, tt.name, err)
			}

			result, err := templig.MergeYAMLNodes(&nodeA, &nodeB)
			if tt.wantErr {
				if err == nil {
					t.Fatalf("%d - %s: expected error but got nil", i, tt.name)
				}
				return
			}
			if err != nil {
				t.Fatalf("%d - %s: merge error: %v", i, tt.name, err)
			}

			var buf bytes.Buffer
			if encErr := yaml.NewEncoder(&buf).Encode(result.Content[0]); encErr != nil {
				t.Fatalf("%d - %s: encode error: %v", i, tt.name, encErr)
			}

			if buf.String() != tt.want {
				t.Errorf("%d - %s: mismatch\nwant:\n%s\ngot:\n%s", i, tt.name, tt.want, buf.String())
			}
		})
	}
}