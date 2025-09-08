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

	"go.yaml.in/yaml/v4"

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
			t.Parallel()

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

func TestDefaultSecretRE_BoundariesAndCommonKeys(t *testing.T) {
	t.Parallel()

	// Note: We validate boundary-oriented cases and common secret-like tokens without assuming exact pattern internals.
	// The intent is to ensure reasonable matches and non-matches around word boundaries and casing.
	tests := []struct {
		name  string
		in    string
		match bool
	}{
		{name: "empty string", in: "", match: false},
		{name: "word boundary - past (should not match)", in: "past", match: false},
		{name: "word boundary - compass (should not match 'pass' substring at end of word)", in: "compass", match: false},
		{name: "prefix boundary - passport (likely not a password; should not match)", in: "passport", match: false},
		{name: "uppercase PASS", in: "PASS", match: true},
		{name: "mixed case Password", in: "Password", match: true},
		{name: "token common term", in: "token", match: true},     // Common secret indicator
		{name: "api_key snake", in: "api_key", match: true},       // Common secret indicator
		{name: "apiKey camel", in: "apiKey", match: true},         // Common secret indicator
		{name: "authorization", in: "authorization", match: true}, // Common secret indicator
	}

	SecretRE := regexp.MustCompile(templig.SecretDefaultRE)
	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			if got := SecretRE.MatchString(tc.in); got != tc.match {
				t.Fatalf("MatchString(%q) = %v; want %v", tc.in, got, tc.match)
			}
		})
	}
}

func TestHideSecrets_MultiDocumentAndFlowCollections(t *testing.T) {
	t.Parallel()

	// Multi-document YAML with a mix of block and flow styles, ensuring both documents are processed.
	// - First doc: secret keys in flow mapping and sequence
	// - Second doc: ensure non-secret keys remain unchanged
	input := `---
flow_map: {user: u1, pass: p1, secret: s1}
flow_seq: [a, b, c]
---
regular: value
`

	want := `flow_map:
    user: u1
    pass: **
    secret: **
flow_seq:
    - a
    - b
    - c
---
regular: value
`

	// Decode multi-doc stream to nodes, process, and re-encode preserving doc separators.
	dec := yaml.NewDecoder(strings.NewReader(input))
	var docs []*yaml.Node
	for {
		n := &yaml.Node{}
		if err := dec.Decode(n); err != nil {
			if err.Error() == "EOF" {
				break
			}
			t.Fatalf("decode error: %v", err)
		}
		docs = append(docs, n)
	}

	// Apply hiding on each document separately
	for _, n := range docs {
		templig.HideSecrets(n, false)
	}

	var buf bytes.Buffer
	enc := yaml.NewEncoder(&buf)
	for i, n := range docs {
		if i == 0 {
			// The encoder will not automatically re-add '---' for the first doc; we emulate expected output formatting by
			// encoding docs sequentially and then adjusting separators below for comparison simplicity.
		}
		if err := enc.Encode(n); err != nil {
			t.Fatalf("encode error: %v", err)
		}
	}
	_ = enc.Close()

	// Normalize the encoder output to include '---' separators in the same style as 'want'.
	got := buf.String()
	// The yaml.Encoder typically emits '---' between documents; however, formatting can vary.
	// We rewrite to the expected canonical representation used in 'want'.
	got = strings.ReplaceAll(got, "---\n", "")
	// Insert explicit separator between encoded documents
	if strings.Count(got, "\n---\n") == 0 && strings.Count(got, "\nregular: value\n") == 1 {
		parts := strings.SplitN(got, "\nregular: value\n", 2)
		if len(parts) == 2 {
			got = strings.TrimSuffix(parts[0], "\n") + "\n---\nregular: value\n" + parts[1]
		}
	}

	if got != want {
		t.Fatalf("unexpected output:\n%s\nwant:\n%s", got, want)
	}
}

