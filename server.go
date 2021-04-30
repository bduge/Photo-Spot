package main

import (
	"context"
	// "fmt"
	"html/template"
	"log"
	"net/http"

	"github.com/gorilla/mux"
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
	userID string
	password string
}

// Contest collection in Mongo
type Contest struct {
	Name string
	State int
	Password string
}

// ContestEntry collection in Mongo
type ContestEntry struct {
	ContestID string
	ImagePath string
	Votes int
	Owner string
}

func connectDB() {
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
}

func main() {
	router := mux.NewRouter()
	// connectDB()

	// Serve static files
	fs := http.FileServer(http.Dir("static/"))
	tmplMap := make(map[string]*template.Template)
	tmplMap["index.html"] = template.Must(template.ParseFiles("static/index.html", "static/base.html"))

	router.PathPrefix("/static/").Handler(http.StripPrefix("/static/", fs))
	router.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		tmplMap["index.html"].ExecuteTemplate(w, "base", "HELLO WORLD 555")
	}).Methods("GET")

	http.ListenAndServe(":3000", router)
}