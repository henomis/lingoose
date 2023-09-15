package types

import "encoding/json"

type M map[string]interface{}

type Meta map[string]interface{}

// String returns the metadata as a JSON string
func (m Meta) String() string {
	jsonData, err := json.Marshal(m)
	if err != nil {
		return ""
	}

	return string(jsonData)
}
