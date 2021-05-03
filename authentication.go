package main

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"html/template"
	"time"
	"log"
	"github.com/gorilla/sessions"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// ********
// Handlers
// ********

func logoutHandler(w http.ResponseWriter, r *http.Request, s *sessions.CookieStore) {
    session, err := s.Get(r, "session")
    if err == nil { //If there is no error, then remove session
		if session.Values["loggedin"] != "false" {
			session.Values["loggedin"] = "false"
			session.Save(r, w)
		}
    }
    http.Redirect(w, r, "/login", 302) 
    //redirect to login irrespective of error or not
}

func loginHandler(
	w http.ResponseWriter,
	r *http.Request,
	s *sessions.CookieStore,
	tmplMap map[string]*template.Template,
	userCollection *mongo.Collection,
) {
    session, err := s.Get(r, "session")

    if err != nil {
		tmplMap["login.html"].ExecuteTemplate(w, "base", nil) 
		// in case of error during 
		// fetching session info, execute login template
    } else {
		isLoggedIn := session.Values["loggedin"]
		if isLoggedIn != "true" {
			if r.Method == "POST" {
				fmt.Println("Login POST called")
				username := r.PostFormValue("username")
				password := r.PostFormValue("password")
				if (verifyCredentials(username, password, userCollection)) {
					session.Values["loggedin"] = "true"
					session.Values["username"] = username
					session.Save(r, w)
					http.Redirect(w, r, "/contests", 302)
					return
				}
			}
			tmplMap["login.html"].ExecuteTemplate(w, "base", nil)
			return
		}
		http.Redirect(w, r, "/contests", 302)
    }
}

func signupHandler(
	w http.ResponseWriter,
	r *http.Request,
	s *sessions.CookieStore,
	tmplMap map[string]*template.Template,
	userCollection *mongo.Collection,
) {
	if r.Method == "POST" {
		fmt.Println("Signup POST called")
		username := r.PostFormValue("username")
		password := r.PostFormValue("password")
		fmt.Printf("%s %s\n", username, password)
		err := createNewUser(username, password, userCollection)
		if err != nil {
			http.Redirect(w, r, "/signup", 302)	
		}
		http.Redirect(w, r, "/login", 302)
	} else {
		tmplMap["signup.html"].ExecuteTemplate(w, "base", nil)
	}
}

func homeHandler(
	w http.ResponseWriter,
	r *http.Request,
	s *sessions.CookieStore,
) {
	if isLoggedIn(r, s) {
		http.Redirect(w, r, "/contests", 302)
	}
	http.Redirect(w, r, "/", 302)
}


// *******
// Helpers
// *******

func isLoggedIn(r *http.Request, s *sessions.CookieStore) bool {
	session, _ := s.Get(r, "session")
	if session.Values["loggedin"] == "true" {
		return true
	}
	return false
}

func verifyCredentials(username string, password string, userCollection *mongo.Collection) bool {
	var user User
	if (username == "" || password == "") {
		return false
	}
	err := userCollection.FindOne(context.TODO(), bson.D{{"username", username}}).Decode(&user)
	if err != nil {
		log.Print(err)
		return false
	}
	if user.Password != password {
		return false
	}
	return true
}

func createNewUser(username string, password string, userCollection *mongo.Collection) error {
	if username == "" || password == "" {
		return errors.New("Invalid username or password")
	}
	opts := options.Count().SetLimit(1).SetMaxTime(5 * time.Second)
	count, countErr := userCollection.CountDocuments(context.TODO(), bson.D{{"username", username}}, opts)
	if countErr != nil {
		return countErr
	}
	if count != 0 {
		return errors.New("User ID already exists")
	}
	newUser := User{username, password}
	fmt.Printf("%v\n", newUser)
	_, insertErr := userCollection.InsertOne(context.TODO(), newUser)
	if insertErr != nil {
		return insertErr
	}
	return nil
}