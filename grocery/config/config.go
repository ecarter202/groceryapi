package config

import "fmt"

var (
	APIHOST = "192.41.48.147"
	APIPORT = 8081
	APIURL  = fmt.Sprintf("https://%s:%d", APIHOST, APIPORT)

	DEVAPIURL = fmt.Sprintf("http://127.0.0.1:%d", APIPORT)
)
