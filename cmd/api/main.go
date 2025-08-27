package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/segmentio/kafka-go"
)

type Event struct {
	UserID int    `json:"user_id"`
	Action string `json:"action"`
	TS     int64  `json:"ts"`
}

func main() {
	broker := os.Getenv("KAFKA_BROKER_ADDR")
	if broker == "" {
		broker = "kafka:9092"
	}
	writer := kafka.NewWriter(kafka.WriterConfig{
		Brokers:      []string{broker},
		Topic:        "events",
		RequiredAcks: int(kafka.RequireOne),
		Balancer:     &kafka.LeastBytes{},
	})

	mux := http.NewServeMux()
	mux.HandleFunc("/v1/publish", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		defer r.Body.Close()
		var ev Event
		if err := json.NewDecoder(r.Body).Decode(&ev); err != nil {
			http.Error(w, "bad json: "+err.Error(), http.StatusBadRequest)
			return
		}
		ev.TS = time.Now().Unix()
		b, _ := json.Marshal(ev)

		msg := kafka.Message{
			Key:   []byte((time.Now().Format("20060102150405"))),
			Value: b,
			Time:  time.Now(),
		}
		ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
		defer cancel()
		if err := writer.WriteMessages(ctx, msg); err != nil {
			http.Error(w, "kafka write failed: "+err.Error(), http.StatusBadGateway)
			return
		}
		w.WriteHeader(http.StatusAccepted)
		w.Write([]byte(`{"status":"queued"}`))
	})

	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	log.Println("api listening on :8080")
	log.Fatal(http.ListenAndServe(":8080", mux))
}
