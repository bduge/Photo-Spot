package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"time"
	"net/http"
	// "html/template"
	"github.com/gorilla/sessions"
	// "go.mongodb.org/mongo-driver/bson"
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


func createContestHandler(
	w http.ResponseWriter,
	r *http.Request,
	s *sessions.CookieStore,
	contestCollection *mongo.Collection,
) {
	if r.Method == "POST" {
		session, err := s.Get(r, "session")
		if err != nil {
			log.Println(err)
			http.Redirect(w, r, "/contests", 500)
			return
		}
		contestName := r.PostFormValue("contestname")
		contestDescription := r.PostFormValue("contestdescription")
		contestOwner := session.Values["username"].(string)
		currentTime := time.Now()

		newContest := Contest{contestName, OPEN, contestDescription, contestOwner, currentTime}
		insertResult, insertErr := contestCollection.InsertOne(context.TODO(), newContest)
		if insertErr != nil {
			log.Println(insertErr)
			http.Redirect(w, r, "/contests", 500)
			return
		}
		http.Redirect(w, r, fmt.Sprintf("/contests/%s", insertResult.InsertedID), 200)
	} else {
		
	}
	
}