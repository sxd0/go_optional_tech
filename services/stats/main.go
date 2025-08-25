package main

import (
	"context"
	"encoding/json"
	"log"
	"net"
	"net/http"
	"os"
	"sort"

	"github.com/redis/go-redis/v9"
	"google.golang.org/grpc"

	"go_optional_tech/proto/statspb"
)

func getenv(key, def string) string {
	if v, ok := os.LookupEnv(key); ok && v != "" {
		return v
	}
	return def
}

type server struct {
	statspb.UnimplementedStatsServiceServer
	rdb *redis.Client
}

func (s *server) GetStats(ctx context.Context, _ *statspb.Empty) (*statspb.StatsResponse, error) {
	iter := s.rdb.Scan(ctx, 0, "events:count:*", 200).Iterator()
	items := make([]*statspb.StatItem, 0)
	for iter.Next(ctx) {
		key := iter.Val()
		if val, err := s.rdb.Get(ctx, key).Int64(); err == nil {
			items = append(items, &statspb.StatItem{Key: key, Value: val})
		}
	}
	sort.Slice(items, func(i, j int) bool { return items[i].Key < items[j].Key })
	return &statspb.StatsResponse{Items: items}, nil
}

func (s *server) GetRecent(ctx context.Context, _ *statspb.Empty) (*statspb.RecentResponse, error) {
	evs, _ := s.rdb.LRange(ctx, "events:recent", 0, 19).Result()
	if evs == nil {
		evs = []string{}
	}
	return &statspb.RecentResponse{Events: evs}, nil
}

func main() {
	redisAddr := getenv("REDIS_ADDR", "redis:6379")
	httpAddr := getenv("HTTP_ADDR", ":8082")
	grpcAddr := getenv("GRPC_ADDR", ":9090")

	rdb := redis.NewClient(&redis.Options{Addr: redisAddr})
	ctx := context.Background()

	// REST
	http.HandleFunc("/stats", func(w http.ResponseWriter, r *http.Request) {
		iter := rdb.Scan(ctx, 0, "events:count:*", 200).Iterator()
		type kv struct {
			Key string
			Val int64
		}
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

	// gRPC
	go func() {
		lis, err := net.Listen("tcp", grpcAddr)
		if err != nil {
			log.Fatalf("grpc listen err: %v", err)
		}
		gs := grpc.NewServer()
		statspb.RegisterStatsServiceServer(gs, &server{rdb: rdb})
		log.Println("stats gRPC listening on", grpcAddr, "← Redis", redisAddr)
		if err := gs.Serve(lis); err != nil {
			log.Fatalf("grpc serve err: %v", err)
		}
	}()

	log.Println("stats REST listening on", httpAddr, "← Redis", redisAddr)
	log.Fatal(http.ListenAndServe(httpAddr, nil))
}
