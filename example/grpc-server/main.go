package main

import (
	"context"
	"fmt"
	"github.com/sestack/smq/example/grpc-server/pb"
	"google.golang.org/grpc"
	"log"
	"net"
)

type NotifyServiceImpl struct {
}

func NewServiceServerImpl() *NotifyServiceImpl {
	return &NotifyServiceImpl{}
}

func (ns *NotifyServiceImpl) OnClientConnect(ctx context.Context, request *pb.NotifyRequest) (*pb.NotifyResponse, error) {
	fmt.Println("OnClientConnect")
	status := true
	fmt.Println(string(request.Data.Value))
	return &pb.NotifyResponse{Status: status}, nil
}

func (ns *NotifyServiceImpl) OnClientDisconnected(ctx context.Context, request *pb.NotifyRequest) (*pb.NotifyResponse, error) {
	fmt.Println("OnClientDisconnected")
	status := true
	fmt.Println(string(request.Data.Value))
	return &pb.NotifyResponse{Status: status}, nil
}

func (ns *NotifyServiceImpl) OnSubscribe(ctx context.Context, request *pb.NotifyRequest) (*pb.NotifyResponse, error) {
	fmt.Println("OnSubscribe")
	status := true
	fmt.Println(string(request.Data.Value))
	return &pb.NotifyResponse{Status: status}, nil
}

func (ns *NotifyServiceImpl) OnUnSubscribe(ctx context.Context, request *pb.NotifyRequest) (*pb.NotifyResponse, error) {
	fmt.Println("OnUnSubscribe")
	status := true
	fmt.Println(string(request.Data.Value))
	return &pb.NotifyResponse{Status: status}, nil
}

func (ns *NotifyServiceImpl) OnAgentConnect(ctx context.Context, request *pb.NotifyRequest) (*pb.NotifyResponse, error) {
	fmt.Println("OnAgentConnect")
	status := true
	fmt.Println(string(request.Data.Value))
	return &pb.NotifyResponse{Status: status}, nil
}

func (ns *NotifyServiceImpl) OnAgentDisconnected(ctx context.Context, request *pb.NotifyRequest) (*pb.NotifyResponse, error) {
	fmt.Println("OnAgentDisconnected")
	status := true
	fmt.Println(string(request.Data.Value))
	return &pb.NotifyResponse{Status: status}, nil
}


func main() {
	rpcServer := grpc.NewServer()
	pb.RegisterNotifyServer(rpcServer, NewServiceServerImpl())

	listener, err := net.Listen("tcp", "127.0.0.1:9999")
	if err != nil {
		log.Fatal("服务监听端口失败", err)
	}

	_ = rpcServer.Serve(listener)
}
