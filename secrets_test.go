// Copyright the templig contributors.
// SPDX-License-Identifier: MPL-2.0

package templig_test

import (
	"bytes"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"testing"

	"gopkg.in/yaml.v3"

	"github.com/AlphaOne1/templig"
)

func TestDefaultSecretRE(t *testing.T) {
	t.Parallel()

	tests := []struct {
		in    string
		match bool
	}{
		// 0
		{in: "hello", match: false},
		// 1
		{in: "secret", match: true},
		// 2
		{in: "pass", match: true},
		// 3
		{in: "passWord", match: true},
		// 4
		{in: "passWor", match: true},
		// 5
		{in: "past", match: false},
		// 6
		{in: "CeRtIfIcAtE", match: true},
	}

	SecretRE := regexp.MustCompile(templig.SecretDefaultRE)

	for k, v := range tests {
		t.Run(fmt.Sprintf("DefaultSecretRE-%d", k), func(t *testing.T) {
			t.Parallel()

			match := SecretRE.MatchString(v.in)

			if match != v.match {
				t.Errorf("expected %v to match secret, but got %v", v.in, match)
			}
		})
	}
}

func TestHideSecrets(t *testing.T) {
	t.Parallel()

	tests := []struct {
		in            any
		want          any
		hideStructure bool
	}{
		{ // 0
			in:            "hello",
			want:          "hello",
			hideStructure: true,
		},
		{ // 1
			in:            []string{"a", "b", "c"},
			want:          []string{"a", "b", "c"},
			hideStructure: true,
		},
		{ // 2
			in: map[string]any{
				"Hello": "World",
			},
			want: map[string]any{
				"Hello": "World",
			},
			hideStructure: true,
		},
		{ // 3
			in:            "secret",
			want:          "secret",
			hideStructure: true,
		},
		{ // 4
			in: map[string]any{
				"secret": "World",
			},
			want: map[string]any{
				"secret": "*****",
			},
			hideStructure: true,
		},
		{ // 5
			in: map[string]any{
				"connections": []any{
					map[string]any{
						"user": "us",
						"pass": "pa",
					},
				},
			},
			want: map[string]any{
				"connections": []any{
					map[string]any{
						"user": "us",
						"pass": "**",
					},
				},
			},
			hideStructure: true,
		},
		{ // 6
			in: map[any]any{
				1: []any{
					map[string]any{
						"user": "us",
						"pass": "pa",
					},
				},
			},
			want: map[any]any{
				1: []any{
					map[string]any{
						"user": "us",
						"pass": "**",
					},
				},
			},
			hideStructure: true,
		},
		{ // 7
			in: map[string]any{
				"secrets": map[string]any{
					"user": "us",
					"pass": "pa",
				},
			},
			want: map[any]any{
				"secrets": "*",
			},
			hideStructure: true,
		},
		{ // 8
			in: map[string]any{
				"connections": map[string]any{
					"user":    "us",
					"secrets": []string{"a", "b", "c"},
				},
			},
			want: map[string]any{
				"connections": map[string]any{
					"user":    "us",
					"secrets": "*",
				},
			},
			hideStructure: true,
		},
		{ // 9
			in: map[string]any{
				"connections": map[string]any{
					"user":    "us",
					"secrets": []string{"a", "bb", "ccc"},
				},
			},
			want: map[string]any{
				"connections": map[string]any{
					"user":    "us",
					"secrets": []string{"*", "**", "***"},
				},
			},
			hideStructure: false,
		},
		{ // 10
			in: map[string]any{
				"connections": map[string]any{
					"user": "us",
					"secrets": func() []string {
						result := make([]string, 0, 103)

						for i := range cap(result) {
							result = append(result, "*"+strconv.Itoa(i))
						}

						return result
					}(),
				},
			},
			want: map[string]any{
				"connections": map[string]any{
					"user": "us",
					"secrets": func() []string {
						result := make([]string, 0, 103)

						for i := range cap(result) {
							result = append(result, strings.Repeat("*", len(strconv.Itoa(i))+1))
						}

						return result
					}(),
				},
			},
			hideStructure: false,
		},
		{ // 11
			in: map[string]any{
				"connections": map[string]any{
					"user":    "us",
					"secrets": nil,
				},
			},
			want: map[string]any{
				"connections": map[string]any{
					"user":    "us",
					"secrets": "****",
				},
			},
			hideStructure: false,
		},
		{ // 12
			in: map[string]any{
				"connections": map[string]any{
					"user":    "us",
					"secrets": []any{"pass", nil},
				},
			},
			want: map[string]any{
				"connections": map[string]any{
					"user":    "us",
					"secrets": []string{"****", "****"},
				},
			},
			hideStructure: false,
		},
		{ // 13
			in: map[string]any{
				"connections": map[string]any{
					"user":    "us",
					"secrets": []string{"*12345678911234567892123456789312"},
				},
			},
			want: map[string]any{
				"connections": map[string]any{
					"user":    "us",
					"secrets": []string{"**33**"},
				},
			},
			hideStructure: false,
		},
	}

	for testNum, test := range tests {
		t.Run(fmt.Sprintf("HideSecrets-%d", testNum), func(t *testing.T) {
			gotBuf := bytes.Buffer{}
			wantBuf := bytes.Buffer{}

			node := yaml.Node{}
			encodeErr := node.Encode(test.in)

			if encodeErr != nil {
				t.Errorf("%v: could not encode value", testNum)
				return
			}

			templig.HideSecrets(&node, test.hideStructure)

			if err := yaml.NewEncoder(&gotBuf).Encode(&node); err != nil {
				t.Errorf("%v: Got error serializing got", testNum)
			}
			if err := yaml.NewEncoder(&wantBuf).Encode(test.want); err != nil {
				t.Errorf("%v: Got error serializing want", testNum)
			}

			if gotBuf.String() != wantBuf.String() {
				t.Errorf("%v: got %v\nbut wanted %v", testNum, gotBuf.String(), wantBuf.String())
			}
		})
	}
}

func TestHideSecretsNil(t *testing.T) {
	t.Parallel()

	var a *yaml.Node

	templig.HideSecrets(a, true)
}

func TestHideSecretAlias(t *testing.T) {
	t.Parallel()

	input := `
open: &ref |-
    value
pass: *ref`
	want := `open: &ref |-
    *****
pass: *ref
`

	node := &yaml.Node{}

	if decodeErr := yaml.NewDecoder(bytes.NewBufferString(input)).Decode(node); decodeErr != nil {
		t.Errorf("unexpted encode error: %v", decodeErr)

		return
	}

	buf := bytes.Buffer{}

	templig.HideSecrets(node, true)

	if encodeErr := yaml.NewEncoder(&buf).Encode(node); encodeErr != nil {
		t.Errorf("could not encode node: %v", encodeErr)
	}

	if buf.String() != want {
		t.Errorf("unexpected output:\n%v\nwanted:\n%v", buf.String(), want)
	}
}
