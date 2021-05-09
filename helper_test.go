package main

import (
	"testing"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Contest helper tests
func TestCanEndSubmissionSameId(t *testing.T){
	newId := primitive.NewObjectID()
	sameContest := createContest(primitive.NewObjectID(), newId, OPEN)
	if !canEndSubmission(newId, sameContest){
		t.Error("Contest should be able to start vote")
	}
}

func TestCanEndSubmissionDiffId(t *testing.T){
	newId := primitive.NewObjectID()
	sameContest := createContest(primitive.NewObjectID(), primitive.NewObjectID(), OPEN)
	if canEndSubmission(newId, sameContest){
		t.Error("Contest should not be able to start vote")
	}
}

func TestCanEndSubmissionWrongState(t *testing.T){
	newId := primitive.NewObjectID()
	votingContest := createContest(primitive.NewObjectID(), newId, VOTING)
	if canEndSubmission(newId, votingContest){
		t.Error("Voting contest should not be able to start vote")
	}
	concludedContest := createContest(primitive.NewObjectID(), newId, CONCLUDED)
	if canEndSubmission(newId, concludedContest){
		t.Error("Concluded contest should not be able to start vote")
	}
}


func TestCanEndVotingSameId(t *testing.T){
	newId := primitive.NewObjectID()
	sameContest := createContest(primitive.NewObjectID(), newId, VOTING)
	if !canEndVoting(newId, sameContest){
		t.Error("Contest should be able to start vote")
	}
}

func TestCanEndVotingDiffId(t *testing.T){
	newId := primitive.NewObjectID()
	sameContest := createContest(primitive.NewObjectID(), primitive.NewObjectID(), VOTING)
	if canEndVoting(newId, sameContest){
		t.Error("Contest should not be able to start vote")
	}
}

func TestCanEndVotingWrongState(t *testing.T){
	newId := primitive.NewObjectID()
	openContest := createContest(primitive.NewObjectID(), newId, OPEN)
	if canEndVoting(newId, openContest){
		t.Error("Open contest should not be able to start vote")
	}
	concludedContest := createContest(primitive.NewObjectID(), newId, CONCLUDED)
	if canEndVoting(newId, concludedContest){
		t.Error("Concluded contest should not be able to start vote")
	}
}
