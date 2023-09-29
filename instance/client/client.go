package main

import (
	"context"
	"flag"
	"io"
	"log"
	"sync"

	pb "github.com/MelihEmreGuler/gRPC-streaming-service/instance/instancepb"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

var (
	addr   = flag.String("addr", "localhost:50051", "the address to connect to")
	region = flag.String("region", "us-east-1", "the region to query")
)

func RegionRequest() *pb.GetInstancesByRegionRequest {
	return &pb.GetInstancesByRegionRequest{Region: *region}
}

func main() {
	// Create a connection to the server
	flag.Parse()
	conn, err := grpc.Dial(*addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("fail to dial: %v", err)
	}
	defer conn.Close()
	client := pb.NewInstanceClient(conn)
	// Create a context with a timeout
	ctx := context.Background()

	// WaitGroup to wait for all goroutines to finish
	var wg sync.WaitGroup

	// Create a goroutine for GetInstancesByRegion
	wg.Add(1)
	go func() {
		defer wg.Done()
		resp, err := client.GetInstancesByRegion(ctx, RegionRequest())
		if err != nil {
			log.Fatalf("error calling GetInstancesByRegion: %v", err)
		}
		// Receive and process instance responses
		for {
			instance, err := resp.Recv()
			if err == io.EOF {
				break
			}
			if err != nil {
				log.Fatalf("error receiving instance data: %v", err)
			}
			// Print or process the received instance data
			log.Printf("Received instance data: %v\n", instance)
		}
	}()

	// Create a goroutine for SendStatusUpdates
	wg.Add(1)
	go func() {
		defer wg.Done()
		statusStream, err := client.SendStatusUpdates(ctx, &pb.GetInstancesByRegionRequest{})
		if err != nil {
			log.Fatalf("error calling SendStatusUpdates: %v", err)
		}
		// Receive and process status updates
		for {
			status, err := statusStream.Recv()
			if err == io.EOF {
				break
			}
			if err != nil {
				log.Fatalf("error receiving status update: %v", err)
			}
			log.Printf("Received status update: %s\n", status)
		}
	}()

	// Wait for all goroutines to finish
	wg.Wait()
}
