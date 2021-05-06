package main

import (
	"context"
	"errors"
	// "fmt"
	"html/template"
	"log"
	"net/http"
	"time"
	"github.com/gorilla/sessions"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// ***********
// Data Struct
// ***********

type AuthFormData struct {
	Header string
	FormUrl string
	ButtonText string
	RedirectUrl string
	RedirectText string
}

// ********
// Handlers
// ********

// Handler for /logout endpoint
func logoutHandler(w http.ResponseWriter, r *http.Request, s *sessions.CookieStore) {
    session, err := s.Get(r, "session")
    if err == nil { //If there is no error, then remove session
		if session.Values["loggedin"] != "false" {
			session.Values["loggedin"] = "false"
			session.Save(r, w)
		}
    }
    //redirect to login regardless of an error
    http.Redirect(w, r, "/login", 302) 
}


// Handler for /login endpoint
func loginHandler(
	w http.ResponseWriter,
	r *http.Request,
	s *sessions.CookieStore,
	tmplMap map[string]*template.Template,
	userCollection *mongo.Collection,
) {
    session, err := s.Get(r, "session")
	loginData := AuthFormData{
		"Log In",
		"/login",
		"Log In",
		"/signup",
		"Create an Account",
	}
    if err != nil {
		// in case of error during fetching session info, show login page
		tmplMap["authForm.html"].ExecuteTemplate(w, "base", loginData) 
    } else {
		isLoggedIn := session.Values["loggedin"]
		if isLoggedIn != "true" {
			if r.Method == "POST" {
				username := r.PostFormValue("username")
				password := r.PostFormValue("password")
				// Attempt to log user in
				if (verifyCredentials(username, password, userCollection)) {
					userId := getUserId(username, userCollection)
					session.Values["loggedin"] = "true"
					session.Values["username"] = username
					session.Values["userId"] = userId.Hex()
					session.Save(r, w)
					http.Redirect(w, r, "/contests", 302)
					return
				}
			}
			tmplMap["authForm.html"].ExecuteTemplate(w, "base", loginData)
			return
		}
		// Redirect to contests page if logged in
		http.Redirect(w, r, "/contests", 302)
    }
}

// Handler for /signup endpoint
func signupHandler(
	w http.ResponseWriter,
	r *http.Request,
	s *sessions.CookieStore,
	tmplMap map[string]*template.Template,
	userCollection *mongo.Collection,
) {
	if r.Method == "POST" {
		// Attemp to create user
		username := r.PostFormValue("username")
		password := r.PostFormValue("password")
		err := createNewUser(username, password, userCollection)
		if err != nil {
			// Redirect back to sign up page user cannot be created
			http.Redirect(w, r, "/signup", 302)	
			return
		}
		// Redirect to login page if successful
		http.Redirect(w, r, "/login", 302)
		return
	} else {
		// Display sign up page for GET request
		signupData := AuthFormData{
			"Create an Account",
			"/signup",
			"Sign Up",
			"/login",
			"Already have an account? Log in here",
		}
		tmplMap["authForm.html"].ExecuteTemplate(w, "base", signupData)
		return
	}
}

// Handler for /home endpoint
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

// Helper for login required endpoints
func loginRequiredHandlerMixin(
	w http.ResponseWriter,
	r *http.Request,
	s *sessions.CookieStore,
) bool {
	if !isLoggedIn(r, s){
		log.Println("Not authorized for this request")
		http.Redirect(w, r, "/login", 302)
		return true
	}
	return false
}


// *******
// Helpers
// *******

// Return userId as a Mongo ObjectId
func getUserId(username string, userCollection *mongo.Collection) primitive.ObjectID {
	var result User
	err := userCollection.FindOne(context.TODO(), bson.D{{"username", username}}).Decode(&result)
	if err != nil {
		log.Println("User not found")
	}
	return result.Id
}

// Return if user is logged in
func isLoggedIn(r *http.Request, s *sessions.CookieStore) bool {
	session, _ := s.Get(r, "session")
	if session.Values["loggedin"] == "true" {
		return true
	}
	return false
}

// Returns if user credentials are valid
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

// Create new user in database
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
	newUser := User{primitive.NewObjectID(), username, password}
	_, insertErr := userCollection.InsertOne(context.TODO(), newUser)
	if insertErr != nil {
		return insertErr
	}
	return nil
}
