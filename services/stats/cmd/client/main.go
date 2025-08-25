package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"google.golang.org/grpc"
	"go_optional_tech/proto/statspb"
)

func main() {
	conn, err := grpc.Dial("localhost:9090", grpc.WithInsecure())
	if err != nil { log.Fatal(err) }
	defer conn.Close()

	c := statspb.NewStatsServiceClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	s, err := c.GetStats(ctx, &statspb.Empty{})
	if err != nil { log.Fatal("GetStats:", err) }
	fmt.Println("Stats:")
	for _, it := range s.Items {
		fmt.Printf("  %s = %d\n", it.Key, it.Value)
	}

	r, err := c.GetRecent(ctx, &statspb.Empty{})
	if err != nil { log.Fatal("GetRecent:", err) }
	fmt.Println("Recent:")
	for _, e := range r.Events {
		fmt.Println(" ", e)
	}
}
