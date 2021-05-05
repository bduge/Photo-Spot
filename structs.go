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
	OwnerId string `bson:"owner_id"`
	OwnerName string `bson:"owner_name"`
	TimeCreated time.Time `bson:"time_created"`
}

func (c Contest) FormatTime() string {
	return c.TimeCreated.Format("Jan 2")
}

func (c Contest) GetStringId() string {
	return c.Id.Hex()
}

func (c Contest) GetStateString() string {
	if c.State == OPEN {
		return "Accepting Submissions"
	} else if c.State == VOTING {
		return "Voting in Progress"
	} else {
		return "Voting Concluded"
	}
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
