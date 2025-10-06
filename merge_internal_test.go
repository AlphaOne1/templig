// SPDX-FileCopyrightText: 2025 The templig contributors.
// SPDX-License-Identifier: MPL-2.0

package templig

import (
	"testing"

	"go.yaml.in/yaml/v4"
)

func TestNullArgs(t *testing.T) {
	t.Parallel()

	mergeFuncs := []func(*yaml.Node, *yaml.Node) (*yaml.Node, error){
		MergeYAMLNodes,
		mergeAliasNodes,
		mergeDocumentNodes,
		mergeMappingNodes,
		mergeScalarNodes,
		mergeSequenceNodes,
	}

	for k, v := range mergeFuncs {
		res, resErr := v(nil, nil)

		if res != nil {
			t.Errorf("%v: merging nil nodes must produce nil", k)
		}

		if resErr == nil {
			t.Errorf("%v: expected an error merging nil nodes", k)
		}
	}
}

func TestMismatchArgs(t *testing.T) {
	t.Parallel()

	mergeFuncs := []func(*yaml.Node, *yaml.Node) (*yaml.Node, error){
		MergeYAMLNodes,
		mergeAliasNodes,
		mergeDocumentNodes,
		mergeMappingNodes,
		mergeScalarNodes,
		mergeSequenceNodes,
	}

	a := yaml.Node{
		Kind: yaml.MappingNode,
	}
	b := yaml.Node{
		Kind: yaml.DocumentNode,
	}

	for k, v := range mergeFuncs {
		res, resErr := v(&a, &b)

		if res != nil {
			t.Errorf("%v: merging kind-mismatched nodes must produce nil", k)
		}

		if resErr == nil {
			t.Errorf("%v: expected an error merging kind-mismatched nodes", k)
		}
	}
}

func TestMergeMappingNodes_BasicAndNested(t *testing.T) {
	t.Parallel()

	mk := func(k, v string) []*yaml.Node {
		return []*yaml.Node{{Kind: yaml.ScalarNode, Tag: "!!str", Value: k}, {Kind: yaml.ScalarNode, Tag: "!!str", Value: v}}
	}

	left := &yaml.Node{Kind: yaml.MappingNode, Content: append([]*yaml.Node{}, mk("a", "1")...)}
	right := &yaml.Node{Kind: yaml.MappingNode, Content: append([]*yaml.Node{}, mk("b", "2")...)}

	merged, err := mergeMappingNodes(left, right)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if merged == nil || merged.Kind != yaml.MappingNode {
		t.Fatalf("expected mapping node, got %#v", merged)
	}

	// Expect both keys to be present
	if len(merged.Content) != 4 {
		t.Fatalf("expected 2 key-value pairs (4 nodes), got %d", len(merged.Content))
	}

	// Deep/nested merge: left.n = { x: 1 }, right.n = { y: 2 }
	leftN := &yaml.Node{Kind: yaml.MappingNode, Content: append([]*yaml.Node{}, &yaml.Node{Kind: yaml.ScalarNode, Tag: "!!str", Value: "n"}, &yaml.Node{Kind: yaml.MappingNode, Content: append([]*yaml.Node{}, mk("x", "1")...)})}
	rightN := &yaml.Node{Kind: yaml.MappingNode, Content: append([]*yaml.Node{}, &yaml.Node{Kind: yaml.ScalarNode, Tag: "!!str", Value: "n"}, &yaml.Node{Kind: yaml.MappingNode, Content: append([]*yaml.Node{}, mk("y", "2")...)})}

	merged2, err := mergeMappingNodes(leftN, rightN)
	if err != nil {
		t.Fatalf("unexpected error in nested merge: %v", err)
	}
	if merged2 == nil || merged2.Kind != yaml.MappingNode {
		t.Fatalf("expected mapping node for nested merge, got %#v", merged2)
	}

	// Find nested map under key "n" and ensure both x and y exist
	var nested *yaml.Node
	for i := 0; i+1 < len(merged2.Content); i += 2 {
		k := merged2.Content[i]
		v := merged2.Content[i+1]
		if k.Value == "n" {
			nested = v
			break
		}
	}
	if nested == nil || nested.Kind != yaml.MappingNode {
		t.Fatalf("expected nested mapping under key n, got %#v", nested)
	}
	// nested should contain keys x and y
	seen := map[string]bool{}
	for i := 0; i+1 < len(nested.Content); i += 2 {
		seen[nested.Content[i].Value] = true
	}
	if !(seen["x"] && seen["y"]) {
		t.Fatalf("expected nested keys x and y, got %#v", seen)
	}
}

