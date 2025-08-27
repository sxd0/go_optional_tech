package main

import (
	"encoding/json"
	"log"
	"net/http"
	"strings"

	"github.com/redis/go-redis/v9"
	"golang.org/x/net/context"
)

func main() {
	rdb := redis.NewClient(&redis.Options{Addr: "redis:6379"})

	mux := http.NewServeMux()
	mux.HandleFunc("/v1/stats", func(w http.ResponseWriter, r *http.Request) {
		m, err := rdb.HGetAll(context.Background(), "stats:action").Result()
		if err != nil {
			http.Error(w, "redis error: "+err.Error(), http.StatusBadGateway)
			return
		}
		_ = json.NewEncoder(w).Encode(m)
	})
	mux.HandleFunc("/v1/user/", func(w http.ResponseWriter, r *http.Request) {
		uid := strings.TrimPrefix(r.URL.Path, "/v1/user/")
		val, err := rdb.Get(context.Background(), "last_action:"+uid).Result()
		if err == redis.Nil {
			http.Error(w, "no data", http.StatusNotFound)
			return
		}
		if err != nil {
			http.Error(w, "redis error: "+err.Error(), http.StatusBadGateway)
			return
		}
		_ = json.NewEncoder(w).Encode(map[string]string{"user_id": uid, "last_action": val})
	})

	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	log.Println("report listening on :8081")
	log.Fatal(http.ListenAndServe(":8081", mux))
}
