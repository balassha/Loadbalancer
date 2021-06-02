package main

import (
	"fmt"
	"sixt/algorithm/randomAlgorithm"
	"sixt/algorithm/roundRobinAlgorithm"
	Config "sixt/configuration"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestReadConfig(t *testing.T) {
	cases := []struct {
		expected error
	}{
		{
			expected: nil,
		},
	}
	for _, c := range cases {
		have := ReadConfig()
		assert.Equal(t, c.expected, have)
	}
}

func TestSelectUpstream(t *testing.T) {
	cases := []struct {
		name      string
		algorithm Algorithm
		expected  bool
	}{
		{
			name:      "TestRandomSelection",
			algorithm: new(randomAlgorithm.Random),
			expected:  true,
		},
		{
			name:      "TestRoundRobinSelection",
			algorithm: new(roundRobinAlgorithm.RoundRobin),
			expected:  true,
		},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			var mutex sync.Mutex
			have := true
			req := make(chan Config.Request)
			upstreamService := []chan Config.Request{req, req, req}
			lb := &MyLoadBalancer{
				UpstreamService:    upstreamService,
				Current:            0,
				mu:                 &mutex,
				selectionAlgorithm: c.algorithm,
			}
			if lb.SelectUpstream(&lb.Current) == nil {
				have = false
			}
			errMsg := "The Upstream Selector didn't return the Request Channel"
			assert.Equal(t, c.expected, have, fmt.Errorf("%s has Failed with error : %v", c.name, errMsg))
		})
	}

}

func TestRandomSelection(t *testing.T) {
	cases := []struct {
		expected bool
	}{
		{
			expected: true,
		},
	}
	for _, c := range cases {
		var algo Algorithm
		var mutex sync.Mutex
		have := true
		algo = new(randomAlgorithm.Random)

		req1, req2, req3 := make(chan Config.Request), make(chan Config.Request), make(chan Config.Request)
		UpstreamService := []chan Config.Request{req1, req2, req3}
		lb := &MyLoadBalancer{
			UpstreamService: UpstreamService,
			Current:         0,
			mu:              &mutex,
		}

		countRequest := make(map[chan Config.Request]int)
		for i := 0; i < len(lb.UpstreamService); i++ {
			req := algo.GetNextItem(lb.UpstreamService, nil, lb.mu)
			countRequest[req]++
		}

		for _, v := range countRequest {
			if v == len(lb.UpstreamService) {
				have = false
			}
		}

		errMsg := "The Random Selector didn't return the Request Channel in a random manner"
		assert.Equal(t, c.expected, have, fmt.Errorf("TestRandomSelection has Failed with error : %v", errMsg))
	}
}

func TestRoundRobinSelection(t *testing.T) {
	cases := []struct {
		expected bool
	}{
		{
			expected: true,
		},
	}
	for _, c := range cases {
		have := true
		var algo Algorithm
		var mutex sync.Mutex
		algo = new(roundRobinAlgorithm.RoundRobin)

		req := make(chan Config.Request)
		UpstreamService := []chan Config.Request{req, req, req}
		lb := &MyLoadBalancer{
			UpstreamService: UpstreamService,
			Current:         0,
			mu:              &mutex,
		}

		for i := 0; i < len(lb.UpstreamService); i++ {
			req := algo.GetNextItem(lb.UpstreamService, &lb.Current, lb.mu)
			if req != lb.UpstreamService[i] {
				have = false
			}
		}

		errMsg := "The Round Robin Selector didn't return the Request Channel in a round robin order"
		assert.Equal(t, c.expected, have, fmt.Errorf("TestRoundRobinSelection has Failed with error : %v", errMsg))
	}
}

func TestLoadBalancer(t *testing.T) {
	cases := []struct {
		data     interface{}
		expected bool
	}{
		{
			data:     1,
			expected: true,
		},
		{
			data:     nil,
			expected: true,
		},
		{
			data:     1.0,
			expected: true,
		},
		{
			data:     "Test",
			expected: true,
		},
		{
			data:     []int{1, 2, 3},
			expected: true,
		},
		{
			data: struct {
				value int64
			}{value: 9223372036854775807},
			expected: true,
		},
	}
	for _, c := range cases {
		have := runLoadBalancer()

		errMsg := "The Response was not received in Time"
		assert.Equal(t, c.expected, have, fmt.Errorf("TestLoadBalancer has Failed with error : %v", errMsg))
	}
}

func BenchmarkLoadbalancer(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = runLoadBalancer()
	}
}

func runLoadBalancer() bool {
	manager := &TimeServiceManager{}
	ts := manager.Spawn()
	go ts.Run()
	var algo Algorithm
	var mutex sync.Mutex
	algo = new(roundRobinAlgorithm.RoundRobin)

	UpstreamService := make([]chan Config.Request, 0)
	lb := &MyLoadBalancer{
		UpstreamService:    UpstreamService,
		Current:            0,
		mu:                 &mutex,
		selectionAlgorithm: algo,
	}
	lb.RegisterInstance(ts.ReqChan)
	select {
	case <-lb.Request(nil):
	case <-time.After(5 * time.Second):
		return false
	}
	return true
}