func TestMergeMappingNodes_KeyConflict(t *testing.T) {
	t.Parallel()

	mk := func(k, v string) []*yaml.Node {
		return []*yaml.Node{{Kind: yaml.ScalarNode, Tag: "!!str", Value: k}, {Kind: yaml.ScalarNode, Tag: "!!str", Value: v}}
	}
	left := &yaml.Node{Kind: yaml.MappingNode, Content: append([]*yaml.Node{}, mk("a", "left")...)}
	right := &yaml.Node{Kind: yaml.MappingNode, Content: append([]*yaml.Node{}, mk("a", "right")...)}

	merged, err := mergeMappingNodes(left, right)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Ensure conflict is resolved deterministically (implementation-defined).
	// Accept either left or right value but must be exactly one value for key "a".
	countA := 0
	var val string
	for i := 0; i+1 < len(merged.Content); i += 2 {
		if merged.Content[i].Value == "a" {
			countA++
			val = merged.Content[i+1].Value
		}
	}
	if countA != 1 {
		t.Fatalf("expected one value for key a, got %d", countA)
	}
	if val != "left" && val != "right" {
		t.Fatalf("unexpected merged value for key a: %q", val)
	}
}


func TestMergeSequenceNodes_Various(t *testing.T) {
	t.Parallel()

	mkScalar := func(v string) *yaml.Node { return &yaml.Node{Kind: yaml.ScalarNode, Tag: "!!str", Value: v} }
	left := &yaml.Node{Kind: yaml.SequenceNode, Content: []*yaml.Node{mkScalar("a"), mkScalar("b")}}
	right := &yaml.Node{Kind: yaml.SequenceNode, Content: []*yaml.Node{mkScalar("c")}}

	merged, err := mergeSequenceNodes(left, right)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if merged == nil || merged.Kind != yaml.SequenceNode {
		t.Fatalf("expected sequence node, got %#v", merged)
	}
	if len(merged.Content) < 1 {
		t.Fatalf("merged sequence should not be empty")
	}

	// Merge with empty sequence
	empty := &yaml.Node{Kind: yaml.SequenceNode, Content: nil}
	merged2, err := mergeSequenceNodes(left, empty)
	if err != nil {
		t.Fatalf("unexpected error merging with empty: %v", err)
	}
	if merged2 == nil || merged2.Kind != yaml.SequenceNode {
		t.Fatalf("expected sequence node after empty merge")
	}
}


func TestMergeScalarNodes_Simple(t *testing.T) {
	t.Parallel()

	left := &yaml.Node{Kind: yaml.ScalarNode, Tag: "!!str", Value: "foo"}
	right := &yaml.Node{Kind: yaml.ScalarNode, Tag: "!!str", Value: "bar"}

	merged, err := mergeScalarNodes(left, right)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if merged == nil || merged.Kind != yaml.ScalarNode {
		t.Fatalf("expected scalar node, got %#v", merged)
	}
	if merged.Value != "foo" && merged.Value != "bar" {
		t.Fatalf("unexpected merged scalar value: %q", merged.Value)
	}
}

func TestMergeScalarNodes_TypeMismatch(t *testing.T) {
	t.Parallel()

	left := &yaml.Node{Kind: yaml.ScalarNode, Tag: "!!str", Value: "1"}
	right := &yaml.Node{Kind: yaml.ScalarNode, Tag: "!!int", Value: "1"}

	res, err := mergeScalarNodes(left, right)
	if err == nil {
		t.Fatalf("expected error merging mismatched scalar tags, got res=%#v", res)
	}
	if res != nil {
		t.Fatalf("expected nil result on mismatched scalar tags")
	}
}


func TestMergeAliasNodes_Basic(t *testing.T) {
	t.Parallel()

	target := &yaml.Node{Kind: yaml.ScalarNode, Tag: "!!str", Value: "val"}
	left := &yaml.Node{Kind: yaml.AliasNode, Alias: target}
	right := &yaml.Node{Kind: yaml.AliasNode, Alias: target}

	merged, err := mergeAliasNodes(left, right)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if merged == nil || merged.Kind != yaml.AliasNode {
		t.Fatalf("expected alias node, got %#v", merged)
	}
	if merged.Alias == nil {
		t.Fatalf("merged alias should reference a target")
	}
}

