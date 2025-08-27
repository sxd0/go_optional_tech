package main

import (
	"context"
	"encoding/json"
	"log"
	"os"
	"strconv"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/segmentio/kafka-go"
)

type Event struct {
	UserID int    `json:"user_id"`
	Action string `json:"action"`
	TS     int64  `json:"ts"`
}

func env(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

func main() {
	broker := env("KAFKA_BROKER_ADDR", "kafka:9092")
	redisAddr := env("REDIS_ADDR", "redis:6379")

	rdb := redis.NewClient(&redis.Options{Addr: redisAddr})

	rd := kafka.NewReader(kafka.ReaderConfig{
		Brokers:              []string{broker},
		GroupID:              "aggregators",
		GroupTopics:          []string{"events"},
		WatchPartitionChanges: true,

		MinBytes:       1,
		MaxBytes:       10e6,
		MaxWait:        250 * time.Millisecond,
		CommitInterval: time.Second,
		Logger:         log.New(os.Stdout, "kafka-reader ", 0),
		ErrorLogger:    log.New(os.Stderr, "kafka-reader-err ", 0),
	})


	log.Printf("aggregator started; consuming from Kafka… broker=%s redis=%s\n", broker, redisAddr)

	ctx := context.Background()
	for {
		m, err := rd.ReadMessage(ctx)
		if err != nil {
			log.Printf("kafka read error: %v", err)
			time.Sleep(time.Second)
			continue
		}

		log.Printf("got msg topic=%s partition=%d offset=%d key=%q val=%s",
			m.Topic, m.Partition, m.Offset, string(m.Key), string(m.Value))

		var ev Event
		if err := json.Unmarshal(m.Value, &ev); err != nil {
			log.Printf("bad event json: %v", err)
			continue
		}

		if err := rdb.HIncrBy(ctx, "stats:action", ev.Action, 1).Err(); err != nil {
			log.Printf("redis hincrby err: %v", err)
		}
		if err := rdb.Set(ctx, "last_action:"+strconv.Itoa(ev.UserID), ev.Action, 0).Err(); err != nil {
			log.Printf("redis set err: %v", err)
		}
	}
}
