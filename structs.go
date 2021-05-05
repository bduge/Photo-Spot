package main

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Enum types for contest state
const (
	OPEN = iota
	VOTING
	CONCLUDED
)

// User collection in Mongo
type User struct {
	Id primitive.ObjectID `bson:"_id"`
	Username string `bson:"username"`
	Password string `bson:"password"`
}

// Contest collection in Mongo
type Contest struct {
	Id primitive.ObjectID `bson:"_id"`
	Name string `bson:"name"`
	State int `bson:"state"`
	Description string `bson:"description"`
	Owner string `bson:"owner"`
	TimeCreated time.Time `bson:"time_created"`
}

// ContestEntry collection in Mongo
type ContestEntry struct {
	Id primitive.ObjectID `bson:"_id"`
	ContestID string
	ImagePath string
	Title string
	Votes int
	Owner string
}