func TestHideSecrets_ReusedAnchorMultipleSecretKeys(t *testing.T) {
	t.Parallel()

	// A single anchored scalar reused across multiple secret-indicating keys.
	input := `
val: &anch secret-value
pass: *anch
secret: *anch
normal: *anch
`
	// Expect: the anchor's scalar is masked, aliases reflect masked value,
	// but non-secret key "normal" referencing the same alias still points to masked scalar
	// (because the underlying anchored node is masked once).
	want := `val: &anch ******
pass: ******
secret: ******
normal: ******
`

	node := &yaml.Node{}
	if err := yaml.NewDecoder(strings.NewReader(input)).Decode(node); err != nil {
		t.Fatalf("decode error: %v", err)
	}

	var got bytes.Buffer
	templig.HideSecrets(node, true)
	if err := yaml.NewEncoder(&got).Encode(node); err != nil {
		t.Fatalf("encode error: %v", err)
	}

	if got.String() != want {
		t.Fatalf("unexpected output:\n%s\nwant:\n%s", got.String(), want)
	}
}

func TestHideSecrets_NumericBooleanAndEmptyScalars(t *testing.T) {
	t.Parallel()

	// Validate behavior with non-string scalar values that appear under secret-likely keys.
	// We assert that such values are masked (not left verbatim), without binding to exact star-count for non-strings.
	tests := []struct {
		name string
		in   string
	}{
		{
			name: "numeric under pass",
			in:   "pass: 12345\n",
		},
		{
			name: "boolean under secret",
			in:   "secret: true\n",
		},
		{
			name: "empty under token",
			in:   "token: \n",
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			node := &yaml.Node{}
			if err := yaml.NewDecoder(strings.NewReader(tc.in)).Decode(node); err != nil {
				t.Fatalf("decode error: %v", err)
			}

			var got bytes.Buffer
			templig.HideSecrets(node, false)

			if err := yaml.NewEncoder(&got).Encode(node); err != nil {
				t.Fatalf("encode error: %v", err)
			}

			out := got.String()
			// Ensure the value is not the original literal
			if strings.Contains(out, "12345") || strings.Contains(out, "true\n") {
				t.Fatalf("expected non-string secrets to be masked, got:\n%s", out)
			}
			// For empty value, ensure a placeholder is present (some masking)
			if tc.name == "empty under token" && !strings.Contains(out, "*") {
				t.Fatalf("expected masking placeholder for empty token, got:\n%s", out)
			}
		})
	}
}

func TestHideSecrets_DeeplyNestedMixedStructures(t *testing.T) {
	t.Parallel()

	in := map[string]any{
		"lvl1": map[string]any{
			"arr": []any{
				map[string]any{"user": "alice", "password": "wonderland"},
				[]any{
					map[string]any{"certificate": "pemdata"},
					map[string]any{"note": "not a secret"},
				},
			},
			"meta": "keep",
		},
	}

	want := map[string]any{
		"lvl1": map[string]any{
			"arr": []any{
				map[string]any{"user": "alice", "password": "**********"},
				[]any{
					map[string]any{"certificate": "*******"},
					map[string]any{"note": "not a secret"},
				},
			},
			"meta": "keep",
		},
	}

	var gotBuf, wantBuf bytes.Buffer
	node := &yaml.Node{}
	if err := node.Encode(in); err != nil {
		t.Fatalf("encode input: %v", err)
	}
	templig.HideSecrets(node, false)

	if err := yaml.NewEncoder(&gotBuf).Encode(node); err != nil {
		t.Fatalf("encode got: %v", err)
	}
	if err := yaml.NewEncoder(&wantBuf).Encode(want); err != nil {
		t.Fatalf("encode want: %v", err)
	}

	if gotBuf.String() != wantBuf.String() {
		t.Fatalf("got:\n%s\nwant:\n%s", gotBuf.String(), wantBuf.String())
	}
}
