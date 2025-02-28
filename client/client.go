package main

import (
	"bufio"
	"flag"
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/cauakath/chat/model"
	"github.com/gorilla/websocket"
)

const (
	SERVER_URL = "ws://localhost:3030/ws"
)

var (
	received_messages = make(chan model.Message)
)

func main() {
	clientIdPtr := flag.String("client", "uuid", "the client id")

	flag.Parse()

	if *clientIdPtr == "" {
		log.Println("You need to pass a client as argument!")
		return
	}

	conn, err := connectWs(*clientIdPtr)
	if err != nil {
		log.Println(err)
		return
	}

	defer conn.Close()

	sigc := make(chan os.Signal, 1)

	go handleReceivedMessages(conn, sigc)

	signal.Notify(sigc, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-sigc
		closeSession(conn, *clientIdPtr)
		os.Exit(1)
	}()

	for {
		sendMessage(conn, *clientIdPtr)
	}
}

func connectWs(clientId string) (*websocket.Conn, error) {
	log.Printf("\nConnecting to %s with clientId: %s", SERVER_URL, clientId)

	conn, _, err := websocket.DefaultDialer.Dial(SERVER_URL, nil)
	if err != nil {
		return &websocket.Conn{}, err
	}

	err = conn.WriteJSON(model.Message{
		Sender:    clientId,
		Receiver:  "server",
		Text:      clientId,
		Type:      model.NewClient,
		Timestamp: time.Now(),
	})

	if err != nil {
		log.Println("Error on connect to server")
		return &websocket.Conn{}, err
	}

	return conn, nil
}

func sendMessage(conn *websocket.Conn, clientId string) {
	scanner := bufio.NewScanner(os.Stdin)
	scanner.Scan()
	msg := scanner.Text()

	parts := strings.SplitN(msg, " ", 2)
	var recipientID string
	if len(parts) > 1 && strings.HasPrefix(parts[0], "to:") {
		recipientID = strings.TrimPrefix(parts[0], "to:")
		msg = parts[1]
	}

	if recipientID == "" {
		log.Println("Recipient ID not provided. Please include recipient ID in the message 'to:<id> >message>'.")
		return
	}

	err := conn.WriteJSON(model.Message{
		Sender:    clientId,
		Receiver:  recipientID,
		Text:      msg,
		Type:      model.Chat,
		Timestamp: time.Now(),
	})

	if err != nil {
		log.Println("Error on send message to recipientId")
		return
	}
}

func handleReceivedMessages(conn *websocket.Conn, sigc chan os.Signal) {
	for {
		var msg model.Message
		err := conn.ReadJSON(&msg)
		if err != nil {
			log.Println("Error reading message")
			return
		}

		switch msg.Type {
		case model.Chat:
			log.Printf("\n%s %s: %s", formatTimestamp(msg.Timestamp), msg.Sender, msg.Text)
		case model.Error:
			log.Printf("\n%s Error from server: %s", formatTimestamp(msg.Timestamp), msg.Text)

			if msg.Text == string(model.ErrorUserAlreadyExists) {
				sigc <- os.Interrupt
			}
		}
	}
}

func formatTimestamp(time time.Time) string {
	return time.Format("2006-01-02 15:04:05")
}

func closeSession(conn *websocket.Conn, clientId string) {
	if err := conn.WriteJSON(model.Message{
		Sender:    clientId,
		Receiver:  "server",
		Type:      model.EndSession,
		Text:      "Session end",
		Timestamp: time.Now(),
	}); err != nil {
		log.Println("Error to send end session message to server")
	}
}
