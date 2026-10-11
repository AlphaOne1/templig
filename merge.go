// SPDX-FileCopyrightText: 2026 The templig contributors.
// SPDX-License-Identifier: MPL-2.0

package templig

import (
	"errors"
	"fmt"
	"reflect"

	"go.yaml.in/yaml/v4"
)

var (

	// ErrNodeNil is an error returned when a provided node is nil.
	ErrNodeNil = errors.New("node is nil")

	// ErrNodeKindMismatch is an error returned when two nodes have incompatible kinds during an operation.
	ErrNodeKindMismatch = errors.New("node kind mismatch")

	// ErrNodeTypeUnhandled is an error returned when a node has an unhandled or unsupported type during processing.
	ErrNodeTypeUnhandled = errors.New("node type unhandled")

	// ErrAliasNodeExpected is an error returned when an operation
	// expects an alias node but encounters a different type.
	ErrAliasNodeExpected = errors.New("alias node expected")

	// ErrUnexpectedDocumentNodeConfiguration is an error returned
	// when a document node has an unexpected or invalid configuration.
	ErrUnexpectedDocumentNodeConfiguration = errors.New("unexpected document node configuration")

	// ErrUnequalNameAnchors is an error returned when operations on named anchors fail
	// due to unequal anchor definitions.
	ErrUnequalNameAnchors = errors.New("unequal named anchors not yet supported")
)

// MergeYAMLNodes merges the content of node `b` into node `a`.
// If `a` contains already an element with the same name and of the same kind as `b`,
// they are merged recursively.
func MergeYAMLNodes(nodeA, nodeB *yaml.Node) (*yaml.Node, error) {
	if nodeA == nil || nodeB == nil {
		return nil, ErrNodeNil
	}

	// an empty node (e.g. from an empty document) does not change the other side
	if nodeB.Kind == 0 {
		ret := *nodeA

		return &ret, nil
	}

	if nodeA.Kind == 0 {
		ret := *nodeB

		return &ret, nil
	}

	// aliases are resolved before comparing the kinds of both sides
	resolvedA := resolveAlias(nodeA)
	resolvedB := resolveAlias(nodeB)

	if resolvedA == nil || resolvedB == nil {
		return nil, fmt.Errorf("alias without target: %w", ErrNodeNil)
	}

	// null replaces and is replaced by any kind of node
	if isNullNode(resolvedA) || isNullNode(resolvedB) {
		return copyResolved(nodeB), nil
	}

	if resolvedA.Kind != resolvedB.Kind {
		return nil, ErrNodeKindMismatch
	}

	nodeB = resolvedB

	res, resErr := mergeByKind(nodeA, nodeB)

	if len(nodeA.Anchor) > 0 {
		if len(nodeB.Anchor) > 0 && nodeA.Anchor != nodeB.Anchor {
			res = nil
			resErr = fmt.Errorf("%w (source %v, merge %v)",
				ErrUnequalNameAnchors,
				nodeA.Anchor,
				nodeB.Anchor)
		}
	} else {
		nodeA.Anchor = nodeB.Anchor
	}

	return res, resErr
}

// mergeByKind dispatches the merge of nodeA and nodeB to the function handling the kind of nodeA.
func mergeByKind(nodeA, nodeB *yaml.Node) (*yaml.Node, error) {
	switch nodeA.Kind {
	case yaml.DocumentNode:
		return mergeDocumentNodes(nodeA, nodeB)
	case yaml.SequenceNode:
		return mergeSequenceNodes(nodeA, nodeB)
	case yaml.MappingNode:
		return mergeMappingNodes(nodeA, nodeB)
	case yaml.ScalarNode:
		return mergeScalarNodes(nodeA, nodeB)
	case yaml.AliasNode:
		return mergeAliasNodes(nodeA, nodeB)
	default:
		return nil, fmt.Errorf("unhandled node type %v: %w", nodeA.Kind, ErrNodeTypeUnhandled)
	}
}

// resolveAlias follows the alias chain of node and returns the first non-alias node.
func resolveAlias(node *yaml.Node) *yaml.Node {
	for node != nil && node.Kind == yaml.AliasNode {
		node = node.Alias
	}

	return node
}

// isNullNode reports whether node "is" or "is an alias to", a null scalar.
func isNullNode(node *yaml.Node) bool {
	node = resolveAlias(node)

	return node != nil && node.Kind == yaml.ScalarNode && node.ShortTag() == "!!null"
}

// copyResolved returns a copy of the node, resolving it first if it is an alias.
// Like in mergeAliasNodes, the anchor of a resolved alias target is not copied.
func copyResolved(node *yaml.Node) *yaml.Node {
	resolved := resolveAlias(node)

	if resolved == nil || resolved == node {
		ret := *node

		return &ret
	}

	ret := *resolved
	ret.Anchor = ""

	return &ret
}

