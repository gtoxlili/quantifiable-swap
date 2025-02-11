package client

import (
	"github.com/gtoxlili/quantifiable-swap/constants"
	"net/http"
	"net/url"
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
