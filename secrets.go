// SPDX-FileCopyrightText: 2026 The templig contributors.
// SPDX-License-Identifier: MPL-2.0

package templig

import (
	"fmt"
	"regexp"
	"strings"

	"go.yaml.in/yaml/v4"
)

// SecretDefaultRE is the default regular expression used to identify secret values automatically.
const SecretDefaultRE = "(?i)key|secret|pass(?:word)?|cert(?:ificate)?"

// SecretRE is the regular expression used to identify secret values automatically.
// In case there are different properties to identify secrets, extend it. Access to this variable
// is not synchronized, thus modifying it shall be done before working with templig.
var SecretRE = regexp.MustCompile(SecretDefaultRE)

type secretWorkItem struct {
	node   *yaml.Node
	secret bool
}

// HideSecrets hides secrets in the given YAML node structure.
// Secrets are identified using the given `secretRE` parameter. If that parameter is nil, [SecretDefaultRE] is used
// instead to construct a new regexp to prevent silent complete failures to hide secrets.
// Depending on the parameter `hideStructure`, the structure of the secret is hidden too (`true`) or visible (`false`).
func HideSecrets(node *yaml.Node, hideStructure bool, secretRE *regexp.Regexp) {
	// initialWorkQueueDepth is an assumption about the maximum depth of the YAML document structure. It will not limit
	// the real depth, but should, for average use cases of templig, be enough.
	const initialWorkQueueDepth = 20
	const queueCompactionFrequency = 100

	if secretRE == nil {
		// we create a new regexp, as we cannot enforce, that nobody changed the SecretRE. This is unlike other
		// functions, like Config.SetSecretRE, that have errors indicating a wrongful usage.
		secretRE = regexp.MustCompile(SecretDefaultRE)
	}

	workQueue := make([]secretWorkItem, 1, initialWorkQueueDepth)
	workQueue[0] = secretWorkItem{
		node:   node,
		secret: false,
	}

	var newWork []secretWorkItem
	var currentWork secretWorkItem

	for i := 1; len(workQueue) > 0; i++ {
		if i%queueCompactionFrequency == 0 {
			// forces to allocate a new base array, really disposing of its used head part.
			workQueue = append([]secretWorkItem(nil), workQueue...)
		}

		currentWork = workQueue[0]

		if currentWork.secret {
			newWork = hideAll(currentWork.node, hideStructure)
		} else {
			newWork = hideSecrets(currentWork.node, secretRE)
		}

		if len(newWork) > 0 {
			workQueue = append(workQueue[1:], newWork...)
		} else {
			workQueue = workQueue[1:]
		}
	}
}

// hideSecrets identifies and processes secret values in a YAML node structure based on a predefined regular expression.
// It returns a slice of secretWorkItem, which contains nodes and a flag indicating whether they are secrets or not.
// Mapping nodes are handled differently due to their structured key-value pair content.
// Non-mapping nodes' content is scanned and added to the result with the secret flag set to false.
// Returns nil if the provided YAML node is nil.
func hideSecrets(node *yaml.Node, secretRE *regexp.Regexp) []secretWorkItem {
	if node == nil {
		return nil
	}

	var result []secretWorkItem

	// just mapping nodes need special handling in this step
	if node.Kind == yaml.MappingNode {
		result = make([]secretWorkItem, 0, len(node.Content)/2)

		// The content is a sequence of key-value pairs, thus the content length is even.
		// Subtracting 1 in the check accounts for potential uneven (invalid) content length.
		for i := 0; i+1 < len(node.Content); i += 2 {
			if secretRE.MatchString(node.Content[i].Value) {
				result = append(result, secretWorkItem{
					node:   node.Content[i+1],
					secret: true,
				})
			} else {
				result = append(result, secretWorkItem{
					node:   node.Content[i+1],
					secret: false,
				})
			}
		}
	} else {
		result = make([]secretWorkItem, 0, len(node.Content))

		for _, v := range node.Content {
			result = append(result, secretWorkItem{
				node:   v,
				secret: false,
			})
		}
	}

	return result
}

// hideAll processes a YAML node to mask sensitive data. Hides structure based on `hideStructure` when true.
func hideAll(node *yaml.Node, hideStructure bool) []secretWorkItem {
	const secretLengthCutoff = 32

	switch node.Kind {
	case yaml.ScalarNode:
		node.Tag = "!!str"
		if len(node.Value) < secretLengthCutoff {
			node.Value = strings.Repeat("*", len(node.Value))
		} else {
			node.Value = fmt.Sprintf("**%d**", len(node.Value))
		}
	case yaml.AliasNode:
		if node.Alias != nil {
			return []secretWorkItem{{
				node:   node.Alias,
				secret: true,
			}}
		}
	default:
		if hideStructure {
			node.Kind = yaml.ScalarNode
			node.Tag = "!!str"
			node.Value = "*"
			node.Content = nil
		} else {
			result := make([]secretWorkItem, 0, len(node.Content))

			for _, v := range node.Content {
				result = append(result, secretWorkItem{
					node:   v,
					secret: true,
				})
			}

			return result
		}
	}

	return nil
}
