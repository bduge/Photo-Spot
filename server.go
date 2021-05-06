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


// Connect to MongoDB
func getMongoClient(uri string) *mongo.Client {
	// Use local db instance for demonstration purposes
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
	// HARD VALUES FOR DEVELOPMENT/DEMONSTRATION PURPOSES ONLY
	dbName := "photospot"
	dbUri := "mongodb://localhost:27017"
	secretKey := "superdupersecret42"

	// MongoDB setup
	client := getMongoClient(dbUri)
	userCollection := client.Database(dbName).Collection("users")
	contestCollection := client.Database(dbName).Collection("contests")
	contestEntryCollection := client.Database(dbName).Collection("contestEntries")

	// Setup cookie store for sessions
	// Authentication logic from:
	// https://thewhitetulip.gitbooks.io/webapp-with-golang-anti-textbook/content/manuscript/4.0authentication.html
	store := sessions.NewCookieStore([]byte(secretKey))

	// Serve static files
	fs := http.FileServer(http.Dir("static/"))
	router := mux.NewRouter()
	router.PathPrefix("/static/").Handler(http.StripPrefix("/static/", fs))

	// Initialize templates
	tmplMap := make(map[string]*template.Template)
	tmplMap["index.html"] = template.Must(template.ParseFiles("static/index.html", "static/base.html"))
	tmplMap["authForm.html"] = template.Must(template.ParseFiles("static/authForm.html", "static/base.html"))
	tmplMap["contests.html"] = template.Must(template.ParseFiles("static/contests.html", "static/base.html"))
	tmplMap["createContest.html"] = template.Must(template.ParseFiles("static/createContest.html", "static/base.html"))
	tmplMap["contestDetailOpen.html"] = template.Must(template.ParseFiles(
		"static/contestDetailOpen.html",
		"static/contestDetail.html",
		"static/base.html",
	))

	// Index route
	router.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		tmplMap["index.html"].ExecuteTemplate(w, "base", nil)
	}).Methods("GET")

	// Authentication routes
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


	// Contest routes (authentication needed)
	router.HandleFunc("/contests", func(w http.ResponseWriter, r *http.Request) {
		if loginRequiredHandlerMixin(w, r, store) {
			return
		}
		contestIndexHandler(w, r, store, tmplMap, contestCollection)
	}).Methods("GET")

	router.HandleFunc("/contests/{contestId}", func(w http.ResponseWriter, r *http.Request) {
		if loginRequiredHandlerMixin(w, r, store) {
			return
		}
		vars := mux.Vars(r)
		contestId := vars["contestId"]
		contestDetailHandler(w, r, store, tmplMap, contestCollection, contestEntryCollection, contestId)
	}).Methods("GET")

	router.HandleFunc("/contests/{contestId}/submit", func(w http.ResponseWriter, r *http.Request) {
		if loginRequiredHandlerMixin(w, r, store) {
			return
		}
		vars := mux.Vars(r)
		contestId := vars["contestId"]
		contestPhotoSubmissionHandler(w, r, store, tmplMap, contestEntryCollection, contestId)
	}).Methods("POST")

	router.HandleFunc("/create-contest", func(w http.ResponseWriter, r *http.Request) {
		if loginRequiredHandlerMixin(w, r, store) {
			return
		}
		createContestHandler(w, r, store, tmplMap, contestCollection)
	}).Methods("GET", "POST")

	// Start server
	fmt.Println("Server running")
	http.ListenAndServe(":3000", router)
}