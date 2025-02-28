package main

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/cauakath/chat/model"
	"github.com/gorilla/websocket"
)

var (
	upgrader = websocket.Upgrader{
		WriteBufferSize: 1024,
		ReadBufferSize:  1024,
	}
	client_pool  = make(map[string]*websocket.Conn)
	chat_history = make(chan model.Message)
)

func main() {
	mux := http.NewServeMux()

	mux.HandleFunc("/ws", handleConnections)

	log.Println("Server initializing on port 3030")

	go handleMessages()

	err := http.ListenAndServe(":3030", mux)
	if err != nil {
		log.Fatal("Error on start server: ", err)
	}
}

func handleConnections(w http.ResponseWriter, h *http.Request) {
	conn, err := upgrader.Upgrade(w, h, nil)
	if err != nil {
		log.Println(err)
		return
	}

	defer conn.Close()

	for {
		var msg model.Message
		err := conn.ReadJSON(&msg)
		if err != nil {
			log.Println(err)
			return
		}

		switch msg.Type {
		case model.NewClient:
			err := addNewClient(msg, conn)
			if err != nil {
				log.Println(err)
				return
			}
		case model.Chat:
			chat_history <- msg
		case model.EndSession:
			err := removeClient(msg)
			if err != nil {
				log.Println(err)
				return
			}
			break
		default:
			log.Println("Unknown message:", msg.Type)
		}
	}
}

func handleMessages() {
	for msg := range chat_history {
		receiver, exists := client_pool[msg.Receiver]
		if !exists {
			log.Printf("\nReceiver %s is not connected", msg.Receiver)
			client_pool[msg.Sender].WriteJSON(model.Message{
				Sender:    "server",
				Receiver:  msg.Sender,
				Text:      fmt.Sprintf("%s is not connected on server", msg.Receiver),
				Type:      model.Error,
				Timestamp: time.Now(),
			})
			continue
		}

		receiver.WriteJSON(msg)
	}
}

func addNewClient(msg model.Message, conn *websocket.Conn) error {
	clientId := msg.Text

	if _, exists := client_pool[clientId]; exists {
		log.Println("User already exists:", clientId)
		err := conn.WriteJSON(model.Message{
			Sender:    "server",
			Receiver:  clientId,
			Text:      string(model.ErrorUserAlreadyExists),
			Type:      model.Error,
			Timestamp: time.Now(),
		})
		if err != nil {
			log.Println("Error on send error message:", err)
			return err
		}
	} else {
		client_pool[clientId] = conn
		log.Println("New user added:", clientId)
	}

	return nil
}

func removeClient(msg model.Message) error {
	clientId := msg.Sender

	log.Printf("\nRemoving user %s from pool", clientId)
	delete(client_pool, clientId)

	return nil
}
