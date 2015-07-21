package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"

	"github.com/gorilla/mux"
	_ "github.com/lib/pq"
)

type Client struct {
	Id      int64  `json:"id"`
	Name    string `json:"name"`
	Email   string `json:"email"`
	Company string `json:"company"`
	Country string `json:"country"`
}

var db *sql.DB

func main() {
	var err error
	db, err = sql.Open("postgres", "postgres://tmpjr@localhost/postgres?sslmode=disable")
	if err != nil {
		log.Fatalf("Could not open connection: %v", err)
	}

	err = db.Ping()
	if err != nil {
		log.Fatalf("Error on opening database connection: %s", err.Error())
	}

	_, err = db.Exec("SET search_path TO playground")
	if err != nil {
		log.Fatal(err)
	}

	router := mux.NewRouter().StrictSlash(true)
	router.HandleFunc("/", Index)
	router.HandleFunc("/api/clients", ClientsHandler).Methods("GET")
	router.HandleFunc("/api/clients/{id}", ClientHandler).Methods("GET")
	router.HandleFunc("/api/client", ClientCreateHandler).Methods("POST")
	router.HandleFunc("/api/client", ClientUpdateHandler).Methods("PUT")
	router.HandleFunc("/api/client", ClientDeleteHandler).Methods("DELETE")
	log.Fatal(http.ListenAndServe(":2020", router))
}

func Index(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(w, "Welcome to my API")
}

func ClientDeleteHandler(w http.ResponseWriter, r *http.Request) {
	var client Client
	body, err := ioutil.ReadAll(io.LimitReader(r.Body, 1048576))
	if err != nil {
		log.Fatalf("Error reading input: ", err)
	}

	if err := r.Body.Close(); err != nil {
		log.Fatalf("Error closing body: ", err)
	}

	if err := json.Unmarshal(body, &client); err != nil {
		w.Header().Set("Content-Type", "application/json; charset=UTF-8")
		w.WriteHeader(422) // unprocessable entity
		if err := json.NewEncoder(w).Encode(err); err != nil {
			log.Fatalf("Error encoding json: ", err)
		}
	}

	_, err = db.Exec("DELETE FROM clients WHERE id = $1", client.Id)
	if err != nil {
		log.Fatalf("Could not execute DELETE query: ", err)
	}

	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(client); err != nil {
		log.Fatalf("Could not encode json: ", err)
	}
}

func ClientUpdateHandler(w http.ResponseWriter, r *http.Request) {
	var client Client
	body, err := ioutil.ReadAll(io.LimitReader(r.Body, 1048576))
	if err != nil {
		log.Fatalf("Error reading input: ", err)
	}

	if err := r.Body.Close(); err != nil {
		log.Fatalf("Error closing body: ", err)
	}

	if err := json.Unmarshal(body, &client); err != nil {
		w.Header().Set("Content-Type", "application/json; charset=UTF-8")
		w.WriteHeader(422) // unprocessable entity
		if err := json.NewEncoder(w).Encode(err); err != nil {
			log.Fatalf("Error encoding json: ", err)
		}
	}

	_, err = db.Exec("UPDATE clients SET name = $1, email = $2, company = $3, country = $4 WHERE id = $5",
		client.Name, client.Email, client.Company, client.Country, client.Id)
	if err != nil {
		log.Fatalf("Could not execute UPDATE query: ", err)
	}

	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(client); err != nil {
		log.Fatalf("Could not encode json: ", err)
	}
}

func ClientCreateHandler(w http.ResponseWriter, r *http.Request) {
	var client Client
	body, err := ioutil.ReadAll(io.LimitReader(r.Body, 1048576))
	if err != nil {
		log.Fatalf("Error reading input: ", err)
	}

	if err := r.Body.Close(); err != nil {
		log.Fatalf("Error closing body: ", err)
	}

	if err := json.Unmarshal(body, &client); err != nil {
		w.Header().Set("Content-Type", "application/json; charset=UTF-8")
		w.WriteHeader(422) // unprocessable entity
		if err := json.NewEncoder(w).Encode(err); err != nil {
			log.Fatalf("Error encoding json: ", err)
		}
	}

	_ = db.QueryRow("INSERT INTO clients (name,email,company,country) VALUES ($1, $2, $3, $4) RETURNING id",
		client.Name, client.Email, client.Company, client.Country).Scan(&client.Id)
	if err != nil {
		log.Fatalf("Could not execute INSERT query: ", err)
	}

	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(http.StatusCreated)
	if err := json.NewEncoder(w).Encode(client); err != nil {
		log.Fatalf("Could not encode json: ", err)
	}
}

func ClientHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	var client Client
	err := db.QueryRow("SELECT * FROM clients WHERE id = $1", id).Scan(&client.Id, &client.Name, &client.Email, &client.Company, &client.Country)
	if err != nil {
		log.Fatalf("Error fetching client: %v", err)
	}

	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	if err = json.NewEncoder(w).Encode(client); err != nil {
		log.Fatalf("Could not encode json: %v", err)
	}
}

func ClientsHandler(w http.ResponseWriter, r *http.Request) {

	rows, err := db.Query("SELECT id,name,email,company,country FROM clients")
	if err != nil {
		log.Fatalf("Error selecting clients: %v", err)
	}

	clients := []Client{}
	for rows.Next() {
		var client Client
		err := rows.Scan(&client.Id, &client.Name, &client.Email, &client.Company, &client.Country)
		if err != nil {
			log.Fatalf("Error: %v", err)
		}
		clients = append(clients, client)
	}

	if err = rows.Err(); err != nil {
		log.Fatalf("Error parsing rows: %v", err)
	}

	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	if err = json.NewEncoder(w).Encode(clients); err != nil {
		log.Fatalf("Could not encode json: %v", err)
	}
}
