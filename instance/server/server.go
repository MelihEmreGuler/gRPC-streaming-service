package main

import (
	"flag"
	"fmt"
	pb "github.com/MelihEmreGuler/gRPC-streaming-service/instance/instancepb"
	"github.com/golang/protobuf/ptypes/timestamp"
	"google.golang.org/grpc"
	"log"
	"net"
	"time"
)

var (
	port            = flag.Int("port", 50051, "The server port")
	sampleInstances = []*pb.Instance{
		{
			AmiLaunchIndex: 1,
			Architecture:   pb.ArchitectureValues_X86_64,
			BlockDeviceMappings: []*pb.InstanceBlockDeviceMapping{
				{
					DeviceName: "sda1",
					Ebs: &pb.EbsInstanceBlockDevice{
						AttachTime:          &timestamp.Timestamp{Seconds: 1629434214},
						DeleteOnTermination: true,
						Status:              pb.AttachmentStatus_attached,
						VolumeId:            "vol-0123456789abcdef0",
					},
				},
			},
		},
		{
			AmiLaunchIndex: 2,
			Architecture:   pb.ArchitectureValues_arm64,
			BlockDeviceMappings: []*pb.InstanceBlockDeviceMapping{
				{
					DeviceName: "sdb1",
					Ebs: &pb.EbsInstanceBlockDevice{
						AttachTime:          &timestamp.Timestamp{Seconds: 1629434214},
						DeleteOnTermination: false,
						Status:              pb.AttachmentStatus_detached,
						VolumeId:            "vol-0fedcba9876543210",
					},
				},
			},
		},
		// Add more instances as needed
	}
)

type instanceServer struct {
	pb.UnimplementedInstanceServer
	instances []*pb.Instance
}

// Create a channel to signal when instances are ready
var instancesReadyCh = make(chan struct{})

// GetInstancesByRegion returns the instances in the given region.
func (s *instanceServer) GetInstancesByRegion(req *pb.GetInstancesByRegionRequest, stream pb.Instance_GetInstancesByRegionServer) error {
	log.Printf("Received GetInstancesByRegion request from client: \n %+v\n", req)

	// Simulate some work to retrieve instances (replace this with your actual logic)
	go func() {
		time.Sleep(1 * time.Second) // Simulate work
		// When instances are ready, signal it
		close(instancesReadyCh)
	}()

	err := stream.Send(&pb.GetInstancesByRegionResponse{Instances: s.instances})
	if err != nil {
		return err
	}
	// Wait for the instances to send to the client
	<-instancesReadyCh
	return nil
}

// SendStatusUpdates sends status updates to the client.
func (s *instanceServer) SendStatusUpdates(req *pb.GetInstancesByRegionRequest, stream pb.Instance_SendStatusUpdatesServer) error {
	// Notify the client that scanning is starting
	statusUpdate := &pb.StatusUpdate{Message: "Scanning started..."}
	if err := stream.Send(statusUpdate); err != nil {
		return err
	}

	//communicate with the GetInstancesByRegion function and wait here until sending the instances to client
	// Wait for instances to be ready
	<-instancesReadyCh

	// Notify the client that scanning is complete
	statusUpdate = &pb.StatusUpdate{Message: "Scanning completed."}
	if err := stream.Send(statusUpdate); err != nil {
		return err
	}

	return nil
}

func newServer() *instanceServer {
	s := &instanceServer{
		instances: sampleInstances,
	}
	return s
}

func main() {
	flag.Parse()
	lis, err := net.Listen("tcp", fmt.Sprintf("localhost:%d", *port))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	var opts []grpc.ServerOption

	grpcServer := grpc.NewServer(opts...)
	pb.RegisterInstanceServer(grpcServer, newServer())
	err = grpcServer.Serve(lis)
	if err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
