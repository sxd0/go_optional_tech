package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"sort"

	"github.com/redis/go-redis/v9"
)

func getenv(key, def string) string {
	if v, ok := os.LookupEnv(key); ok && v != "" {
		return v
	}
	return def
}

func main() {
	redisAddr := getenv("REDIS_ADDR", "redis:6379")
	addr := getenv("HTTP_ADDR", ":8082")

	rdb := redis.NewClient(&redis.Options{Addr: redisAddr})
	ctx := context.Background()

	http.HandleFunc("/stats", func(w http.ResponseWriter, r *http.Request) {
		iter := rdb.Scan(ctx, 0, "events:count:*", 200).Iterator()
		type kv struct{ Key string; Val int64 }
		items := make([]kv, 0) 
		for iter.Next(ctx) {
			key := iter.Val()
			if val, err := rdb.Get(ctx, key).Int64(); err == nil {
				items = append(items, kv{Key: key, Val: val})
			}
		}
		sort.Slice(items, func(i, j int) bool { return items[i].Key < items[j].Key })
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(items)
	})

	http.HandleFunc("/events", func(w http.ResponseWriter, r *http.Request) {
		evs, _ := rdb.LRange(ctx, "events:recent", 0, 19).Result()
		if evs == nil {
			evs = []string{} 
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(evs)
	})


	log.Println("stats listening on", addr, "← Redis", redisAddr)
	log.Fatal(http.ListenAndServe(addr, nil))
}
