package main

import (
	"context"
	"io/ioutil"
	"os"

	// "errors"
	// "fmt"
	"html/template"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/sessions"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

// Handler for /contests endpoint
func contestIndexHandler(
	w http.ResponseWriter,
	r *http.Request,
	s *sessions.CookieStore,
	tmplMap map[string]*template.Template,
	contestCollection *mongo.Collection,
) {
	// Fetch all contests
	// TODO: Create option to filter contests by state, name, etc
	var contests []*Contest
	cursor, err := contestCollection.Find(context.TODO(), bson.D{})
	if err != nil {
		log.Println("Couldn't find contests")
	}
	for cursor.Next(context.TODO()) {
		var curContest Contest
		err := cursor.Decode(&curContest)
		if err != nil {
			log.Println("Couldn't get contest")
		}
		contests = append(contests, &curContest)
	}
	if err := cursor.Err(); err != nil {
		log.Println(err)
	}

	cursor.Close(context.TODO())

	tmplMap["contests.html"].ExecuteTemplate(w, "base", contests)
}

// handler to render contest detail page
func contestDetailHandler(
	w http.ResponseWriter,
	r *http.Request,
	s *sessions.CookieStore,
	tmplMap map[string]*template.Template,
	contestCollection *mongo.Collection,
	contestEntryCollection *mongo.Collection,
	contestVoteCollection *mongo.Collection,
	contestId string,
) {
	// fetch necessary data
	var contest Contest
	contestObjId, idErr := primitive.ObjectIDFromHex(contestId)
	if idErr != nil {
		log.Println("Contest ID not valid")
		http.Redirect(w, r, "/contests", 302)
		return
	}
	err := contestCollection.FindOne(context.TODO(), bson.D{{"_id", contestObjId}}).Decode(&contest)
	if err != nil {
		log.Println("Contest not found")
		http.Redirect(w, r, "/contests", 302)
		return
	}
	session, err := s.Get(r, "session")
	if err != nil {
		log.Println(err)
		http.Redirect(w, r, "/contests", 302)
		return
	}
	userId, err := primitive.ObjectIDFromHex(session.Values["userId"].(string))
	if err != nil {
		log.Println(err)
		http.Redirect(w, r, "/contests", 302)
		return
	}
	entryCount := getNumSubmissions(contestObjId, contestEntryCollection)
	if contest.IsOpen() {
		// View for contest in open state
		tmplMap["contestDetailOpen.html"].ExecuteTemplate(w, "base", ContestDetailData{
			Contest: contest,
			ShowSubmitForm: canUserSubmit(userId, contestObjId, contestEntryCollection),
			EntryCount: entryCount,
			ShowEndSubmission: canEndSubmission(userId, contest),
		})
	} else if contest.IsVoting() {
		// View for contest in voting state
		entries := getContestEntries(contestObjId, contestEntryCollection)
		tmplMap["contestDetailVoting.html"].ExecuteTemplate(w, "base", ContestDetailData{
			Contest: contest,
			Entries: entries,
			ShowVoteForm: canUserVote(userId, contestObjId, contestVoteCollection),
			EntryCount: entryCount,
			ShowEndVoting: canEndVoting(userId, contest),
		})
	} else {
		// View for concluded contest
		winners := getContestWinners(contestObjId, contestEntryCollection, contestVoteCollection)
		tmplMap["contestDetailConcluded.html"].ExecuteTemplate(w, "base", ContestDetailData{
			Contest: contest,
			Entries: winners,
			EntryCount: entryCount,
		})
	}
}

// Handler for contest submission endpoint
func contestPhotoSubmissionHandler(
	w http.ResponseWriter,
	r *http.Request,
	s *sessions.CookieStore,
	tmplMap map[string]*template.Template,
	contestEntryCollection *mongo.Collection,
	contestId string,
) {
	// Get data and format IDs
	session, err := s.Get(r, "session")
	if err != nil {
		log.Println(err)
		http.Redirect(w, r, "/contests", 302)
		return
	}
	contestOwnerName := session.Values["username"].(string)
	entryOwnerId, err := primitive.ObjectIDFromHex(session.Values["userId"].(string))
	if err != nil {
		log.Println(err)
		http.Redirect(w, r, "/contests/", 302)
		return
	}
	contestObjId, err := primitive.ObjectIDFromHex(contestId)
	if err != nil {
		log.Println(err)
		http.Redirect(w, r, "/contests/", 302)
		return
	}

	// Check if user is allowed to make submission
	if !canUserSubmit(entryOwnerId, contestObjId, contestEntryCollection) {
		log.Println("User doesn't have permission to enter contest")
		http.Redirect(w, r, "/contests/" + contestId, 302)
		return
	}

	// Fetch image from form
	// Max image size of 10 MB
	r.ParseMultipartForm(10 << 20)
	uploadedFile, handler, err := r.FormFile("img")
	if err != nil {
		log.Println("Couldn't fetch file")
		http.Redirect(w, r, "/contests/" + contestId, 302)
		return
	}
	defer uploadedFile.Close()

	// Create new file on server to store image
	entryId := primitive.NewObjectID()
	entryName := r.PostFormValue("imgName")
	imagePath := "uploadedImages/" + entryId.Hex() + handler.Filename
	newFile, err := os.Create(imagePath)
	if err != nil {
		log.Printf("Issue saving file %v\n", err)
		http.Redirect(w, r, "/contests/" + contestId, 302)
		return
	}

	// Write data to new file
	fileBytes, err := ioutil.ReadAll(uploadedFile)
	if err != nil {
		log.Println("Couldn't read file")
		http.Redirect(w, r, "/contests/" + contestId, 302)
		return
	}
	newFile.Write(fileBytes)
	
	// Save entry in database
	newEntry := ContestEntry{
		entryId,
		contestObjId,
		"/" + imagePath,
		entryName,
		entryOwnerId,
		contestOwnerName,
	}
	_, insertErr := contestEntryCollection.InsertOne(context.TODO(), newEntry)
	if insertErr != nil {
		log.Println(insertErr)
		http.Redirect(w, r, "/contests", 302)
		return
	}
	http.Redirect(w, r, "/contests/" + contestId, 302)
	return
}

// Handler to create new contest
func createContestHandler(
	w http.ResponseWriter,
	r *http.Request,
	s *sessions.CookieStore,
	tmplMap map[string]*template.Template,
	contestCollection *mongo.Collection,
) {
	if r.Method == "POST" {
		session, err := s.Get(r, "session")
		if err != nil {
			log.Println(err)
			http.Redirect(w, r, "/contests", 302)
			return
		}
		// Fetch data from form and current session
		contestName := r.PostFormValue("contestname")
		contestDescription := r.PostFormValue("contestdescription")
		contestOwnerId := session.Values["userId"].(string)
		ownerObjId, idErr := primitive.ObjectIDFromHex(contestOwnerId)
		if idErr != nil {
			log.Println("User ID not valid")
			http.Redirect(w, r, "/contests", 302)
			return
		}
		contestOwnerName := session.Values["username"].(string)
		currentTime := time.Now()

		// Create contest and save in database
		newContest := Contest{
			primitive.NewObjectID(),
			contestName,
			OPEN,
			contestDescription,
			ownerObjId,
			contestOwnerName,
			currentTime,
		}
		insertResult, insertErr := contestCollection.InsertOne(context.TODO(), newContest)
		if insertErr != nil {
			log.Println(insertErr)
			http.Redirect(w, r, "/contests", 302)
			return
		}
		http.Redirect(w, r, "/contests/" + insertResult.InsertedID.(primitive.ObjectID).Hex(), 302)
		return
	} else {
		// Render create contest form
		tmplMap["createContest.html"].ExecuteTemplate(w, "base", nil)
		return
	}
}

// Handler for request to start contest voting period
// TODO: VERIFY CONTEST CAN START VOTING
func contestChangeStateHandler(
	w http.ResponseWriter,
	r *http.Request,
	s *sessions.CookieStore,
	contestCollection *mongo.Collection,
	contestId string,
	state int,
) {
	contestObjId, err := primitive.ObjectIDFromHex(contestId)
	if err != nil {
		log.Println(err)
		http.Redirect(w, r, "/contests", 302)
		return
	}
	update := bson.D{{"$set", bson.D{{"state", state}}}}
	_, updateErr := contestCollection.UpdateOne(
		context.TODO(),
		bson.D{{"_id", contestObjId}},
		update,
	)
	if updateErr != nil {
		log.Println(err)
	}
	http.Redirect(w, r, "/contests/" + contestId, 302)
}

func contestVoteHandler(
	w http.ResponseWriter,
	r *http.Request,
	s *sessions.CookieStore,
	contestCollection *mongo.Collection,
	contestVoteCollection *mongo.Collection,
	contestEntryCollection *mongo.Collection,
	contestId string,
) {
	session, err := s.Get(r, "session")
	if err != nil {
		log.Println(err)
		http.Redirect(w, r, "/contests", 302)
		return
	}
	// Verify contestEntry exists and can be voted on
	imageVote := r.PostFormValue("image-vote")
	entryId, err := primitive.ObjectIDFromHex(imageVote)
	if err != nil {
		log.Println(err)
		http.Redirect(w, r, "/contests", 302)
		return
	}
	contestObjId, err := primitive.ObjectIDFromHex(contestId)
	if err != nil {
		log.Println(err)
		http.Redirect(w, r, "/contests", 302)
		return
	}
	entryCount, countErr := contestEntryCollection.CountDocuments(
		context.TODO(),
		bson.D{{"_id", entryId}, {"contest_id", contestObjId}},
	)
	if countErr != nil {
		log.Println(countErr)
		http.Redirect(w, r, "/contests", 302)
		return
	}
	if entryCount != 1 {
		log.Println("Request not valid")
		http.Redirect(w, r, "/contests", 302)
		return
	}
	
	// Validate current user
	voterId, err := primitive.ObjectIDFromHex(session.Values["userId"].(string))
	if err != nil {
		log.Println(err)
		http.Redirect(w, r, "/contests/", 302)
		return
	}

	// Create vote object and store
	newContestVote := ContestVote{
		primitive.NewObjectID(),
		contestObjId,
		entryId,
		voterId,
	}
	_, insertErr := contestVoteCollection.InsertOne(context.TODO(), newContestVote)
	if insertErr != nil {
		log.Println(insertErr)
		http.Redirect(w, r, "/contests", 302)
		return
	}
	http.Redirect(w, r, "/contests/" + contestId, 302)
	return
}

// *******
// Helpers
// *******

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

