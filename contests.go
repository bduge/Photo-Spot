package main

import (
	"context"
	// "errors"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/sessions"

	// "go.mongodb.org/mongo-driver/bson"
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

type ContestData struct {
	contests []Contest
}

func contestIndexHandler(
	w http.ResponseWriter,
	r *http.Request,
	s *sessions.CookieStore,
	tmplMap map[string]*template.Template,
	contestCollection *mongo.Collection,
) {

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
		contestOwner := session.Values["userId"].(string)
		currentTime := time.Now()
		newContest := Contest{
			primitive.NewObjectID(),
			contestName,
			OPEN,
			contestDescription,
			contestOwner,
			currentTime,
		}
		insertResult, insertErr := contestCollection.InsertOne(context.TODO(), newContest)
		if insertErr != nil {
			log.Println(insertErr)
			http.Redirect(w, r, "/contests", 302)
			return
		}
		http.Redirect(w, r, fmt.Sprintf("/contests/%s", insertResult.InsertedID.(primitive.ObjectID).Hex()), 302)
		return
	} else {
		tmplMap["createContest.html"].ExecuteTemplate(w, "base", nil)
		return
	}
	
}