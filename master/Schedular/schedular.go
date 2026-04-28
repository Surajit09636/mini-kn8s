package schedular

import (
	"log"
	"sync"
	"sync/atomic"
	"time"
	"bytes"
	"encoding/json"
	"net/http"
)

type WorkerNode struct {
	URL      string
	LastPing time.Time
	Tasks    []string 
}

var (
	WorkerNodes  []WorkerNode
	currentIndex uint64 = 0
	mu           sync.RWMutex
)

// RegisterWorker is used by the master API to dynamically add a node or update its heartbeat
func RegisterWorker(workerURL string) {
	mu.Lock()
	defer mu.Unlock()

	for i, node := range WorkerNodes {
		if node.URL == workerURL {
			WorkerNodes[i].LastPing = time.Now()
			return
		}
	}

	WorkerNodes = append(WorkerNodes, WorkerNode{URL: workerURL, LastPing: time.Now()})
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

	return WorkerNodes[index].URL
}

func MonitorHealth() {
	for {
		time.Sleep(10 * time.Second) // Check every 10 seconds
		mu.Lock()
		now := time.Now()
		var healthyNodes []WorkerNode
		var deadTasks []string // To collect tasks that need rescheduling
		
		for _, node := range WorkerNodes {
			if now.Sub(node.LastPing) <= 30*time.Second {
				healthyNodes = append(healthyNodes, node)
			} else {
				log.Printf("[Schedular] Worker %s timed out. Removing from active pool.", node.URL)
				deadTasks = append(deadTasks, node.Tasks...) // Salvage the tasks!
			}
		}

		if len(WorkerNodes) != len(healthyNodes) {
			WorkerNodes = healthyNodes
			log.Printf("[Schedular] Active pool size updated: %d", len(WorkerNodes))
		}
		mu.Unlock() // Important: Unlock BEFORE making network calls

		// Reschedule any salvaged tasks outside the lock to prevent deadlocks
		for _, image := range deadTasks {
			go RescheduleTask(image)
		}
	}
}


//Assign Task records what container is running on which worker
func AssignTask(workerURL string, image string) {
	mu.Lock()
	defer mu.Unlock()
	for i, node := range WorkerNodes {
		if node.URL == workerURL {
			WorkerNodes[i].Tasks = append(WorkerNodes[i].Tasks, image)
			log.Printf("[Schedular] Assigned task %s to worker %s", image, workerURL)
			return
		}
	}
}

func RemoveTask(workerURL string, image string) {
	mu.Lock()
	defer mu.Unlock()

	for i, node := range WorkerNodes {
		if node.URL != workerURL {
			continue
		}

		for j, task := range node.Tasks {
			if task == image {
				WorkerNodes[i].Tasks = append(node.Tasks[:j], node.Tasks[j+1:]...)
				log.Printf("[Schedular] Removed task %s from worker %s", image, workerURL)
				return
			}
		}
	}
}

//RescheduleTask finds a new healthy node and pushes thetask to it
func RescheduleTask(image string) {
	workerURL := GetNextWorker()
	if workerURL == "" {
		log.Printf("[Schedular] cannot reschedule %s, no healthy workers available!", image)
		return
	}

	log.Printf("[Schedular] Rescheduling %s to worer: %s", image, workerURL)

	payload := map[string]string{"image": image}

	jsonData, _ := json.Marshal(payload)

	resp, err := http.Post(workerURL+"/run", "application/json", bytes.NewBuffer(jsonData))

	if err != nil || resp.StatusCode != http.StatusOK {
		log.Printf("[Schedular] Failed to reschedule %s on %s.", image, workerURL)
		if resp != nil {
			resp.Body.Close()
		}
		return
	}
	resp.Body.Close()

	AssignTask(workerURL, image)
	log.Printf("[Schedular] Successfully self-healed task %s -> %s", image, workerURL)
}
