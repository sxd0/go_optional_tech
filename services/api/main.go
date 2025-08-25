package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"time"

	kafka "github.com/segmentio/kafka-go"
)

type Event struct {
	Type    string                 `json:"type"`
	Payload map[string]interface{} `json:"payload"`
}

func getenv(key, def string) string {
	if v, ok := os.LookupEnv(key); ok && v != "" {
		return v
	}
	return def
}

func main() {
	broker := getenv("KAFKA_BROKER", "kafka:9092")
	topic := getenv("KAFKA_TOPIC", "events")
	addr := getenv("HTTP_ADDR", ":8080")

	writer := &kafka.Writer{
		Addr:                   kafka.TCP(broker),
		Topic:                  topic,
		Balancer:               &kafka.LeastBytes{},
		RequiredAcks:           kafka.RequireAll,
		AllowAutoTopicCreation: true,
	}

	http.HandleFunc("/events", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		defer r.Body.Close()

		var e Event
		if err := json.NewDecoder(r.Body).Decode(&e); err != nil {
			http.Error(w, "bad json: "+err.Error(), http.StatusBadRequest)
			return
		}
		if e.Type == "" {
			http.Error(w, "missing field: type", http.StatusBadRequest)
			return
		}
		body, _ := json.Marshal(e)

		ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
		defer cancel()
		if err := writer.WriteMessages(ctx, kafka.Message{Value: body}); err != nil {
			log.Println("kafka write failed:", err)
			http.Error(w, "kafka write failed: "+err.Error(), http.StatusBadGateway)
			return
		}
		log.Printf("published event type=%q size=%d", e.Type, len(body))
		w.WriteHeader(http.StatusAccepted)
	})

	log.Println("api listening on", addr, "→ Kafka", broker, "topic", topic)
	log.Fatal(http.ListenAndServe(addr, nil))
}
