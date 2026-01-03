package queue

import (
	"encoding/json"
	"strings"

	"github.com/yvasiyarov/php_session_decoder/php_serialize"
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

	return php_serialize.UnSerialize(commandStr)
}

// Helper to extract a property from a PHP object (public, protected, or private)
func GetPHPProperty(obj any, propName string) any {
	// Check if it's a PhpObject (pointer or struct)
	if phpObj, ok := obj.(*php_serialize.PhpObject); ok {
		// Try public first
		if v, ok := phpObj.GetPublic(propName); ok {
			return v
		}
		// Try protected
		if v, ok := phpObj.GetProtected(propName); ok {
			return v
		}
		// Try private (requires class context usually, but checking implementation might support just name if simple)
		// Or we can check Members directly if exposed via GetMembers()
		// GetPrivate likely needs simple name.
		if v, ok := phpObj.GetPrivate(propName); ok {
			return v
		}

		// If not found via helper methods, iterate members manually?
		// GetMembers returns PhpArray (map[PhpValue]PhpValue usually)
		members := phpObj.GetMembers()
		for k, v := range members {
			kStr, ok := k.(string)
			if !ok {
				continue
			}
			// Protected: \0*\0propName
			// Private: \0ClassName\0propName
			if strings.Contains(kStr, "\x00"+propName) {
				return v
			}
			if kStr == propName {
				return v
			}
		}
	}

	// Fallback if it's just a map (array)
	if m, ok := obj.(map[interface{}]interface{}); ok {
		if val, ok := m[propName]; ok {
			return val
		}
	}

	// Fallback for PhpArray type alias if existing
	if m, ok := obj.(php_serialize.PhpArray); ok {
		if val, ok := m[propName]; ok {
			return val
		}
	}

	return nil
}
