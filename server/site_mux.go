package main

import "regexp"

// We store site client lists by the host name only.
var siteBuckets = make(map[string]*ClientList)

var disallowedHosts = []string{
	"localhost",
}

var ipDetect = regexp.MustCompile(`[0-9]{1,3}\.[0-9]{1,3}\.[0-9]{1,3}\.[0-9]{1,3}`)
