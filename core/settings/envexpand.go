package settings

import (
	"bytes"
	"os"
	"regexp"

	"go.yaml.in/yaml/v4"
)

var envPlaceholder = regexp.MustCompile(`\$\{([A-Za-z_][A-Za-z0-9_]*)\}`)

// expandEnv replaces every ${NAME} in the file's values with that environment
// variable, empty when unset, so a config kept in a public repo can carry
// placeholders instead of personal values. It works on parsed values, never
// on the raw text, so a multi-line value cannot break the YAML, and a bare
// $NAME, the form shell commands use, is left alone.
func expandEnv(raw []byte) ([]byte, bool, error) {
	if !envPlaceholder.Match(raw) {
		return raw, false, nil
	}
	var root yaml.Node
	if err := yaml.Unmarshal(raw, &root); err != nil {
		return nil, false, err
	}
	expandNode(&root)
	var out bytes.Buffer
	enc := yaml.NewEncoder(&out)
	if err := enc.Encode(&root); err != nil {
		return nil, false, err
	}
	return out.Bytes(), true, enc.Close()
}

func expandNode(n *yaml.Node) {
	if n.Kind == yaml.ScalarNode {
		n.Value = envPlaceholder.ReplaceAllStringFunc(n.Value, func(token string) string {
			return os.Getenv(envPlaceholder.FindStringSubmatch(token)[1])
		})
		return
	}
	for _, child := range n.Content {
		expandNode(child)
	}
}
