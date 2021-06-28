package main

import (
	"encoding/json"
	"fmt"
	"go-proto-tags/pb"
	"io"
	"log"
	"net/http"
	"time"

	"google.golang.org/grpc"
)

func main() {

	conn, err := grpc.Dial(":8888", grpc.WithInsecure())
	if err != nil {
		panic(err)
	}

	messengerClient := pb.NewMessengerClient(conn)

	api := NewMessengerApi(messengerClient)
	http.HandleFunc("/messages/new", POST(api.PostMessage))
	http.HandleFunc("/messages", GET(api.GetMessages))

	log.Println("Messenger API listening on :8889")

	log.Fatal(http.ListenAndServe(":8889", nil))
}

func GET(next http.HandlerFunc) http.HandlerFunc {
	return WithMethod(http.MethodGet, next)
}

func POST(next http.HandlerFunc) http.HandlerFunc {
	return WithMethod(http.MethodPost, next)
}

func WithMethod(method string, next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != method {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		next(w, r)
	}
}

type Client interface {
	PostMessage(http.ResponseWriter, *http.Request)
	GetMessages(http.ResponseWriter, *http.Request)
}

type client struct {
	messenger pb.MessengerClient
}

func NewMessengerApi(messenger pb.MessengerClient) Client {
	return &client{messenger: messenger}
}

func (c *client) PostMessage(w http.ResponseWriter, r *http.Request) {

	body, err := io.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusUnprocessableEntity)
		fmt.Fprintf(w, "%s", err.Error())
		return
	}

	var input pb.Message
	err = json.Unmarshal(body, &input)
	if err != nil {
		w.WriteHeader(http.StatusUnprocessableEntity)
		fmt.Fprintf(w, "%s", err.Error())
		return
	}

	resp, err := c.messenger.CreateMessage(r.Context(), &input)
	if err != nil {
		w.WriteHeader(http.StatusUnprocessableEntity)
		fmt.Fprintf(w, "%s", err.Error())
		return
	}

	w.WriteHeader(http.StatusCreated)

	message := &JsonMessage{}
	message.Fill(resp)

	_ = json.NewEncoder(w).Encode(message)
}

type JsonMessage struct {
	ID        string `json:"id"`
	Title     string `json:"title"`
	Body      string `json:"body"`
	Author    string `json:"author"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
}

func (jm *JsonMessage) Fill(m *pb.Message) {
	jm.ID = m.Id
	jm.Title = m.Title
	jm.Body = m.Body
	jm.Author = m.Author
	jm.CreatedAt = time.Unix(m.Created, 0).String()
	jm.UpdatedAt = time.Unix(m.Updated, 0).String()
}

func (c *client) GetMessages(w http.ResponseWriter, r *http.Request) {
	var messages []*JsonMessage

	stream, err := c.messenger.GetMessages(r.Context(), &pb.GetMessagesRequest{})
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, "%s", err.Error())
		return
	}

	for {

		resp, err := stream.Recv()

		if err == io.EOF {
			break
		}

		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			fmt.Fprintf(w, "%s", err.Error())
			return
		}

		message := &JsonMessage{}
		message.Fill(resp)

		messages = append(messages, message)
	}

	_ = json.NewEncoder(w).Encode(messages)
}