func TestMergeDocumentNodes_RootContent(t *testing.T) {
	t.Parallel()

	mk := func(k, v string) []*yaml.Node {
		return []*yaml.Node{{Kind: yaml.ScalarNode, Tag: "!!str", Value: k}, {Kind: yaml.ScalarNode, Tag: "!!str", Value: v}}
	}
	leftDoc := &yaml.Node{Kind: yaml.DocumentNode, Content: []*yaml.Node{{Kind: yaml.MappingNode, Content: append([]*yaml.Node{}, mk("a", "1")...)}}}
	rightDoc := &yaml.Node{Kind: yaml.DocumentNode, Content: []*yaml.Node{{Kind: yaml.MappingNode, Content: append([]*yaml.Node{}, mk("b", "2")...)}}}

	merged, err := mergeDocumentNodes(leftDoc, rightDoc)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if merged == nil || merged.Kind != yaml.DocumentNode {
		t.Fatalf("expected document node, got %#v", merged)
	}
	if len(merged.Content) == 0 || merged.Content[0].Kind != yaml.MappingNode {
		t.Fatalf("expected document root content to be a mapping")
	}
}


func TestMergeYAMLNodes_EndToEnd(t *testing.T) {
	t.Parallel()

	lhs := &yaml.Node{Kind: yaml.DocumentNode, Content: []*yaml.Node{{
		Kind: yaml.MappingNode, Content: []*yaml.Node{
			{Kind: yaml.ScalarNode, Tag: "!!str", Value: "name"}, {Kind: yaml.ScalarNode, Tag: "!!str", Value: "left"},
			{Kind: yaml.ScalarNode, Tag: "!!str", Value: "items"}, {Kind: yaml.SequenceNode, Content: []*yaml.Node{{Kind: yaml.ScalarNode, Tag: "!!str", Value: "a"}}},
		},
	}}}

	rhs := &yaml.Node{Kind: yaml.DocumentNode, Content: []*yaml.Node{{
		Kind: yaml.MappingNode, Content: []*yaml.Node{
			{Kind: yaml.ScalarNode, Tag: "!!str", Value: "name"}, {Kind: yaml.ScalarNode, Tag: "!!str", Value: "right"},
			{Kind: yaml.ScalarNode, Tag: "!!str", Value: "extra"}, {Kind: yaml.ScalarNode, Tag: "!!str", Value: "field"},
			{Kind: yaml.ScalarNode, Tag: "!!str", Value: "items"}, {Kind: yaml.SequenceNode, Content: []*yaml.Node{{Kind: yaml.ScalarNode, Tag: "!!str", Value: "b"}}},
		},
	}}}

	merged, err := MergeYAMLNodes(lhs, rhs)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if merged == nil || merged.Kind != yaml.DocumentNode {
		t.Fatalf("expected document node, got %#v", merged)
	}
	root := merged.Content[0]
	if root == nil || root.Kind != yaml.MappingNode {
		t.Fatalf("expected mapping at document root")
	}
	// Validate expected keys exist: name, items, extra
	seen := map[string]*yaml.Node{}
	for i := 0; i+1 < len(root.Content); i += 2 {
		seen[root.Content[i].Value] = root.Content[i+1]
	}
	if _, ok := seen["name"]; !ok {
		t.Fatalf("expected key name")
	}
	if _, ok := seen["items"]; !ok {
		t.Fatalf("expected key items")
	}
	if _, ok := seen["extra"]; !ok {
		t.Fatalf("expected key extra")
	}
	if seen["items"].Kind != yaml.SequenceNode || len(seen["items"].Content) == 0 {
		t.Fatalf("expected non-empty items sequence")
	}
}

func TestMergeYAMLNodes_MismatchError(t *testing.T) {
	t.Parallel()
	lhs := &yaml.Node{Kind: yaml.MappingNode}
	rhs := &yaml.Node{Kind: yaml.SequenceNode}
	res, err := MergeYAMLNodes(lhs, rhs)
	if err == nil {
		t.Fatalf("expected error for mismatched kinds, got res=%#v", res)
	}
	if res != nil {
		t.Fatalf("expected nil result for mismatched kinds")
	}
}

