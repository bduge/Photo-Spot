package main

import (
	"context"
	"fmt"
	"html/template"
	"log"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/gorilla/sessions"

	// "go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const (
	OPEN = iota
	VOTING
	CONCLUDED
)

// User collection in Mongo
type User struct {
	Username string
	Password string
}

// Contest collection in Mongo
type Contest struct {
	Name string
	State int
	Description string
	Owner string
}

// ContestEntry collection in Mongo
type ContestEntry struct {
	ContestID string
	ImagePath string
	Title string
	Votes int
	Owner string
}

func getMongoClient() *mongo.Client {
	// Use local db instance for demonstration purposes
	uri := "mongodb://localhost:27017"
	clientOptions := options.Client().ApplyURI(uri)
	client, err := mongo.Connect(context.TODO(), clientOptions)
	if err != nil {
		log.Fatal(err)
	}

	err = client.Ping(context.TODO(), nil)

	if err != nil {
		log.Fatal(err)
	}

	return client
}

func main() {
	router := mux.NewRouter()

	// Hardcoded secret key for demonstration only purposes
	secretKey := "superdupersecret42"

	// // MongoDB setup
	client := getMongoClient()
	userCollection := client.Database("photospot").Collection("users")


	// // Setup cookie store for sessions
	// // Authentication logic from:
	// // https://thewhitetulip.gitbooks.io/webapp-with-golang-anti-textbook/content/manuscript/4.0authentication.html
	store := sessions.NewCookieStore([]byte(secretKey))

	// Serve static files
	fs := http.FileServer(http.Dir("static/"))
	router.PathPrefix("/static/").Handler(http.StripPrefix("/static/", fs))

	// Initialize templates
	tmplMap := make(map[string]*template.Template)
	tmplMap["index.html"] = template.Must(template.ParseFiles("static/index.html", "static/base.html"))
	tmplMap["signup.html"] = template.Must(template.ParseFiles("static/signup.html", "static/base.html"))
	tmplMap["login.html"] = template.Must(template.ParseFiles("static/login.html", "static/base.html"))
	// tmplMap["authForm.html"] = template.Must(template.ParseFiles("static/login.html", "static/base.html"))
	// tmplMap["index.html"] = template.Must(template.ParseFiles("static/index.html", "static/base.html"))

	
	// Routes
	router.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		tmplMap["index.html"].ExecuteTemplate(w, "base", nil)
	}).Methods("GET")

	router.HandleFunc("/signup", func(w http.ResponseWriter, r *http.Request) {
		signupHandler(w, r, store, tmplMap, userCollection)
	}).Methods("GET", "POST")

	router.HandleFunc("/login", func(w http.ResponseWriter, r *http.Request) {
		loginHandler(w, r, store, tmplMap, userCollection)
	}).Methods("GET", "POST")

	router.HandleFunc("/logout", func(w http.ResponseWriter, r *http.Request) {
		logoutHandler(w, r, store)
	}).Methods("POST")

	router.HandleFunc("/home", func(w http.ResponseWriter, r *http.Request) {
		homeHandler(w, r, store)
	}).Methods("GET")

	fmt.Println("Server running")
	http.ListenAndServe(":3000", router)
}