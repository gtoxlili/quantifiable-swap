package client

import (
	"github.com/gtoxlili/quantifiable-swap/constants"
	"net/http"
	"net/url"
	"time"
)

var (
	C = &http.Client{
		Timeout: 10 * time.Second,
	}
	CNoTimeout = &http.Client{}
)

func init() {
	if constants.ProxyAddr != "" {
		proxyUrl, _ := url.Parse(constants.ProxyAddr)
		C.Transport = &http.Transport{
			Proxy: http.ProxyURL(proxyUrl),
		}
		CNoTimeout.Transport = &http.Transport{
			Proxy: http.ProxyURL(proxyUrl),
		}
	}
}
