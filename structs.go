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
	OwnerId primitive.ObjectID `bson:"owner_id"`
	OwnerName string `bson:"owner_name"`
	TimeCreated time.Time `bson:"time_created"`
}

// Contest helper methods
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

func (c Contest) IsOpen() bool {
	return c.State == OPEN
}

func (c Contest) IsVoting() bool {
	return c.State == VOTING
}

func (c Contest) IsConcluded() bool {
	return c.State == CONCLUDED
}

// ContestEntry collection in Mongo
type ContestEntry struct {
	Id primitive.ObjectID `bson:"_id"`
	ContestID primitive.ObjectID `bson:"contest_id"`
	ImagePath string `bson:"path"`
	Name string `bson:"title"`
	OwnerId primitive.ObjectID `bson:"owner_id"`
	OwnerName string `bson:"owner_name"`
}

func (c ContestEntry) GetStringId() string {
	return c.Id.Hex()
}

type ContestVote struct {
	Id primitive.ObjectID `bson:"_id"`
	ContestID primitive.ObjectID `bson:"contest_id"`
	EntryID primitive.ObjectID `bson:"entry_id"`
	UserID primitive.ObjectID `bson:"user_id"`
}

// Struct to hold data for rendering contest detail view
type ContestDetailData struct {
	Contest Contest
	ShowSubmitForm bool
	ShowVoteForm bool
	ShowEndSubmission bool
	ShowEndVoting bool
	EntryCount int64
	Entries []ContestEntry
}
