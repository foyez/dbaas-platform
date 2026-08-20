package dbaas

import "time"

type Instance struct {
	ID             string    `json:"id"`
	Name           string    `json:"name"`
	Version        int       `json:"version"`
	Storage        string    `json:"storage"`
	Instances      int       `json:"instances"`
	ReadyInstances int       `json:"readyInstances"`
	Status         string    `json:"status"`
	CreatedAt      time.Time `json:"createdAt"`
}

type InstanceCredentials struct {
	Host     string `json:"host"`
	Port     int    `json:"port"`
	Database string `json:"database"`
	Username string `json:"username"`
	Password string `json:"password"`
	URI      string `json:"uri"`
}

type LogLine struct {
	Timestamp string `json:"timestamp"`
	Line      string `json:"line"`
}
