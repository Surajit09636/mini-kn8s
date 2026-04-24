package models

import "time"

type Deployment struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	UserID    uint      `json:"user_id"` // which User owns this Deployment
	Image     string    `json:"image"`
	Replicas  int       `json:"replicas"`
	Status    string    `json:"status"` //pending, running, failed
	CreatedAt time.Time `json:"created_at"`
	Pods      []Pod     `gorm:"foreignKey:DeploymentID" json:"pods"` //Replational link to pods
}
