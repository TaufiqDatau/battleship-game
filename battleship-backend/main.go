package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/gorilla/websocket"
)

type Shot struct {
	Action string `json:"action"`
	Row    int    `json:"row"`
	Col    int    `json:"col"`
}

type SentData struct {
	Condition string `json:"condition"`
	Row       int    `json:"row"`
	Col       int    `json:"col"`
}

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

var enemyBoard = [10][10]int{
	{0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
	{0, 0, 0, 1, 0, 0, 0, 0, 0, 0},
	{0, 0, 0, 1, 0, 0, 0, 1, 1, 1},
	{0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
	{0, 1, 1, 1, 1, 0, 0, 0, 0, 0},
	{0, 0, 0, 0, 0, 0, 1, 1, 0, 0},
	{0, 0, 0, 0, 1, 0, 0, 0, 1, 0},
	{0, 0, 0, 0, 1, 0, 0, 0, 1, 0},
	{0, 0, 0, 0, 1, 0, 0, 0, 1, 0},
	{0, 0, 0, 0, 1, 0, 0, 0, 1, 0},
}

func homePage(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Home Page")
}

func wsEndpoint(w http.ResponseWriter, r *http.Request) {
	upgrader.CheckOrigin = func(r *http.Request) bool { return true }

	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		fmt.Println(err)
	}

	log.Println("Client Connected")

	reader(ws)
}

func reader(conn *websocket.Conn) {
	for {
		// Read message from client
		_, message, err := conn.ReadMessage()
		if err != nil {
			log.Println("Error reading message:", err)
			return
		}

		// Convert message to string
		messageStr := string(message)
		fmt.Printf("Received message: %s\n", messageStr)

		// Try to unmarshal the message into a Shot struct
		var shot Shot
		err = json.Unmarshal(message, &shot)
		if err != nil {
			log.Println("Error parsing JSON:", err)
			continue // If it's not a valid shot JSON, keep waiting for the next message
		}

		// Handle the action (fire shot in this case)
		if shot.Action == "fire_shot" {
			log.Printf("Firing shot at row %d, col %d\n", shot.Row, shot.Col)

			condition := "miss"
			if enemyBoard[shot.Row][shot.Col] == 1 {
				condition = "hit"
			}

			var responseData SentData = SentData{Condition: condition, Row: shot.Row, Col: shot.Col}
			fmt.Println(responseData)
			responseJSON, err := json.Marshal(responseData)

			if err != nil {
				log.Println("Error marshalling JSON:", err)
				continue
			}

			if err := conn.WriteMessage(websocket.TextMessage, responseJSON); err != nil {
				log.Println("Error sending message:", err)
				return
			}
		}

	}

}

func setupRoutes() {
	http.HandleFunc("/", homePage)
	http.HandleFunc("/ws", wsEndpoint)
}

func main() {
	fmt.Println("Hello World")
	setupRoutes()
	log.Fatal(http.ListenAndServe(":8080", nil))
}