package schedular

import(
	"sync/atomic"
)

// Workernodes represents a list of available workers
// in real K8s, this would be discovered or register dynamically
// just for now I have to hardcoded this later I have upgrade it into dynamic change
var WorkerNodes = []string{
	"http://localhost:8082",
	"http://localhost:8083",
	"http://localhost:8084",
}

var currentIndex uint64 = 0

// GetNextWorker returns the next worker URL using a thread-safe round-robin approach.
func GetNextWorker() string {
	if len(WorkerNodes) == 0 {
		return ""
	}

	// Automatically increment the index to prevent race conditions during multiple requests
	val := atomic.AddUint64(&currentIndex, 1)

	// calculate the index using modulo operator against the number of workers
	index := (val - 1) % uint64(len(WorkerNodes))

	return WorkerNodes[index]
}