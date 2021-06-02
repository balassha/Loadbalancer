package configuration

import "fmt"

var (
	LogLevel  = "error"
	isDebug   = false
	Algorithm = "round-robin"
)

const (
	LoadBalancerConfigFile = "LoadBalancerConfig.json"
	Random                 = "random"
	RoundRobin             = "round-robin"
	ServiceUnavailable     = "503"
	Debug                  = "debug"
	Error                  = "error"
)

// LoadBalancerConfig is used to read configuration parameters of Load Balancer
type LoadBalancerConfig struct {
	Algorithm string `json:"algorithm"`
	Logging   string `json:"logging"`
}

// Request is used for making requests to services behind a load balancer.
type Request struct {
	Payload interface{}
	RspChan chan Response
}

// Response is the value returned by services behind a load balancer.
type Response interface{}

func IsDebug() bool {
	return isDebug
}

func SetDebug(debug bool) {
	isDebug = debug
}

func GetLogLevel() string {
	return LogLevel
}

func GetAlgorithm() string {
	return Algorithm
}

func PrintIfDebug(msg interface{}) {
	if isDebug {
		fmt.Println(msg)
	}
}

func SetAlgorithmFromConfig(algorithm string) {
	if algorithm != "" {
		switch algorithm {
		case Random:
			Algorithm = Random
		default:
			Algorithm = RoundRobin
		}
	}
}

func SetLoglevelFromConfig(logLevel string) {
	if logLevel != "" {
		switch logLevel {
		case Debug:
			SetDebug(true)
			LogLevel = Debug
		default:
			SetDebug(false)
			LogLevel = Error
		}

	}
}
