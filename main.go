package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math/rand"
	"os"
	"strings"
	"sync"
	"time"

	"sixt/algorithm/randomAlgorithm"
	"sixt/algorithm/roundRobinAlgorithm"
	Config "sixt/configuration"
)

//Interface for choosing Algorithm
type Algorithm interface {
	GetNextItem(items []chan Config.Request, current *int, mu *sync.Mutex) chan Config.Request
}

// LoadBalancer is used for balancing load between multiple instances of a service.
type LoadBalancer interface {
	Request(payload interface{}) chan Config.Response
	RegisterInstance(chan Config.Request)
	RemoveInstance(index int)
}

// MyLoadBalancer is the load balancer you should modify!
type MyLoadBalancer struct {
	mu                 *sync.Mutex
	UpstreamService    []chan Config.Request
	Current            int
	selectionAlgorithm Algorithm
}

// RegisterInstance is currently a dummy implementation. Please implement it!
func (lb *MyLoadBalancer) RegisterInstance(ch chan Config.Request) {
	if ch != nil {
		lb.AddUpstreamService(&ch)
	}
}

// Request is currently a dummy implementation. Please implement it!
func (lb *MyLoadBalancer) Request(payload interface{}) chan Config.Response {
	responseToClient := make(chan Config.Response, 1)
	lb.mu.Lock()
	defer lb.mu.Unlock()
	if len(lb.UpstreamService) == 0 {
		responseToClient <- "Not Time Servers are Registered with Load Balancer."
		return responseToClient
	}

	go lb.Spawn(&responseToClient, payload)

	return responseToClient
}

func (lb *MyLoadBalancer) SelectUpstream(current *int) chan Config.Request {
	return lb.selectionAlgorithm.GetNextItem(lb.UpstreamService, current, lb.mu)
}

func (lb *MyLoadBalancer) Spawn(resp *chan Config.Response, payload interface{}) {
	retryCount := 3
	lb.mu.Lock()
	current := lb.Current
	lb.mu.Unlock()

	downStreamResponse := make(chan Config.Response, 1)
	reqBody := Config.Request{
		Payload: payload,
		RspChan: downStreamResponse,
	}

	var req chan Config.Request
	// Run a loop to retry with next available upstream service if the
	// current Upstream service is busy
	for i := 0; i < retryCount; i++ {
		requestProcessed := false
		//Choose Upstream Service
		req := lb.SelectUpstream(&current)
		select {
		case req <- reqBody:
			Config.PrintIfDebug("Message Sent from Goroutine to Upstream service")
			requestProcessed = true
		case <-time.After(5 * time.Second):
			Config.PrintIfDebug("no message sent from Goroutine to Upstream service")
		}
		if requestProcessed {
			break
		}
	}

	select {
	case rsp := <-downStreamResponse:
		*resp <- rsp
		Config.PrintIfDebug(rsp)
	case <-time.After(5 * time.Second):
		lb.RemoveInstance(current)
		go lb.RecoverTimedOutInstance(&req, resp, &downStreamResponse)
	}
}

func (lb *MyLoadBalancer) RecoverTimedOutInstance(reqChan *chan Config.Request, respChan *chan Config.Response, downStreamResponse *chan Config.Response) {
	//Retry thrice to Recover a TimedOut Instance
	// If not recovered after 3 times, remove instanc
	retryCount := 3
	for i := 0; i < retryCount; i++ {
		select {
		case val := <-*downStreamResponse:
			lb.AddUpstreamService(reqChan)
			*respChan <- val
			return
		case <-time.After(5 * time.Second):
			continue
		}
	}
	*respChan <- Config.ServiceUnavailable
}

func (lb *MyLoadBalancer) AddUpstreamService(reqChan *chan Config.Request) {
	lb.mu.Lock()
	lb.UpstreamService = append(lb.UpstreamService, *reqChan)
	lb.mu.Unlock()
}

func (lb *MyLoadBalancer) RemoveInstance(index int) {
	lb.mu.Lock()
	defer lb.mu.Unlock()
	length := len(lb.UpstreamService)
	if length == 0 {
		return
	}
	if index == -1 {
		index = rand.Intn(len(lb.UpstreamService))
	}
	lb.UpstreamService = append(lb.UpstreamService[:index], lb.UpstreamService[index+1:]...)
}

/******************* Sample Implementation of Weighted Round Robin Algorithm ***********/

// func GetNextUpstream(lb *MyLoadBalancer) chan Request {

// 	// Creating a Upstream Cluster cache with Weight.
// 	UpstreamService := make(map[int]map[string]interface{})
// 	req1 := make(map[string]interface{})
// 	req1["request"] = make(chan Request)
// 	req1["weight"] = 2
// 	UpstreamService[0] = req1
// 	req2 := make(map[string]interface{})
// 	req2["request"] = make(chan Request)
// 	req2["weight"] = 4
// 	UpstreamService[0] = req2

// 	var i int = -1 // indicates the last selected server
// 	var cw int = 0 // indicates the weight of the current schedule
// 	var gcd int = 2
// 	for {
// 		i = (i + 1) % len(lb.UpstreamService)
// 		if i == 0 {
// 			cw = cw - gcd
// 			if cw <= 0 {
// 				cw = getMaxWeight(lb)
// 				if cw == 0 {
// 					return nil
// 				}
// 			}
// 		}

