package rssw

import (
	"net"
	"net/http"
	"time"
)

func getTimeoutHttpClient(timeout int) *http.Client {
	c := http.Client{
		Transport: &http.Transport{
			Dial: func(netw, addr string) (net.Conn, error) {
				deadline := time.Now().Add(time.Duration(timeout) * time.Second)
				c, err := net.DialTimeout(netw, addr, time.Duration(timeout)*time.Second)
				c.SetDeadline(deadline)
				return c, err
			},
		},
	}

	return &c
}

func httpGet(timeout int, link string) (*http.Response, error) {
	return getTimeoutHttpClient(timeout).Get(link)
}
