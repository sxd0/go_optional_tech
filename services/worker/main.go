package main

import (
	"context"
	"encoding/json"
	"log"
	"os"
	"time"

	kafka "github.com/segmentio/kafka-go"
	"github.com/redis/go-redis/v9"
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
	group := getenv("KAFKA_GROUP", "mini-worker")
	redisAddr := getenv("REDIS_ADDR", "redis:6379")
	keepN := 100

	rdb := redis.NewClient(&redis.Options{Addr: redisAddr})
	ctx := context.Background()

	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers:               []string{broker},
		GroupID:               group,
		GroupTopics:           []string{topic},
		WatchPartitionChanges: true,
		CommitInterval:        time.Second,
		MinBytes:              1,
		MaxBytes:              10e6,
		Logger:                log.New(os.Stdout, "kafka-reader ", log.LstdFlags),
		ErrorLogger:           log.New(os.Stderr, "kafka-reader-err ", log.LstdFlags),
	})
	defer reader.Close()

	log.Println("worker consuming", topic, "from", broker, "as group", group, "→ redis", redisAddr)

	for {
		m, err := reader.ReadMessage(ctx)
		if err != nil {
			log.Println("kafka read error:", err)
			time.Sleep(time.Second)
			continue
		}
		var e Event
		if err := json.Unmarshal(m.Value, &e); err != nil {
			log.Println("bad event json:", err)
			continue
		}
		log.Printf("consumed offset=%d type=%q", m.Offset, e.Type)

		typ := e.Type
		if typ == "" {
			typ = "unknown"
		}
		if err := rdb.Incr(ctx, "events:count:"+typ).Err(); err != nil {
			log.Println("redis INCR error:", err)
		} else {
			log.Printf("incr ok key=events:count:%s", typ)
		}

		raw := string(m.Value)
		if err := rdb.LPush(ctx, "events:recent", raw).Err(); err != nil {
			log.Println("redis LPUSH error:", err)
		} else {
			log.Printf("lpush ok key=events:recent")
		}
		if err := rdb.LTrim(ctx, "events:recent", 0, int64(keepN-1)).Err(); err != nil {
			log.Println("redis LTRIM error:", err)
		}
	}
}
