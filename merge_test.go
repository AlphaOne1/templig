// SPDX-FileCopyrightText: 2025 The templig contributors.
// SPDX-License-Identifier: MPL-2.0

package templig_test

import (
	"bytes"
	"testing"

	"go.yaml.in/yaml/v4"

	"github.com/AlphaOne1/templig"
)

func TestMerge(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		a       string
		b       string
		want    string
		wantErr bool
	}{
		{ // 0
			name:    "sequence concat",
			a:       `["a", 2, 3, 4]`,
			b:       `["b", 5, 6]`,
			want:    `["a", 2, 3, 4, "b", 5, 6]`,
			wantErr: false,
		},
		{ // 1
			name:    "mapping concat",
			a:       `{"a": 2, "b": 3, "c": 4}`,
			b:       `{"d": 5, "e": 6}`,
			want:    `{"a": 2, "b": 3, "c": 4, "d": 5, "e": 6}`,
			wantErr: false,
		},
		{ // 2
			name:    "mapping concat with overlap",
			a:       `{"a": 2, "b": 3, "c": 4}`,
			b:       `{"b": 5, "c": 6, "d": 7}`,
			want:    `{"a": 2, "b": 5, "c": 6, "d": 7}`,
			wantErr: false,
		},
		{ // 3
			name:    "mapping recursive",
			a:       `{"a": {"a": 1, "b": 2}}`,
			b:       `{"a": {"a": 2, "c": 3}}`,
			want:    `{"a": {"a": 2, "b": 2, "c": 3}}`,
			wantErr: false,
		},
		{ // 4
			name:    "mapping recursive",
			a:       `{"a": {"a": 1, "b": 2}}`,
			b:       `{"a": ["a", 2]}`,
			want:    ``,
			wantErr: true,
		},
		{ // 5
			name: "alias contained",
			a: `
a: &ref
    a: 3
b: *ref`,
			b: `
a:
    a: 4`,
			want: `a: &ref
    a: 4
b: *ref`,
			wantErr: false,
		},
		{ // 6
			name: "alias modified",
			a: `
a: &ref
    a: 3
b: *ref`,
			b: `
b:
    a: 5`,
			want: `a: &ref
    a: 3
b:
    a: 5`,
			wantErr: false,
		},
		{ // 7
			name: "alias in merge",
			a: `
a: &ref
    a: 3
b: *ref`,
			b: `
x-blah: &blah
    a: 5
b: *blah`,
			want: `a: &ref
    a: 3
b:
    a: 5
x-blah: &blah
    a: 5`,
			wantErr: false,
		},
		{ // 8
			name: "anchor mismatch",
			a: `
a: &ref0
    a: 3`,
			b: `
a: &ref1
    a: 3`,
			want:    ``,
			wantErr: true,
		},
	}

	for testNum, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			var nodeA yaml.Node
			var nodeB yaml.Node

			test.want += "\n"

			if aErr := yaml.Unmarshal([]byte(test.a), &nodeA); aErr != nil {
				t.Errorf("%v - %v: could not unmarshal a: %v", testNum, test.name, aErr)
			}

			if bErr := yaml.Unmarshal([]byte(test.b), &nodeB); bErr != nil {
				t.Errorf("%v - %v: could not unmarshal b: %v", testNum, test.name, bErr)
			}

			result, resultErr := templig.MergeYAMLNodes(&nodeA, &nodeB)

			if !test.wantErr && resultErr != nil {
				t.Errorf("%v - %v: could not merge: %v", testNum, test.name, resultErr)

				return
			}

			if test.wantErr && resultErr == nil {
				t.Errorf("%v - %v: should not have been able to merge", testNum, test.name)

				return
			}

			if resultErr != nil {
				return
			}

			buf := bytes.Buffer{}

			buf.Reset()

			if err := yaml.NewEncoder(&buf).Encode(result.Content[0]); err != nil {
				t.Errorf("%v - %v: could not marshal document to yaml: %v",
					testNum, test.name, err)
			}

			if buf.String() != test.want {
				t.Errorf("%v - %v: result mismatch, wanted:\n%v\nbut got:\n%v",
					testNum, test.name, test.want, buf.String())
			}
		})
	}
}

func TestStrangeDocumentNodes(t *testing.T) {
	t.Parallel()

	nodeA := yaml.Node{
		Kind: yaml.DocumentNode,
	}
	nodeB := yaml.Node{
		Kind: yaml.DocumentNode,
		Content: []*yaml.Node{
			{
				Kind: yaml.ScalarNode,
			},
		},
	}

	res, resErr := templig.MergeYAMLNodes(&nodeA, &nodeB)

	if res != nil {
		t.Errorf("merging strange document nodes must produce nil")
	}

	if resErr == nil {
		t.Errorf("expected an error merging strange document nodes")
	}
}

func TestUnknownNodeKind(t *testing.T) {
	t.Parallel()

	a := yaml.Node{
		Kind: 0,
	}
	b := yaml.Node{
		Kind: 0,
	}

	res, resErr := templig.MergeYAMLNodes(&a, &b)

	if res != nil {
		t.Errorf("merging unknown node kind must produce nil")
	}

	if resErr == nil {
		t.Errorf("expected an error merging unknown node kind")
	}
}
