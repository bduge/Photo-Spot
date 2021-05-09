package main

import (
	"testing"
	"time"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Test contest methods
func createContest(id primitive.ObjectID, ownerId primitive.ObjectID, state int) Contest {
	return Contest {
		id,
		"test contest",
		state,
		"contest for unit test",
		ownerId,
		"Bill",
		time.Now(),
	}
}

func TestOpenContest(t *testing.T){
	contest := createContest(primitive.NewObjectID(), primitive.NewObjectID(), OPEN)
	if contest.IsVoting() || contest.IsConcluded() || !contest.IsOpen() {
		t.Error("Contest state should be OPEN")
	}
}

func TestVotingContest(t *testing.T){
	contest := createContest(primitive.NewObjectID(), primitive.NewObjectID(), VOTING)
	if !contest.IsVoting() || contest.IsConcluded() || contest.IsOpen() {
		t.Error("Contest state should be VOTING")
	}
}

func TestConcludedContest(t *testing.T){
	contest := createContest(primitive.NewObjectID(), primitive.NewObjectID(), CONCLUDED)
	if contest.IsVoting() || !contest.IsConcluded() || contest.IsOpen() {
		t.Error("Contest state should be CONCLUDED")
	}
}

func TestContestGetId(t *testing.T){
	newId := primitive.NewObjectID()
	contest := createContest(newId, primitive.NewObjectID(), OPEN)
	if newId.Hex() != contest.GetStringId() {
		t.Error("IDs do not match")
	}
}
