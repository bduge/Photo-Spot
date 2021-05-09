package main

import (
	"context"
	"log"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)


// Helper to check if user is able to make submission to contest
func canUserSubmit(
	userId primitive.ObjectID,
	contestId primitive.ObjectID,
	contestEntryCollection *mongo.Collection,
) bool {
	entryCount, countErr := contestEntryCollection.CountDocuments(
		context.TODO(),
		bson.D{{"contest_id", contestId}, {"owner_id", userId}},
	)
	if countErr != nil {
		log.Println(countErr)
		return false
	}
	return entryCount == 0
}

func canUserVote(
	userId primitive.ObjectID,
	contestId primitive.ObjectID,
	contestVoteCollection *mongo.Collection,
) bool {
	entryCount, countErr := contestVoteCollection.CountDocuments(
		context.TODO(),
		bson.D{{"contest_id", contestId}, {"user_id", userId}},
	)
	if countErr != nil {
		log.Println(countErr)
		return false
	}
	return entryCount == 0
}

// Check if current user can end submission period
func canEndSubmission(
	userId primitive.ObjectID,
	contest Contest,
) bool {
	return contest.State == OPEN && contest.OwnerId == userId
}

// Checks if current user can end voting period
func canEndVoting(
	userId primitive.ObjectID,
	contest Contest,
) bool {
	return contest.State == VOTING && contest.OwnerId == userId
}

// Get the number of submissions to a contest
func getNumSubmissions(
	contestId primitive.ObjectID,
	contestEntryCollection *mongo.Collection,
) int64 {
	entryCount, countErr := contestEntryCollection.CountDocuments(
		context.TODO(),
		bson.D{{"contest_id", contestId}},
	)
	if countErr != nil {
		log.Println(countErr)
		return -1
	}
	return entryCount
}

// Get the entries submitted to a contest
func getContestEntries(
	contestId primitive.ObjectID,
	contestEntryCollection *mongo.Collection,
) []ContestEntry {
	var entries []ContestEntry
	// Fetch entries into cursor
	cursor, err := contestEntryCollection.Find(context.TODO(), bson.D{{"contest_id", contestId}})
	if err != nil {
		log.Println("Couldn't find contests")
	}
	for cursor.Next(context.TODO()) {
		// For each entry, add to array (slice)
		var curEntry ContestEntry
		err := cursor.Decode(&curEntry)
		if err != nil {
			log.Println("Couldn't get contest")
		}
		entries = append(entries, curEntry)
	}
	if err := cursor.Err(); err != nil {
		log.Println(err)
	}
	cursor.Close(context.TODO())
	return entries
}

// getContestWinners(contestObjId, contestEntryCollection, contestVoteCollection)
func getContestWinners(
	contestId primitive.ObjectID,
	contestEntryCollection *mongo.Collection,
	contestVoteCollection *mongo.Collection,
) []ContestEntry {
	var winners []ContestEntry
	entries := getContestEntries(contestId, contestEntryCollection)
	maxVotes := 0

	for _, entry := range entries {
		votes, err := contestVoteCollection.CountDocuments(
			context.TODO(),
			bson.D{{"contest_id", contestId}, {"entry_id", entry.Id}},
		)
		if err != nil {
			log.Println(err)
			continue
		}
		if votes > int64(maxVotes) {
			maxVotes = int(votes)
			winners = []ContestEntry{entry}
		} else if votes == int64(maxVotes) {
			winners = append(winners, entry)
		}
	}
	return winners;
}
