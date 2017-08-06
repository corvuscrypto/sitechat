package main

import (
	"net/http"
	"strings"
)

func connectHandler(w http.ResponseWriter, r *http.Request) {
	host := strings.ToLower(r.URL.Query().Get("host"))
	if host == "" || !isAllowedHost(host) {
		return
	}
	client := NewClient(w, r)
	list, ok := siteBuckets[host]
	if !ok {
		list = NewClientList()
		siteBuckets[host] = list
	}
	list.AddClient(client)
}

func main() {
	http.HandleFunc("/connect", connectHandler)
	http.ListenAndServe(":80", nil)
}
