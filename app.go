package main

type (
	config struct {
		adminName string
		adminPass string
	}
)

var e = createMux()
var cfg = loadConfig()