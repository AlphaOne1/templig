// SPDX-FileCopyrightText: 2026 The templig contributors.
// SPDX-License-Identifier: MPL-2.0

package templig_test

import (
	"bytes"
	"errors"
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
		{ // 9
			name: "null replaced by mapping",
			a: `
a:
b: 1`,
			b: `
a:
    x: 1`,
			want: `a:
    x: 1
b: 1`,
			wantErr: false,
		},
		{ // 10
			name: "null replaced by sequence",
			a: `
a: ~`,
			b: `
a:
    - 1`,
			want: `a:
    - 1`,
			wantErr: false,
		},
		{ // 11
			name: "mapping replaced by null",
			a: `
a:
    x: 1
b: 1`,
			b: `
a: null`,
			want: `a: null
b: 1`,
			wantErr: false,
		},
		{ // 12
			name: "null replaced by alias",
			a: `
a: null`,
			b: `
r: &r
    k: 1
a: *r`,
			want: `a:
    k: 1
r: &r
    k: 1`,
			wantErr: false,
		},
		{ // 13
			name: "empty override",
			a: `
a: 1`,
			b:       ``,
			want:    `a: 1`,
			wantErr: false,
		},
		{ // 14
			name:    "empty base",
			a:       `# only a comment`,
			b:       `a: 1`,
			want:    `a: 1`,
			wantErr: false,
		},
		{ // 15
			name:    "keys of different type",
			a:       `1: a`,
			b:       `"1": b`,
			want:    "1: a\n\"1\": b",
			wantErr: false,
		},
		{ // 16
			name:    "int keys in different notation",
			a:       `1: a`,
			b:       `0x1: b`,
			want:    `1: b`,
			wantErr: false,
		},
		{ // 17
			name:    "bool keys in different notation",
			a:       `true: a`,
			b:       `True: b`,
			want:    `true: b`,
			wantErr: false,
		},
		{ // 18
			name: "alias as key",
			a:    `a: 1`,
			b: `
k: &k a
*k : 2`,
			want:    "a: 2\nk: &k a",
			wantErr: false,
		},
		{ // 19
			name:    "non-scalar keys are not matched",
			a:       `[x]: 1`,
			b:       `[x]: 2`,
			want:    "? [x]\n: 1\n? [x]\n: 2",
			wantErr: false,
		},
		{ // 20
			name:    "undecodable keys compared by tag and value",
			a:       `!!int abc: 1`,
			b:       `!!int abc: 2`,
			want:    `!!int abc: 2`,
			wantErr: false,
		},
		{ // 21
			name: "kind mismatch through alias",
			a:    `a: [1]`,
			b: `
r: &r {k: 1}
a: *r`,
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

	// Kind 0 denotes an empty node, so an otherwise undefined kind is used
	a := yaml.Node{
		Kind: yaml.Kind(255),
	}
	b := yaml.Node{
		Kind: yaml.Kind(255),
	}

	res, resErr := templig.MergeYAMLNodes(&a, &b)

	if res != nil {
		t.Errorf("merging unknown node kind must produce nil")
	}

	if resErr == nil {
		t.Errorf("expected an error merging unknown node kind")
	}
}

func TestAliasWithoutTarget(t *testing.T) {
	t.Parallel()

	a := yaml.Node{
		Kind:  yaml.ScalarNode,
		Value: "a",
	}
	b := yaml.Node{
		Kind: yaml.AliasNode,
	}

	res, resErr := templig.MergeYAMLNodes(&a, &b)

	if res != nil {
		t.Errorf("merging an alias without target must produce nil")
	}

	if !errors.Is(resErr, templig.ErrNodeNil) {
		t.Errorf("expected ErrNodeNil merging an alias without target, but got %v", resErr)
	}
}
