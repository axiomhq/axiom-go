package config

import "net/url"

const apiURLStr = "https://api.axiom.co"

var apiURL *url.URL

func init() {
	var err error
	apiURL, err = url.ParseRequestURI(apiURLStr)
	if err != nil {
		panic(err)
	}
}

// APIURL is the api url of the hosted version of Axiom.
func APIURL() *url.URL {
	return apiURL
}