func mergeAliasNodes(nodeA, nodeB *yaml.Node) (*yaml.Node, error) {
	if nodeA == nil || nodeB == nil {
		return nil, ErrNodeNil
	}

	if nodeA.Kind != yaml.AliasNode || nodeA.Alias == nil {
		return nil, ErrAliasNodeExpected
	}

	tmp := *nodeA.Alias
	tmp.Anchor = ""

	return MergeYAMLNodes(&tmp, nodeB)
}

func mergeDocumentNodes(nodeA, nodeB *yaml.Node) (*yaml.Node, error) {
	if nodeA == nil || nodeB == nil {
		return nil, ErrNodeNil
	}

	if nodeA.Kind != yaml.DocumentNode || nodeB.Kind != yaml.DocumentNode {
		return nil, ErrNodeKindMismatch
	}

	// this is the top level, the YAML library does not support multiple documents in one file
	if len(nodeA.Content) == 1 && len(nodeB.Content) == 1 {
		ret := *nodeA
		ret.Content = make([]*yaml.Node, 1)

		merged, mergedErr := MergeYAMLNodes(nodeA.Content[0], nodeB.Content[0])

		if mergedErr != nil {
			return nil, mergedErr
		}

		ret.Content[0] = merged

		return &ret, nil
	}

	return nil, ErrUnexpectedDocumentNodeConfiguration
}

func mergeScalarNodes(a, b *yaml.Node) (*yaml.Node, error) {
	if a == nil || b == nil {
		return nil, ErrNodeNil
	}

	if a.Kind != yaml.ScalarNode || b.Kind != yaml.ScalarNode {
		return nil, ErrNodeKindMismatch
	}

	ret := *b

	return &ret, nil
}

func mergeSequenceNodes(nodeA, nodeB *yaml.Node) (*yaml.Node, error) {
	if nodeA == nil || nodeB == nil {
		return nil, ErrNodeNil
	}

	if nodeA.Kind != yaml.SequenceNode || nodeB.Kind != yaml.SequenceNode {
		return nil, ErrNodeKindMismatch
	}

	ret := *nodeA
	ret.Content = make([]*yaml.Node, 0, len(nodeA.Content)+len(nodeB.Content))
	ret.Content = append(ret.Content, nodeA.Content...)
	ret.Content = append(ret.Content, nodeB.Content...)

	return &ret, nil
}

// rawScalarKey is the fallback representation of a scalar key that cannot be decoded
// into a comparable value.
type rawScalarKey struct {
	tag   string
	value string
}

// scalarKeyValue returns a comparable representation of a mapping key, resolving aliases.
// Two keys are equal if their decoded values are of the same type and value, e.g. `1` and `0x1`,
// but not `1` and `"1"`. The second return value is false for keys that are not scalars,
// as those cannot be decoded into Go types anyway.
func scalarKeyValue(node *yaml.Node) (any, bool) {
	node = resolveAlias(node)

	if node == nil || node.Kind != yaml.ScalarNode {
		return nil, false
	}

	var value any

	// the comparability check only guards against panics on ==, scalars decode to comparable types
	if err := node.Decode(&value); err != nil || (value != nil && !reflect.TypeOf(value).Comparable()) {
		return rawScalarKey{tag: node.ShortTag(), value: node.Value}, true
	}

	return value, true
}

func mergeAddValue(node, key, value *yaml.Node) error {
	var keyIndex int
	var valueIndex int

	keyValue, keyIsScalar := scalarKeyValue(key)

	for i := 0; keyIsScalar && i+1 < len(node.Content); i += 2 {
		keyIndex = i
		valueIndex = i + 1

		if nodeKeyValue, ok := scalarKeyValue(node.Content[keyIndex]); ok && nodeKeyValue == keyValue {
			merged, mergedErr := MergeYAMLNodes(node.Content[valueIndex], value)

			if mergedErr == nil {
				node.Content[valueIndex] = merged
			}

			return mergedErr
		}
	}

	node.Content = append(node.Content, key, value)

	return nil
}

func mergeMappingNodes(nodeA, nodeB *yaml.Node) (*yaml.Node, error) {
	if nodeA == nil || nodeB == nil {
		return nil, ErrNodeNil
	}

	if nodeA.Kind != yaml.MappingNode || nodeB.Kind != yaml.MappingNode {
		return nil, ErrNodeKindMismatch
	}

	var keyNode *yaml.Node
	var valueNode *yaml.Node

	ret := *nodeA
	ret.Content = make([]*yaml.Node, 0, len(nodeA.Content)+len(nodeB.Content))
	ret.Content = append(ret.Content, nodeA.Content...)

	for i := 0; i+1 < len(nodeB.Content); i += 2 {
		keyNode = nodeB.Content[i]
		valueNode = nodeB.Content[i+1]

		if mergeErr := mergeAddValue(&ret, keyNode, valueNode); mergeErr != nil {
			return nil, mergeErr
		}
	}

	return &ret, nil
}
