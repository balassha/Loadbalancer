package roundRobinAlgorithm

import (
	"sync"

	Config "sixt/configuration"
)

type RoundRobin struct{}

func (r *RoundRobin) GetNextItem(items []chan Config.Request, current *int, mu *sync.Mutex) chan Config.Request {
	mu.Lock()
	defer mu.Unlock()
	item := items[*current]
	*current++
	// Check if the end of slice is reached
	// and we need to reset the counter and start
	// from the beginning
	if *current >= len(items) {
		*current = 0
	}
	return item
}
