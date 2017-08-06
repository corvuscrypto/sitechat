package main

import "encoding/json"

type event struct {
	Type string          `json:"type"`
	Data json.RawMessage `json:"data"`
}

type locationUpdate struct {
	Host string `json:"host"`
	Path string `json:"path"`
}

type messageEvent struct {
	Username string
	Message  string `json:"message"`
}
