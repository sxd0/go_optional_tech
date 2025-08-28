package main

import (
	"context"
	"encoding/json"
	"log"
	"net"
	"net/http"
	"os"
	"strings"

	"github.com/redis/go-redis/v9"
	"google.golang.org/grpc"

	"google.golang.org/grpc/reflection"
	"go_optional_tech/internal/report/grpcserver"
	statsv1 "go_optional_tech/proto/gen/go/stats/v1"
)

func main() {
	redisAddr := os.Getenv("REDIS_ADDR")
	if redisAddr == "" {
		redisAddr = "redis:6379"
	}

	rdb := redis.NewClient(&redis.Options{Addr: redisAddr})

	// --- HTTP (:8081) как раньше ---
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
		if uid == "" {
			http.Error(w, "user id required", http.StatusBadRequest)
			return
		}
		val, err := rdb.Get(context.Background(), "last_action:"+uid).Result()
		if err == redis.Nil {
			http.Error(w, "no data", http.StatusNotFound)
			return
		}
		if err != nil {
			http.Error(w, "redis error: "+err.Error(), http.StatusBadGateway)
			return
		}
		_ = json.NewEncoder(w).Encode(map[string]string{
			"user_id":     uid,
			"last_action": val,
		})
	})
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	go func() {
		log.Println("report HTTP listening on :8081")
		log.Fatal(http.ListenAndServe(":8081", mux))
	}()

	// --- gRPC (:9090) ---
	gs := grpc.NewServer()
	statsv1.RegisterStatsServiceServer(gs, grpcserver.New(rdb))
	reflection.Register(gs)

	lis, err := net.Listen("tcp", ":9090")
	if err != nil {
		log.Fatalf("grpc listen: %v", err)
	}
	log.Println("report gRPC listening on :9090")
	log.Fatal(gs.Serve(lis))
}
