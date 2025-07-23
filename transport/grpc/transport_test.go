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
	facade.EXPECT().SendMessage(in).Return(nil)

	conn, cleanup := startTestGRPCServer(t, facade)
	t.Cleanup(cleanup)
	client := pb.NewFacadeGRPCClient(conn)

	_, err := client.SendMessage(context.Background(), &pb.SendMessageRequest{
		NewMessage: &pb.NewMessage{
			ChatId: 1,
			Text:   "hi",
		},
	})
	assert.NoError(t, err)
}

func TestSendMessageAlbum(t *testing.T) {
	t.Parallel()

	facade := mocks.NewFacadeGRPC(t)
	expectedMessages := []*dto.NewMessage{
		{ChatId: 1, Text: "first", ReplyToMessageId: 10, FilePath: "123"},
		{ChatId: 1, Text: "second", ReplyToMessageId: 10, FilePath: "456"},
	}
	facade.EXPECT().SendMessageAlbum(expectedMessages).Return(nil)

	conn, cleanup := startTestGRPCServer(t, facade)
	t.Cleanup(cleanup)
	client := pb.NewFacadeGRPCClient(conn)

	_, err := client.SendMessageAlbum(context.Background(), &pb.SendMessageAlbumRequest{
		NewMessages: []*pb.NewMessage{
			{ChatId: 1, Text: "first", ReplyToMessageId: 10, FilePath: "123"},
			{ChatId: 1, Text: "second", ReplyToMessageId: 10, FilePath: "456"},
		},
	})
	assert.NoError(t, err)
}

func TestForwardMessage(t *testing.T) {
	t.Parallel()

	facade := mocks.NewFacadeGRPC(t)
	facade.EXPECT().ForwardMessage(int64(1), int64(42)).Return(nil)

	conn, cleanup := startTestGRPCServer(t, facade)
	t.Cleanup(cleanup)
	client := pb.NewFacadeGRPCClient(conn)

	_, err := client.ForwardMessage(context.Background(), &pb.ForwardMessageRequest{ChatId: 1, MessageId: 42})
	assert.NoError(t, err)
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
	in := &dto.Message{Id: 42, ChatId: 1, Text: "upd", FilePath: ""}
	facade.EXPECT().UpdateMessage(in).Return(nil)

	conn, cleanup := startTestGRPCServer(t, facade)
	t.Cleanup(cleanup)
	client := pb.NewFacadeGRPCClient(conn)

	_, err := client.UpdateMessage(context.Background(), &pb.UpdateMessageRequest{
		Message: &pb.Message{
			Id:       42,
			ChatId:   1,
			Text:     "upd",
			FilePath: "",
		},
	})
	assert.NoError(t, err)
}

func TestDeleteMessage(t *testing.T) {
	t.Parallel()

	facade := mocks.NewFacadeGRPC(t)
	facade.EXPECT().DeleteMessages(int64(1), []int64{42}).Return(nil)

	conn, cleanup := startTestGRPCServer(t, facade)
	t.Cleanup(cleanup)
	client := pb.NewFacadeGRPCClient(conn)

	_, err := client.DeleteMessages(context.Background(), &pb.DeleteMessagesRequest{ChatId: 1, MessageIds: []int64{42}})
	assert.NoError(t, err)
}

func TestGetMessageLink(t *testing.T) {
	t.Parallel()

	facade := mocks.NewFacadeGRPC(t)
	link := "https://t.me/c/1/2"
	facade.EXPECT().GetMessageLink(int64(1), int64(2)).Return(link, nil)

	conn, cleanup := startTestGRPCServer(t, facade)
	t.Cleanup(cleanup)
	client := pb.NewFacadeGRPCClient(conn)

	resp, err := client.GetMessageLink(context.Background(), &pb.GetMessageLinkRequest{ChatId: 1, MessageId: 2})
	assert.NoError(t, err)
	assert.Equal(t, link, resp.Link)
}

func TestGetMessageLinkInfo(t *testing.T) {
	t.Parallel()

	facade := mocks.NewFacadeGRPC(t)
	link := "https://t.me/c/1/2"
	msg := &dto.Message{Id: 2, ChatId: 1, Text: "", Forward: true}
	facade.EXPECT().GetMessageLinkInfo(link).Return(msg, nil)

	conn, cleanup := startTestGRPCServer(t, facade)
	t.Cleanup(cleanup)
	client := pb.NewFacadeGRPCClient(conn)

	resp, err := client.GetMessageLinkInfo(context.Background(), &pb.GetMessageLinkInfoRequest{Link: link})
	assert.NoError(t, err)
	assert.Equal(t, int64(2), resp.Message.Id)
	assert.Equal(t, int64(1), resp.Message.ChatId)
	assert.True(t, resp.Message.Forward)
}

func TestGetChatHistory(t *testing.T) {
	t.Parallel()

	facade := mocks.NewFacadeGRPC(t)
	expectedMessages := []*dto.Message{
		{Id: 101, ChatId: 1, Text: "message 1", Forward: false},
		{Id: 102, ChatId: 1, Text: "message 2", Forward: true},
	}
	facade.EXPECT().GetChatHistory(int64(1), int64(100), int32(0), int32(2)).Return(expectedMessages, nil)

	conn, cleanup := startTestGRPCServer(t, facade)
	t.Cleanup(cleanup)
	client := pb.NewFacadeGRPCClient(conn)

	resp, err := client.GetChatHistory(context.Background(), &pb.GetChatHistoryRequest{
		ChatId:        1,
		FromMessageId: 100,
		Offset:        0,
		Limit:         2,
	})
	assert.NoError(t, err)
	assert.Len(t, resp.Messages, 2)

	assert.Equal(t, int64(101), resp.Messages[0].Id)
	assert.Equal(t, int64(1), resp.Messages[0].ChatId)
	assert.Equal(t, "message 1", resp.Messages[0].Text)
	assert.False(t, resp.Messages[0].Forward)

	assert.Equal(t, int64(102), resp.Messages[1].Id)
	assert.Equal(t, int64(1), resp.Messages[1].ChatId)
	assert.Equal(t, "message 2", resp.Messages[1].Text)
	assert.True(t, resp.Messages[1].Forward)
}
