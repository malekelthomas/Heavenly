package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"

	"golang.org/x/crypto/bcrypt"

	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type User struct {
	/* ID          string `json:"ID"` */
	Name  string `json:"Name"`
	Email string `json:"Email"`
	/* DOB 		string `json:"DOB"` */
	Password string `json:"Password"`
}

type Cake struct {
	Type      string `json:"Type"`
	Size      int    `json:"Size"`
	Inventory int    `json:"Inventory"`
}

var client *mongo.Client
var env = godotenv.Load()

func connectToDB() (*mongo.Client, error) {

	clientOptions := options.Client().ApplyURI(os.Getenv("MONGO_CONNECTION"))
	client, err := mongo.Connect(context.TODO(), clientOptions)
	if err != nil {
		return nil, err
	}
	return client, err
}

func insertUser(name string, email string, password string) {
	user := User{name, email, password}
	collection := client.Database(os.Getenv("DB_NAME")).Collection("users")

	insertResults, err := collection.InsertOne(context.TODO(), user)

	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Inserted user with ID:", insertResults.InsertedID)
}

func findUserByName(name string) {
	collection := client.Database(os.Getenv("DB_NAME")).Collection("users")

	var user bson.M
	err := collection.FindOne(context.TODO(), bson.M{"name": name}).Decode(&user)
	if err != nil {
		fmt.Println(err)
	}

	fmt.Println("Found User", user["name"])
}

func registerUser(w http.ResponseWriter, r *http.Request) {
	var newUser User
	reqBody, err := ioutil.ReadAll(r.Body)
	if err != nil {
		fmt.Fprintf(w, "Please enter name, email, and password")
	}
	json.Unmarshal(reqBody, &newUser)
	hash, err := bcrypt.GenerateFromPassword([]byte(newUser.Password), bcrypt.MinCost)
	if err != nil {
		fmt.Println(err)
	}
	newUser.Password = string(hash)
	insertUser(newUser.Name, newUser.Email, newUser.Password)
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(newUser)

}

func getAllUsersFromDB() []bson.M {
	collection := client.Database(os.Getenv("DB_NAME")).Collection("users")
	cursor, err := collection.Find(context.TODO(), bson.M{})
	if err != nil {
		log.Fatal(err)
	}

	var users []bson.M
	if err = cursor.All(context.TODO(), &users); err != nil {
		log.Fatal(err)
	}

	return users

}

func getAllUsers(w http.ResponseWriter, r *http.Request) {
	z := getAllUsersFromDB()
	json.NewEncoder(w).Encode(z)
}

func getUser(w http.ResponseWriter, r *http.Request) {
	username := mux.Vars(r)["username"]
	findUserByName(username)

}

func insertCake(cake_type string, size int, inventory int) {
	cake := Cake{cake_type, size, inventory}
	collection := client.Database(os.Getenv("DB_NAME")).Collection("cakes")
	insertResults, err := collection.InsertOne(context.TODO(), cake)

	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Cake Type of %s with size %d inventory: %d, Inserted with ID: %v", cake.Type, cake.Size, cake.Inventory, insertResults.InsertedID)
}

func addCake(w http.ResponseWriter, r *http.Request) {
	var newCake Cake
	reqBody, err := ioutil.ReadAll(r.Body)
	if err != nil {
		fmt.Fprintf(w, "Please enter Type, Size, and Inventory")
	}
	json.Unmarshal(reqBody, &newCake)
	insertCake(newCake.Type, newCake.Size, newCake.Inventory)
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(newCake)
}

func findCakeByType(cake_type string) {
	collection := client.Database(os.Getenv("DB_NAME")).Collection("cakes")

	var cake bson.M
	err := collection.FindOne(context.TODO(), bson.M{"type": cake_type}).Decode(&cake)
	if err != nil {
		fmt.Println(err)
	}

	fmt.Println("Found Cake", cake["type"])
}

func getCake(w http.ResponseWriter, r *http.Request) {
	cakeType := mux.Vars(r)["cakeType"]
	findCakeByType(cakeType)
}

func homeLink(w http.ResponseWriter, r *http.Request) {
	fmt.Fprint(w, "Welcome home!")
}
func main() {
	var err error
	client, err = connectToDB()
	if err != nil {
		log.Fatal(err)
	}
	r := mux.NewRouter().StrictSlash(true)
	r.HandleFunc("/", homeLink)
	r.HandleFunc("/users", registerUser).Methods("POST")
	r.HandleFunc("/users", getAllUsers).Methods("GET")
	r.HandleFunc("/cakes", addCake).Methods("POST")
	r.HandleFunc("/users/{username}", getUser).Methods("GET")
	r.HandleFunc("/cakes/{cakeType:.*}", getCake).Methods("GET")
	/* r.HandleFunc("/event", createEvent).Methods("POST")
	r.HandleFunc("/event/{id}", getEvent).Methods("GET")
	r.HandleFunc("/event/{id}", updateEvent).Methods("PATCH")
	r.HandleFunc("/events", getAllEvents).Methods("GET")
	*/

	log.Fatal(http.ListenAndServe(":3000", r))
}
