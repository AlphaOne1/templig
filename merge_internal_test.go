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
