package notify

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/golang/protobuf/ptypes/any"
	"github.com/sestack/smq/global"
	"github.com/sestack/smq/notify/pb"
	"google.golang.org/grpc"
	"google.golang.org/grpc/keepalive"
	"log"
	"time"
)

type notificationGrpc struct {
	client pb.NotifyClient
}

func InitGrpc() *notificationGrpc {
	conn, err := grpc.DialContext(context.Background(), global.CONFIG.Notify.Server,
		grpc.WithInsecure(),
		grpc.WithBlock(),
		grpc.WithKeepaliveParams(keepalive.ClientParameters{
			Time:                10 * time.Second,
			Timeout:             100 * time.Millisecond,
			PermitWithoutStream: true}),
	)
	if err != nil {
		log.Fatalf("grpc.DialContext err: %v", err)
	}

	client := pb.NewNotifyClient(conn)

	return &notificationGrpc{
		client: client,
	}
}

func (h *notificationGrpc) OnClientConnect(data interface{}) bool {
	byteData, err := json.Marshal(data)
	if err != nil {
		log.Printf("Could not convert data input to bytes")
	}

	e := &any.Any{
		TypeUrl: "anything",
		Value:   byteData,
	}

	resp, err := h.client.OnClientConnect(context.Background(), &pb.NotifyRequest{
		Data: e,
	})
	if err != nil {
		fmt.Println(err)
		return false
	}

	return resp.Status
}
func (h *notificationGrpc) OnClientDisconnected(data interface{}) bool {
	byteData, err := json.Marshal(data)
	if err != nil {
		log.Printf("Could not convert data input to bytes")
	}

	e := &any.Any{
		TypeUrl: "anything",
		Value:   byteData,
	}

	resp, err := h.client.OnClientDisconnected(context.Background(), &pb.NotifyRequest{
		Data: e,
	})
	if err != nil {
		return false
	}

	return resp.Status
}
func (h *notificationGrpc) OnSubscribe(data interface{}) bool {
	byteData, err := json.Marshal(data)
	if err != nil {
		log.Printf("Could not convert data input to bytes")
	}

	e := &any.Any{
		TypeUrl: "anything",
		Value:   byteData,
	}

	resp, err := h.client.OnSubscribe(context.Background(), &pb.NotifyRequest{
		Data: e,
	})
	if err != nil {
		return false
	}

	return resp.Status
}
func (h *notificationGrpc) OnUnSubscribe(data interface{}) bool {
	byteData, err := json.Marshal(data)
	if err != nil {
		log.Printf("Could not convert data input to bytes")
	}

	e := &any.Any{
		TypeUrl: "anything",
		Value:   byteData,
	}

	resp, err := h.client.OnUnSubscribe(context.Background(), &pb.NotifyRequest{
		Data: e,
	})
	if err != nil {
		return false
	}

	return resp.Status
}
func (h *notificationGrpc) OnAgentConnect(data interface{}) bool {
	byteData, err := json.Marshal(data)
	if err != nil {
		log.Printf("Could not convert data input to bytes")
	}

	e := &any.Any{
		TypeUrl: "anything",
		Value:   byteData,
	}

	resp, err := h.client.OnAgentConnect(context.Background(), &pb.NotifyRequest{
		Data: e,
	})
	if err != nil {
		return false
	}

	return resp.Status
}
func (h *notificationGrpc) OnAgentDisconnected(data interface{}) bool {
	byteData, err := json.Marshal(data)
	if err != nil {
		log.Printf("Could not convert data input to bytes")
	}

	e := &any.Any{
		TypeUrl: "anything",
		Value:   byteData,
	}

	resp, err := h.client.OnAgentDisconnected(context.Background(), &pb.NotifyRequest{
		Data: e,
	})
	if err != nil {
		return false
	}

	return resp.Status
}
