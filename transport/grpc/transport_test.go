package grpc

import (
	"context"
	"net"
	"testing"

	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc"
	"google.golang.org/grpc/test/bufconn"

	"github.com/comerc/budva43/app/dto/grpc/dto"
	"github.com/comerc/budva43/transport/grpc/mocks"
	pb "github.com/comerc/budva43/transport/grpc/pb"
)

const bufSize = 1024 * 1024

func startTestGRPCServer(t *testing.T, facade *mocks.FacadeGRPC) (*grpc.ClientConn, func()) {
	lis := bufconn.Listen(bufSize)
	server := grpc.NewServer()
	tr := New(facade)
	pb.RegisterFacadeGRPCServer(server, tr)
	go func() {
		_ = server.Serve(lis)
	}()
	conn, err := grpc.DialContext(context.Background(), "bufnet",
		grpc.WithContextDialer(func(ctx context.Context, s string) (net.Conn, error) {
			return lis.Dial()
		}),
		grpc.WithInsecure(), // TODO: deprecated
	)
	assert.NoError(t, err)
	return conn, func() {
		server.GracefulStop()
		_ = lis.Close()
	}
}

func TestSendMessage(t *testing.T) {
	t.Parallel()

	facade := mocks.NewFacadeGRPC(t)
	in := &dto.NewMessage{ChatId: 1, Text: "hi"}
	out := &dto.Message{Id: 42, ChatId: 1, Text: "hi"}
	facade.EXPECT().SendMessage(in).Return(out, nil)

	conn, cleanup := startTestGRPCServer(t, facade)
	t.Cleanup(cleanup)
	client := pb.NewFacadeGRPCClient(conn)

	resp, err := client.SendMessage(context.Background(), &pb.SendMessageRequest{ChatId: 1, Text: "hi"})
	assert.NoError(t, err)
	assert.Equal(t, int64(42), resp.Message.Id)
	assert.Equal(t, int64(1), resp.Message.ChatId)
	assert.Equal(t, "hi", resp.Message.Text)
}

func TestForwardMessage(t *testing.T) {
	t.Parallel()

	facade := mocks.NewFacadeGRPC(t)
	out := &dto.Message{Id: 42, ChatId: 1, Text: "hi"}
	facade.EXPECT().ForwardMessage(int64(1), int64(42)).Return(out, nil)

	conn, cleanup := startTestGRPCServer(t, facade)
	t.Cleanup(cleanup)
	client := pb.NewFacadeGRPCClient(conn)

	resp, err := client.ForwardMessage(context.Background(), &pb.ForwardMessageRequest{ChatId: 1, MessageId: 42})
	assert.NoError(t, err)
	assert.Equal(t, int64(42), resp.Message.Id)
	assert.Equal(t, int64(1), resp.Message.ChatId)
	assert.Equal(t, "hi", resp.Message.Text)
}

func TestGetMessages(t *testing.T) {
	t.Parallel()

	facade := mocks.NewFacadeGRPC(t)
	facade.EXPECT().GetMessages(int64(1), []int64{1, 2}).Return([]*dto.Message{
		{Id: 1, ChatId: 1, Text: "a"},
		{Id: 2, ChatId: 1, Text: "b"},
	}, nil)

	conn, cleanup := startTestGRPCServer(t, facade)
	t.Cleanup(cleanup)
	client := pb.NewFacadeGRPCClient(conn)

	resp, err := client.GetMessages(context.Background(), &pb.GetMessagesRequest{ChatId: 1, MessageIds: []int64{1, 2}})
	assert.NoError(t, err)
	assert.Len(t, resp.Messages, 2)
	assert.Equal(t, "a", resp.Messages[0].Text)
	assert.Equal(t, "b", resp.Messages[1].Text)
}

func TestGetLastMessage(t *testing.T) {
	t.Parallel()

	facade := mocks.NewFacadeGRPC(t)
	facade.EXPECT().GetLastMessage(int64(1)).Return(&dto.Message{Id: 42, ChatId: 1, Text: "hi"}, nil)

	conn, cleanup := startTestGRPCServer(t, facade)
	t.Cleanup(cleanup)
	client := pb.NewFacadeGRPCClient(conn)
	resp, err := client.GetLastMessage(context.Background(), &pb.GetLastMessageRequest{ChatId: 1})
	assert.NoError(t, err)
	assert.Equal(t, int64(42), resp.Message.Id)
	assert.Equal(t, int64(1), resp.Message.ChatId)
	assert.Equal(t, "hi", resp.Message.Text)
}

func TestGetMessage(t *testing.T) {
	t.Parallel()

	facade := mocks.NewFacadeGRPC(t)
	out := &dto.Message{Id: 42, ChatId: 1, Text: "hi"}
	facade.EXPECT().GetMessage(int64(1), int64(42)).Return(out, nil)

	conn, cleanup := startTestGRPCServer(t, facade)
	t.Cleanup(cleanup)
	client := pb.NewFacadeGRPCClient(conn)

	resp, err := client.GetMessage(context.Background(), &pb.GetMessageRequest{ChatId: 1, MessageId: 42})
	assert.NoError(t, err)
	assert.Equal(t, int64(42), resp.Message.Id)
	assert.Equal(t, int64(1), resp.Message.ChatId)
	assert.Equal(t, "hi", resp.Message.Text)
}

func TestUpdateMessage(t *testing.T) {
	t.Parallel()

	facade := mocks.NewFacadeGRPC(t)
	in := &dto.Message{Id: 42, ChatId: 1, Text: "upd"}
	out := &dto.Message{Id: 42, ChatId: 1, Text: "upd"}
	facade.EXPECT().UpdateMessage(in).Return(out, nil)

	conn, cleanup := startTestGRPCServer(t, facade)
	t.Cleanup(cleanup)
	client := pb.NewFacadeGRPCClient(conn)

	resp, err := client.UpdateMessage(context.Background(), &pb.UpdateMessageRequest{MessageId: 42, ChatId: 1, Text: "upd"})
	assert.NoError(t, err)
	assert.Equal(t, int64(42), resp.Message.Id)
	assert.Equal(t, int64(1), resp.Message.ChatId)
	assert.Equal(t, "upd", resp.Message.Text)
}

func TestDeleteMessage(t *testing.T) {
	t.Parallel()

	facade := mocks.NewFacadeGRPC(t)
	facade.EXPECT().DeleteMessages(int64(1), []int64{42}).Return(true, nil)

	conn, cleanup := startTestGRPCServer(t, facade)
	t.Cleanup(cleanup)
	client := pb.NewFacadeGRPCClient(conn)

	resp, err := client.DeleteMessages(context.Background(), &pb.DeleteMessagesRequest{ChatId: 1, MessageIds: []int64{42}})
	assert.NoError(t, err)
	assert.True(t, resp.Success)
}