// 		if weight, _ := lb.UpstreamService[i]["weight"].(int); weight >= cw {
// 			return lb.UpstreamService[i]["request"].(chan Request)
// 		}
// 	}
// }

// func getMaxWeight(lb *MyLoadBalancer) int {
// 	max := 0
// 	for _, v := range lb.UpstreamService {
// 		if weight, _ := v["weight"]; weight.(int) >= max {
// 			max = weight.(int)
// 		}
// 	}
// 	return max
// }

/*************************************************************************************/

/******************************************************************************
 *  STANDARD TIME SERVICE IMPLEMENTATION -- MODIFY IF YOU LIKE                *
 ******************************************************************************/

// TimeService is a single instance of a time service.
type TimeService struct {
	Dead            chan struct{}
	ReqChan         chan Config.Request
	AvgResponseTime float64
}

// Run will make the TimeService start listening to the two channels Dead and ReqChan.
func (ts *TimeService) Run() {
	for {
		select {
		case <-ts.Dead:
			return
		case req := <-ts.ReqChan:
			processingTime := time.Duration(ts.AvgResponseTime+1.0-rand.Float64()) * time.Second
			time.Sleep(processingTime)
			req.RspChan <- time.Now()
		}
	}
}

func ReadConfig() error {
	//Read LoadBalancer Config from LoadBalancerConfig.json
	jsonFile, err := os.Open(Config.LoadBalancerConfigFile)
	if err != nil {
		errMsg := `unable to find Load Balancer Configuration. 
					   Setting default value for Selection Algorithm : Round Robin`
		return fmt.Errorf(errMsg)
	}
	defer jsonFile.Close()

	// read our opened jsonFile as a byte array.
	byteValue, err := ioutil.ReadAll(jsonFile)
	if err != nil {
		errMsg := `unable to Read Load Balancer Configuration. 
					   Setting default value for Selection Algorithm : Round Robin`
		return fmt.Errorf(errMsg)
	}

	var config Config.LoadBalancerConfig

	err = json.Unmarshal(byteValue, &config)
	if err != nil {
		errMsg := `unable to Marshal Load Balancer Configuration. 
					   Setting default value for Selection Algorithm : Round Robin`
		return fmt.Errorf(errMsg)
	}

	// Set Load balancer configuration
	Config.SetAlgorithmFromConfig(config.Algorithm)
	Config.SetLoglevelFromConfig(config.Logging)

	return nil
}

func GetSelectionAlgorithm() Algorithm {
	switch Config.Algorithm {
	case Config.Random:
		return new(randomAlgorithm.Random)
	default:
		return new(roundRobinAlgorithm.RoundRobin)
	}
}

/******************************************************************************
 *  CLI -- YOU SHOULD NOT NEED TO MODIFY ANYTHING BELOW                       *
 ******************************************************************************/

// main runs an interactive console for spawning, killing and asking for the
// time.
func main() {
	rand.Seed(int64(time.Now().Nanosecond()))

	// Reading Algorithm type from Configuration
	if err := ReadConfig(); err != nil {
		Config.PrintIfDebug(err)
	}

	bio := bufio.NewReader(os.Stdin)
	var mu sync.Mutex
	var lb LoadBalancer = &MyLoadBalancer{
		mu:                 &mu,
		selectionAlgorithm: GetSelectionAlgorithm(),
	}

	manager := &TimeServiceManager{}
	stopServer := false
	for {
		if stopServer {
			break
		}
		fmt.Printf("> ")
		cmd, err := bio.ReadString('\n')
		if err != nil {
			fmt.Println("Error reading command: ", err)
			continue
		}
		switch strings.TrimSpace(cmd) {
		case "kill":
			manager.Kill()
			lb.RemoveInstance(-1)
		case "spawn":
			ts := manager.Spawn()
			lb.RegisterInstance(ts.ReqChan)
			go ts.Run()
		case "time":
			select {
			case rsp := <-lb.Request(nil):
				fmt.Println(rsp)
			case <-time.After(5 * time.Second):
				fmt.Println("Timeout")
			}
		case "stop":
			stopServer = true
			fmt.Println("Server Stopped")
		default:
			fmt.Printf("Unknown command: %s Available commands: time, spawn, kill, stop\n", cmd)
		}
	}
}

// TimeServiceManager is responsible for spawning and killing.
type TimeServiceManager struct {
	Instances []TimeService
}

// Kill makes a random TimeService instance unresponsive.
func (m *TimeServiceManager) Kill() {
	if len(m.Instances) > 0 {
		n := rand.Intn(len(m.Instances))
		close(m.Instances[n].Dead)
		m.Instances = append(m.Instances[:n], m.Instances[n+1:]...)
	}
}

// Spawn creates a new TimeService instance.
func (m *TimeServiceManager) Spawn() TimeService {
	ts := TimeService{
		Dead:            make(chan struct{}, 0),
		ReqChan:         make(chan Config.Request, 10),
		AvgResponseTime: rand.Float64() * 3,
	}
	m.Instances = append(m.Instances, ts)

	return ts
}
