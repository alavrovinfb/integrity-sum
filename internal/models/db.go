package models

import "time"

type HashData struct {
	ID           int
	Hash         string
	FullFileName string
	Algorithm    string
	NamePod      string
}

type Release struct {
	ID        int
	Name      string
	CreatedAt time.Time
	Image     string
}
