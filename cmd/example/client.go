package main

import (
	"context"
	"fmt"
	"os"

	"github.com/datskos/ratelimit/pkg/proto"
	"google.golang.org/grpc"
)

func main() {
	conn, err := grpc.Dial("localhost:8080", grpc.WithInsecure())
	if err != nil {
		fmt.Printf("error establishing connection: %s\n", err)
		os.Exit(-1)
	}

	defer conn.Close()
	client := proto.NewRateLimitServiceClient(conn)
	req := &proto.ReduceRequest{
		Key:               "sms:543",
		MaxAmount:         3,
		RefillAmount:      1,
		RefillDurationSec: 60,
	}
	resp, err := client.Reduce(context.Background(), req)
	if err != nil {
		fmt.Printf("error executing reduce command %s\n", err)
		os.Exit(-1)
	}

	fmt.Println("got resp", resp)
}
