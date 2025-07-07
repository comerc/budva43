package grpc

import (
	"context"
	"net"
	"testing"

	dto "github.com/comerc/budva43/app/dto/grpc"
	"github.com/comerc/budva43/transport/grpc/mocks"
	pb "github.com/comerc/budva43/transport/grpc/pb"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc"
	"google.golang.org/grpc/test/bufconn"
)

const bufSize = 1024 * 1024

func bufconnDialer(lis *bufconn.Listener) func(context.Context, string) (net.Conn, error) {
	return func(ctx context.Context, s string) (net.Conn, error) {
		return lis.Dial()
	}
}

func startTestGRPCServer(t *testing.T, facade *mocks.FacadeGRPC) (*grpc.ClientConn, func()) {
	lis := bufconn.Listen(bufSize)
	server := grpc.NewServer()
	tr := New(facade)
	pb.RegisterFacadeGRPCServer(server, tr)
	go func() {
		_ = server.Serve(lis)
	}()
	conn, err := grpc.DialContext(context.Background(), "bufnet", grpc.WithContextDialer(bufconnDialer(lis)), grpc.WithInsecure())
	assert.NoError(t, err)
	return conn, func() {
		server.GracefulStop()
		_ = lis.Close()
	}
}

func TestCreateMessage(t *testing.T) {
	facade := mocks.NewFacadeGRPC(t)
	in := &dto.NewMessage{ChatId: 1, Content: "hi"}
	out := &dto.Message{Id: 42, ChatId: 1, Content: "hi"}
	facade.EXPECT().CreateMessage(in).Return(out, nil)

	conn, cleanup := startTestGRPCServer(t, facade)
	defer cleanup()
	client := pb.NewFacadeGRPCClient(conn)

	resp, err := client.CreateMessage(context.Background(), &pb.CreateMessageRequest{ChatId: 1, Content: "hi"})
	assert.NoError(t, err)
	assert.Equal(t, int64(42), resp.Message.MessageId)
	assert.Equal(t, int64(1), resp.Message.ChatId)
	assert.Equal(t, "hi", resp.Message.Content)
}

func TestGetMessages(t *testing.T) {
	facade := mocks.NewFacadeGRPC(t)
	facade.EXPECT().GetMessages(int64(1)).Return([]*dto.Message{
		{Id: 1, ChatId: 1, Content: "a"},
		{Id: 2, ChatId: 1, Content: "b"},
	}, nil)

	conn, cleanup := startTestGRPCServer(t, facade)
	defer cleanup()
	client := pb.NewFacadeGRPCClient(conn)

	resp, err := client.GetMessages(context.Background(), &pb.GetMessagesRequest{ChatId: 1})
	assert.NoError(t, err)
	assert.Len(t, resp.Messages, 2)
	assert.Equal(t, "a", resp.Messages[0].Content)
	assert.Equal(t, "b", resp.Messages[1].Content)
}

func TestGetMessage(t *testing.T) {
	facade := mocks.NewFacadeGRPC(t)
	out := &dto.Message{Id: 42, ChatId: 1, Content: "hi"}
	facade.EXPECT().GetMessage(int64(42)).Return(out, nil)

	conn, cleanup := startTestGRPCServer(t, facade)
	defer cleanup()
	client := pb.NewFacadeGRPCClient(conn)

	resp, err := client.GetMessage(context.Background(), &pb.GetMessageRequest{MessageId: 42})
	assert.NoError(t, err)
	assert.Equal(t, int64(42), resp.Message.MessageId)
	assert.Equal(t, int64(1), resp.Message.ChatId)
	assert.Equal(t, "hi", resp.Message.Content)
}

func TestUpdateMessage(t *testing.T) {
	facade := mocks.NewFacadeGRPC(t)
	in := &dto.Message{Id: 42, ChatId: 1, Content: "upd"}
	out := &dto.Message{Id: 42, ChatId: 1, Content: "upd"}
	facade.EXPECT().UpdateMessage(in).Return(out, nil)

	conn, cleanup := startTestGRPCServer(t, facade)
	defer cleanup()
	client := pb.NewFacadeGRPCClient(conn)

	resp, err := client.UpdateMessage(context.Background(), &pb.UpdateMessageRequest{MessageId: 42, ChatId: 1, Content: "upd"})
	assert.NoError(t, err)
	assert.Equal(t, int64(42), resp.Message.MessageId)
	assert.Equal(t, int64(1), resp.Message.ChatId)
	assert.Equal(t, "upd", resp.Message.Content)
}

func TestDeleteMessage(t *testing.T) {
	facade := mocks.NewFacadeGRPC(t)
	in := &dto.Message{Id: 42}
	out := &dto.Message{Id: 42}
	facade.EXPECT().DeleteMessage(in).Return(out, nil)

	conn, cleanup := startTestGRPCServer(t, facade)
	defer cleanup()
	client := pb.NewFacadeGRPCClient(conn)

	resp, err := client.DeleteMessage(context.Background(), &pb.DeleteMessageRequest{MessageId: 42})
	assert.NoError(t, err)
	assert.True(t, resp.Success)
}

func TestGetClientDone(t *testing.T) {
	facade := mocks.NewFacadeGRPC(t)
	ch := make(chan any)
	facade.EXPECT().GetClientDone().Return((<-chan any)(ch))

	conn, cleanup := startTestGRPCServer(t, facade)
	defer cleanup()
	client := pb.NewFacadeGRPCClient(conn)

	resp, err := client.GetClientDone(context.Background(), &pb.EmptyRequest{})
	assert.NoError(t, err)
	assert.False(t, resp.Done)

	// Закрываем канал, чтобы проверить Done = true
	close(ch)
	resp2, err2 := client.GetClientDone(context.Background(), &pb.EmptyRequest{})
	assert.NoError(t, err2)
	assert.True(t, resp2.Done)
}
