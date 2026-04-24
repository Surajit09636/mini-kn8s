package models

import "time"

type Pod struct {
	ID           uint      `gorm:"primaryKey" json:"id"`
	DeploymentID uint      `json:"deployment_id"`
	WorkerURL    string    `json:"worker_url"`
	ContainerID  string    `json:"container_id"` // the actual Docker container ID
	Status       string    `json:"status"`
	CreatedAt    time.Time `json:"created_at"`
}
