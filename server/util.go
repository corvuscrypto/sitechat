package main

import "encoding/json"

func messageToJSON(username string, message []byte) []byte {
	data, _ := json.Marshal(map[string]interface{}{
		"username": username,
		"message":  string(message),
	})
	return data
}
