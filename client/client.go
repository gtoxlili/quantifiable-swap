package client

import (
	"net/http"
	"net/url"
	"quantifiable-swap/constants"
	"time"
)

var C = &http.Client{
	Timeout: 10 * time.Second,
}

func init() {
	if constants.ProxyAddr != "" {
		proxyUrl, _ := url.Parse(constants.ProxyAddr)
		C.Transport = &http.Transport{
			Proxy: http.ProxyURL(proxyUrl),
		}
	}
}
