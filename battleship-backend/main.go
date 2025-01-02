package main

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"log"
	"net/http"
	"sync"

	"github.com/gorilla/websocket"
)

type GameRoom struct {
	ID         string
	Clients    map[*websocket.Conn]bool
	Broadcast  chan string
	mu         sync.Mutex
	MaxPlayers int
}

type Server struct {
	Rooms map[string]*GameRoom
	mu    sync.Mutex
}

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

// NewGameServer creates a new server
func NewGameServer() *Server {
	return &Server{
		Rooms: make(map[string]*GameRoom),
	}
}

// JoinRoom handles player connections to a room
func (s *Server) JoinRoom(w http.ResponseWriter, r *http.Request) {
	roomID := r.URL.Query().Get("room")
	if roomID == "" {
		// Generate a new room ID
		roomID = generateRoomID()
		http.Redirect(w, r, fmt.Sprintf("/join?room=%s", roomID), http.StatusFound)
		return
	}

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("Upgrade error:", err)
		return
	}

	s.mu.Lock()
	room, exists := s.Rooms[roomID]
	if !exists {
		room = &GameRoom{
			ID:         roomID,
			Clients:    make(map[*websocket.Conn]bool),
			Broadcast:  make(chan string),
			MaxPlayers: 2,
		}
		s.Rooms[roomID] = room
		go room.HandleMessages()
	}
	s.mu.Unlock()

	room.mu.Lock()
	defer room.mu.Unlock()
	if len(room.Clients) >= room.MaxPlayers {
		conn.WriteMessage(websocket.TextMessage, []byte("Room is full"))
		conn.Close()
		return
	}

	room.Clients[conn] = true
	log.Printf("Player joined room %s", roomID)

	go room.HandleClient(conn)
}

// HandleMessages broadcasts messages to all clients in the room
func (room *GameRoom) HandleMessages() {
	for {
		message := <-room.Broadcast
		room.mu.Lock()
		for client := range room.Clients {
			log.Println()
			err := client.WriteMessage(websocket.TextMessage, []byte(message))
			if err != nil {
				log.Println("Write error:", err)
				client.Close()
				delete(room.Clients, client)
			}
		}
		room.mu.Unlock()
	}
}

// HandleClient processes messages from a client
func (room *GameRoom) HandleClient(conn *websocket.Conn) {
	defer func() {
		room.mu.Lock()
		delete(room.Clients, conn)
		room.mu.Unlock()
		conn.Close()
	}()

	for {
		_, msg, err := conn.ReadMessage()
		if err != nil {
			log.Println("Read error:", err)
			return
		}

		room.Broadcast <- fmt.Sprintf("Room %s: %s", room.ID, string(msg))
	}
}

func generateRoomID() string {
	bytes := make([]byte, 4) // 4 bytes = 8 hex characters
	_, err := rand.Read(bytes)
	if err != nil {
		log.Println("Error generating room ID:", err)
		return "default-room" // Fallback room ID
	}
	return hex.EncodeToString(bytes)
}

func main() {
	server := NewGameServer()

	http.HandleFunc("/join", server.JoinRoom)

	log.Println("Server started on :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
