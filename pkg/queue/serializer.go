package queue

import (
	"encoding/json"
	"strings"

	"github.com/elliotchance/phpserialize"
)

// UnserializeCommand attempts to parse the PHP serialized command from the job payload
func UnserializeCommand(data json.RawMessage) (any, error) {
	// First, try to unmarshal data as a map to find "command"
	var dataMap map[string]interface{}
	if err := json.Unmarshal(data, &dataMap); err != nil {
		return nil, err
	}

	commandStr, ok := dataMap["command"].(string)
	if !ok {
		// Not a standard Laravel serialized command job
		return dataMap, nil
	}

	var out interface{}
	err := phpserialize.Unmarshal([]byte(commandStr), &out)
	return out, err
}

// Helper to extract a private property from a PHP object map if needed
// PHP serialized objects often have keys like "\x00*\x00propName" or "\x00ClassName\x00propName"
func GetPHPProperty(obj any, propName string) any {
	m, ok := obj.(map[interface{}]interface{})
	if !ok {
		return nil
	}

	// Try direct match
	if val, ok := m[propName]; ok {
		return val
	}

	// Try protected/private match (* for protected, ClassName for private)
	// We iterate because constructing the exact key with null bytes is tricky in Go string literals
	for k, v := range m {
		ks, ok := k.(string)
		if !ok {
			continue
		}
		if strings.HasSuffix(ks, "\x00"+propName) {
			return v
		}
	}
	return nil
}
