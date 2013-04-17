package rssw

import (
	"net"
	"net/http"
	"time"
)

const (
	maxConnections = 10
)

var semaphores chan int

func initializeSemaphores() {
	semaphores = make(chan int, maxConnections)
	for i := 0; i < maxConnections; i++ {
		semaphores <- 0
	}
}

func getTimeoutHttpClient(timeout int) *http.Client {
	c := http.Client{
		Transport: &http.Transport{
			Dial: func(netw, addr string) (net.Conn, error) {
				deadline := time.Now().Add(time.Duration(timeout) * time.Second)
				c, err := net.DialTimeout(netw, addr, time.Duration(timeout)*time.Second)
				if c != nil {
					c.SetDeadline(deadline)
				}
				return c, err
			},
		},
	}

	return &c
}

func httpGet(timeout int, link string) (*http.Response, error) {
	if semaphores == nil {
		initializeSemaphores()
	}
	<-semaphores
	resp, err := getTimeoutHttpClient(timeout).Get(link)
	semaphores <- 0
	return resp, err
}
