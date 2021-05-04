package main

import "time"

// Enum types for contest state
const (
	OPEN = iota
	VOTING
	CONCLUDED
)

// User collection in Mongo
type User struct {
	Username string
	Password string
}

// Contest collection in Mongo
type Contest struct {
	Name string
	State int
	Description string
	Owner string
	TimeCreated time.Time
}

// ContestEntry collection in Mongo
type ContestEntry struct {
	ContestID string
	ImagePath string
	Title string
	Votes int
	Owner string
}
