package templig

import (
	"errors"
	"fmt"

	"gopkg.in/yaml.v3"
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

	if nodeA.Kind != nodeB.Kind && nodeA.Kind != yaml.AliasNode && nodeB.Kind != yaml.AliasNode {
		return nil, ErrNodeKindMismatch
	}

	for nodeB.Kind == yaml.AliasNode {
		nodeB = nodeB.Alias
	}

	var res *yaml.Node
	var resErr error

	switch nodeA.Kind {
	case yaml.DocumentNode:
		res, resErr = mergeDocumentNodes(nodeA, nodeB)
	case yaml.SequenceNode:
		res, resErr = mergeSequenceNodes(nodeA, nodeB)
	case yaml.MappingNode:
		res, resErr = mergeMappingNodes(nodeA, nodeB)
	case yaml.ScalarNode:
		res, resErr = mergeScalarNodes(nodeA, nodeB)
	case yaml.AliasNode:
		res, resErr = mergeAliasNodes(nodeA, nodeB)
	default:
		resErr = fmt.Errorf("unhandled node type %v: %w", nodeA.Kind, ErrNodeTypeUnhandled)
	}

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

func mergeAddValue(node, key, value *yaml.Node) error {
	var keyIndex int
	var valueIndex int

	for i := 0; i+1 < len(node.Content); i += 2 {
		keyIndex = i
		valueIndex = i + 1

		if node.Content[keyIndex].Kind == yaml.ScalarNode &&
			key.Kind == yaml.ScalarNode &&
			node.Content[keyIndex].Value == key.Value {

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
