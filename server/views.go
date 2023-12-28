package server

import "encoding/json"

func health() []byte {
	// marshal
	marshaledRes, _ := json.Marshal(map[string]interface{}{
		"status": "ok",
	})
	return marshaledRes
}
