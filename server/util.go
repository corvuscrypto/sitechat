package main

import "math/rand"

func isAllowedHost(host string) bool {
	if ipDetect.MatchString(host) {
		return false
	}
	for _, v := range disallowedHosts {
		if v == host {
			return false
		}
	}
	return true
}

const numAdjectives = 3160
const numNouns = 12957

var nameMap = make(map[string]bool)

func generateUsername() string {
	username := adjectives[rand.Intn(numAdjectives)] + " " + nouns[rand.Intn(numNouns)]
	_, ok := nameMap[username]
	for ok {
		username = adjectives[rand.Intn(numAdjectives)] + " " + nouns[rand.Intn(numNouns)]
		_, ok = nameMap[username]
	}
	nameMap[username] = true
	return username
}
