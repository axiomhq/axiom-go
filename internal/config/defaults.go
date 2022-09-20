package config

import "net/url"

const cloudURLStr = "https://cloud.axiom.co"

var cloudURL *url.URL

func init() {
	var err error
	cloudURL, err = url.ParseRequestURI(cloudURLStr)
	if err != nil {
		panic(err)
	}
}

// CloudURL is the url of the cloud hosted version of Axiom.
func CloudURL() *url.URL {
	return cloudURL
}
