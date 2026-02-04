package types

import (
	"bytes"
	"encoding/json"
	"strconv"
)

// unmarshalDynamicJSON handles fields that can be either a JSON array or a JSON map.
// If it's an array, it converts it into a map using the index as the key.
// If it's already a map, it just unmarshals it normally.
func unmarshalDynamicJSON[V any](data []byte) (map[string]V, error) {

	var m = make(map[string]V)
	var trimmedData = bytes.TrimSpace(data)

	if len(trimmedData) > 0 && trimmedData[0] == '[' {

		var slice []V
		if err := json.Unmarshal(data, &slice); err != nil {
			return nil, err
		}

		for i, element := range slice {
			m[strconv.Itoa(i)] = element
		}

	} else if err := json.Unmarshal(data, &m); err != nil {
		return nil, err
	}

	return m, nil

}
