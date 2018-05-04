package main

import (
	"log"
	"net/http"

	"github.com/gorilla/websocket"
)

var clients = make(map[*websocket.Conn]bool) // connected clients
var broadcast = make(chan Message)           // broadcast channel

// configure an upgrader
var upgrader = websocket.Upgrader{}

// Message object
type Message struct {
	Email    string `json:"email"`
	Username string `json:"username"`
	Message  string `json:"message"`
}

// handle connections to upgrade GETS to websocket
func handleConnections(w http.ResponseWriter, r *http.Request) {
	// Upgrade initial GET request to a websocket
	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Fatal(err)
	}
	defer ws.Close()
	// Register a new client
	clients[ws] = true

	for {
		var msg Message
		// Read in a new message as JSON and map it to a Message object
		err := ws.ReadJSON(&msg)
		if err != nil {
			log.Printf("error: %v", err)
			delete(clients, ws)
			break
		}
		broadcast <- msg
	}
}

// handle messages by listening on broadcast channel
func handleMessages() {
	msg := <-broadcast
	// Send it out to every client currently connected
	for client := range clients {
		err := client.WriteJSON(msg)
		if err != nil {
			log.Printf("error: %v", err)
			client.Close()
			delete(clients, client)
		}
	}
}

func main() {
	// create a file server
	fs := http.FileServer(http.Dir("../public"))
	http.Handle("/", fs)

	// Configre the websockt route
	http.HandleFunc("/ws", handleConnections)

	// start listening to incoming messages
	go handleMessages()

	// Start the server on localhost port 8000 and log and errors
	// port := 8000
	log.Println("http server started on port:8000")
	err := http.ListenAndServe(":8000", nil)
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}
