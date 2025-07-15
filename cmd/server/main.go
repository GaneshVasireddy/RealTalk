package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/GaneshVasireddy/RealTalk/config"
	"github.com/GaneshVasireddy/RealTalk/repository/mongo"
)

type User struct {
	Id string `json:"id"`
}

type Message struct {
	Body string `json:"body"`
}

type Event struct {
	User    User    `json:"user"`
	Message Message `json:"message"`
}

type Client struct {
	Id     string
	Writer http.ResponseWriter
}

func initializeServer(config *config.Config) {

	// Initialize the MongoDB connection
	mongoClient := mongo.Connect(&config.Mongo)
	if mongoClient == nil {
		fmt.Println("Failed to connect to MongoDB")
		return
	}

	clients := make(map[string]map[string]map[string]Client)

	http.HandleFunc("/api/v1/channel/{channel_id}/messages", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			fmt.Println("posting a message")

			channelId := r.PathValue("channel_id")
			if channelId == "" {
				http.Error(w, "channel_id is required", http.StatusBadRequest)
				return
			}

			var event Event
			if err := json.NewDecoder(r.Body).Decode(&event); err != nil {
				http.Error(w, "Invalid request body", http.StatusBadRequest)
				return
			}
			BroadcastMessage(event, clients[channelId])
		}
	})

	http.HandleFunc("/api/v1/events", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			fmt.Println("streaming events")

			w.Header().Set("Content-Type", "text/event-stream")
			w.Header().Set("Cache-Control", "no-cache")
			w.Header().Set("Connection", "keep-alive")

			if r.URL.Query().Get("channel_id") == "" || r.URL.Query().Get("user_id") == "" || r.URL.Query().Get("session_id") == "" {
				http.Error(w, "channel_id, user_id and session_id are required", http.StatusBadRequest)
				return
			}

			channelId := r.URL.Query().Get("channel_id")
			userId := r.URL.Query().Get("user_id")
			sessionId := r.URL.Query().Get("session_id")

			fmt.Printf("Client connected: channel_id: %s, user_id: %s and session_id: %s", channelId, userId, sessionId)

			if clients[channelId] == nil {
				clients[channelId] = make(map[string]map[string]Client)
			}

			if clients[channelId][userId] == nil {
				clients[channelId][userId] = make(map[string]Client)
			}

			clients[channelId][userId][sessionId] = Client{
				Id:     userId,
				Writer: w,
			}

			select {
				case <-r.Context().Done():
					fmt.Println("Client disconnected")
					delete(clients[channelId][userId], r.URL.Query()["session_id"][0])
					return
			}
		}
	})
}

func BroadcastMessage(event Event, client map[string]map[string]Client) {
	for _, users := range client {
		for _, c := range users {
			w := c.Writer
			flusher, ok := w.(http.Flusher)
			if !ok {
				http.Error(w, "Streaming unsupported", http.StatusInternalServerError)
				return
			}
			fmt.Fprintf(w, "data: %+v\n", event)
			flusher.Flush() 
		}
	}
}
	

// main is the entry point of the application.
func main() {
	fmt.Println("Starting RealTalk server...")
	config := config.NewConfig()

	// Initialize the server with the configuration
	initializeServer(config)

	// Start the server
	server := &http.Server{
		Addr: fmt.Sprintf(":%d", config.Port),
	}

	go func() {
		fmt.Printf("Server is listening on port %d\n", config.Port)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			fmt.Println("Error starting server:", err)
			return
		}
	}()

	sigChann := make(chan os.Signal, 1)

	signal.Notify(sigChann, syscall.SIGINT, syscall.SIGTERM)
	<-sigChann

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	fmt.Println("Shutting down server...")
	if err := server.Shutdown(ctx); err != nil {
		fmt.Println("Error shutting down server:", err)
		return
	}
	fmt.Println("Server shut down gracefully")
}
