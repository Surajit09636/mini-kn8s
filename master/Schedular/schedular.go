package schedular

import (
	"log"
	"sync"
	"sync/atomic"
)

var (
	WorkerNodes  []string
	currentIndex uint64 = 0
	mu           sync.RWMutex
)

// RegisterWorker is used by the master API to dynamically add a node
func RegisterWorker(workerURL string) {
	mu.Lock()
	defer mu.Unlock()

	for _, node := range WorkerNodes {
		if node == workerURL {
			return
		}
	}

	WorkerNodes = append(WorkerNodes, workerURL)
	log.Printf("[Schedular] New worker Registered successfully! Active pool size: %d", len(WorkerNodes))
}

func GetNextWorker() string {
	mu.RLock()
	defer mu.RUnlock()

	if len(WorkerNodes) == 0 {
		return ""
	}

	val := atomic.AddUint64(&currentIndex, 1)
	index := (val - 1) % uint64(len(WorkerNodes))

	return WorkerNodes[index]
}
