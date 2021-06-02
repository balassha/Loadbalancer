package randomAlgorithm

import (
	"math/rand"
	"sync"

	Config "sixt/configuration"
)

type Random struct{}

func (r *Random) GetNextItem(items []chan Config.Request, current *int, mu *sync.Mutex) chan Config.Request {
	mu.Lock()
	defer mu.Unlock()
	length := len(items)
	if length == 0 {
		return nil
	}
	index := rand.Intn(length)
	return items[index]
}
