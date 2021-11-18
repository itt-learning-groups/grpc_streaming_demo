package main

import (
	"context"
	"fmt"
	"io"

	"github.com/sirupsen/logrus"
	namesservice "gitlab.gs.kount.com/kount/Luke/daily_notes/2021/learnanddevops/grpc_servers/stream_server/pkg/namesvc/gen/go"
	"google.golang.org/grpc"
)

func main() {
	log := logrus.New()

	dialer, err := grpc.Dial("0.0.0.0:50051", grpc.WithInsecure())
	if err != nil {
		log.Fatalf("Failed to dial names service: %v", err)
	}

	client := namesservice.NewNamesServiceClient(dialer)

	gnr := namesservice.GetNamesRequest{}

	stream, err := client.GetNames(context.Background(), &gnr)
	if err != nil {
		log.Errorf("failed to get names: %v", err)
	}

	names := []string{}

	for {
		n, err := stream.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Errorf("name error: %v", err)
		}
		names = append(names, n.Name)
	}

	if len(names) > 0 {
		fmt.Println("Got Names:")
		for _, nm := range names {
			fmt.Println(nm)
		}
	}
}

// func main() {
// 	log := logrus.New()

// 	client, err := grpc.Dial("0.0.0.0:50051", grpc.WithInsecure())
// 	if err != nil {
// 		log.Fatalf("Failed to dial names service: %v", err)
// 	}

// 	namesClient := namesservice.NewNamesServiceClient(client)

// 	stream, err := namesClient.GetNames(context.Background(), &namesservice.GetNamesRequest{})
// 	if err != nil {
// 		log.Errorf("failed to get names: %v", err)
// 	}

// 	names := []string{}

// 	for {
// 		n, err := stream.Recv()
// 		if err == io.EOF {
// 			break
// 		}
// 		if err != nil {
// 			log.Errorf("name error: %v", err)
// 		}
// 		names = append(names, n.Name)
// 	}
// 	if len(names) > 0 {
// 		fmt.Println("Got Names:")
// 		for _, nm := range names {
// 			fmt.Println(nm)
// 		}
// 	}
// }
