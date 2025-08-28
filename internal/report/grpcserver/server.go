package grpcserver

import (
	"context"
	"fmt"

	"github.com/redis/go-redis/v9"
	statsv1 "go_optional_tech/proto/gen/go/stats/v1"
)

type Server struct {
	statsv1.UnimplementedStatsServiceServer
	rdb *redis.Client
}

func New(rdb *redis.Client) *Server {
	return &Server{rdb: rdb}
}

func (s *Server) GetStats(ctx context.Context, _ *statsv1.Empty) (*statsv1.GetStatsResponse, error) {
	m, err := s.rdb.HGetAll(ctx, "stats:action").Result()
	if err != nil {
		return nil, fmt.Errorf("redis HGetAll: %w", err)
	}
	return &statsv1.GetStatsResponse{Counts: m}, nil
}

func (s *Server) GetUserLastAction(ctx context.Context, req *statsv1.GetUserLastActionRequest) (*statsv1.GetUserLastActionResponse, error) {
	key := fmt.Sprintf("last_action:%d", req.GetUserId())
	val, err := s.rdb.Get(ctx, key).Result()
	if err == redis.Nil {
		return &statsv1.GetUserLastActionResponse{
			UserId:     fmt.Sprint(req.GetUserId()),
			LastAction: "",
		}, nil
	}
	if err != nil {
		return nil, fmt.Errorf("redis GET: %w", err)
	}
	return &statsv1.GetUserLastActionResponse{
		UserId:     fmt.Sprint(req.GetUserId()),
		LastAction: val,
	}, nil
}
