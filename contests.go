package main

import (
	"context"
	"io/ioutil"
	"os"

	// "errors"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/sessions"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	// "go.mongodb.org/mongo-driver/mongo/options"
)

// type ContestEntry struct {
// 	ContestId string
// 	ImagePath string
// 	Title string
// 	Votes int
// 	OwnerId string
// }
// type Contest struct {
// 	Name string
// 	State int
// 	Description string
// 	Owner string
// 	TimeCreated time.Time
// }

type ContestDetailData struct {
	Contest Contest
	ShowSubmitForm bool
	ShowVoteForm bool
}

func contestIndexHandler(
	w http.ResponseWriter,
	r *http.Request,
	s *sessions.CookieStore,
	tmplMap map[string]*template.Template,
	contestCollection *mongo.Collection,
) {
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

func contestDetailHandler(
	w http.ResponseWriter,
	r *http.Request,
	s *sessions.CookieStore,
	tmplMap map[string]*template.Template,
	contestCollection *mongo.Collection,
	contestEntryCollection *mongo.Collection,
	contestId string,
) {
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
	if contest.IsOpen() {
		temp := canUserEnter(userId, contestObjId, contestEntryCollection)
		fmt.Println(temp)
		tmplMap["contestDetailOpen.html"].ExecuteTemplate(w, "base", ContestDetailData{
			Contest: contest,
			ShowSubmitForm: temp,
		})
	}
	
}

func contestPhotoSubmissionHandler(
	w http.ResponseWriter,
	r *http.Request,
	s *sessions.CookieStore,
	tmplMap map[string]*template.Template,
	contestEntryCollection *mongo.Collection,
	contestId string,
) {
	session, err := s.Get(r, "session")
	if err != nil {
		log.Println(err)
		http.Redirect(w, r, "/contests", 302)
		return
	}
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

	// Max file size of 10 MB
	r.ParseMultipartForm(10 << 20)
	uploadedFile, handler, err := r.FormFile("img")
	if err != nil {
		log.Println("Couldn't fetch file")
		http.Redirect(w, r, "/contests/" + contestId, 302)
		return
	}
	defer uploadedFile.Close()

	entryId := primitive.NewObjectID()
	entryName := r.PostFormValue("imgName")
	imagePath := "uploadedImages/" + entryId.Hex() + handler.Filename
	newFile, err := os.Create(imagePath)
	if err != nil {
		log.Printf("Issue saving file %v\n", err)
		http.Redirect(w, r, "/contests/" + contestId, 302)
		return
	}

	fileBytes, err := ioutil.ReadAll(uploadedFile)
	if err != nil {
		log.Println("Couldn't read file")
		http.Redirect(w, r, "/contests/" + contestId, 302)
		return
	}
	newFile.Write(fileBytes)
	
	newEntry := ContestEntry{
		entryId,
		contestObjId,
		imagePath,
		entryName,
		entryOwnerId,
		0,
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
		contestName := r.PostFormValue("contestname")
		contestDescription := r.PostFormValue("contestdescription")
		contestOwnerId := session.Values["userId"].(string)
		contestOwnerName := session.Values["username"].(string)
		currentTime := time.Now()
		newContest := Contest{
			primitive.NewObjectID(),
			contestName,
			OPEN,
			contestDescription,
			contestOwnerId,
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
		tmplMap["createContest.html"].ExecuteTemplate(w, "base", nil)
		return
	}
	
}


// Helpers

func canUserEnter(
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
	fmt.Println(entryCount, userId, contestId)
	return entryCount == 0
}
