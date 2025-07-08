package grpc

import (
	"context"
	"net"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/comerc/budva43/app/dto/grpc/dto"
	"github.com/comerc/budva43/transport/grpc/pb"
)

//go:generate mockery --name=facadeGRPC --exported
type facadeGRPC interface {
	GetClientDone() <-chan any
	GetMessages(chatId int64) ([]*dto.Message, error)
	CreateMessage(message *dto.NewMessage) (*dto.Message, error)
	GetMessage(messageId int64) (*dto.Message, error)
	UpdateMessage(message *dto.Message) (*dto.Message, error)
	DeleteMessage(message *dto.Message) (*dto.Message, error)
}

type Transport struct {
	pb.UnimplementedFacadeGRPCServer
	facade facadeGRPC
	server *grpc.Server
	lis    net.Listener
}

func New(facade facadeGRPC) *Transport {
	return &Transport{facade: facade}
}

func (t *Transport) Start() error {
	// TODO: read from config
	lis, err := net.Listen("tcp", ":50051")
	if err != nil {
		return err
	}
	t.lis = lis
	t.server = grpc.NewServer()
	pb.RegisterFacadeGRPCServer(t.server, t)
	go func() {
		_ = t.server.Serve(lis)
	}()
	return nil
}

func (t *Transport) Close() error {
	t.server.GracefulStop()
	return t.lis.Close()
}

func (t *Transport) GetClientDone(ctx context.Context, req *pb.EmptyRequest) (*pb.ClientDoneResponse, error) {
	select {
	case <-t.facade.GetClientDone():
		return &pb.ClientDoneResponse{Done: true}, nil
	default:
		return &pb.ClientDoneResponse{Done: false}, nil
	}
}

func (t *Transport) GetMessages(ctx context.Context, req *pb.GetMessagesRequest) (*pb.GetMessagesResponse, error) {
	msgs, err := t.facade.GetMessages(req.ChatId)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	res := &pb.GetMessagesResponse{}
	for _, m := range msgs {
		res.Messages = append(res.Messages, &pb.Message{
			MessageId: m.Id,
			ChatId:    m.ChatId,
			Text:      m.Text,
		})
	}
	return res, nil
}

func (t *Transport) CreateMessage(ctx context.Context, req *pb.CreateMessageRequest) (*pb.MessageResponse, error) {
	res, err := t.facade.CreateMessage(&dto.NewMessage{
		ChatId: req.ChatId,
		Text:   req.Text,
	})
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &pb.MessageResponse{Message: &pb.Message{
		MessageId: res.Id,
		ChatId:    res.ChatId,
		Text:      res.Text,
	}}, nil
}

func (t *Transport) GetMessage(ctx context.Context, req *pb.GetMessageRequest) (*pb.MessageResponse, error) {
	res, err := t.facade.GetMessage(req.MessageId)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &pb.MessageResponse{Message: &pb.Message{
		MessageId: res.Id,
		ChatId:    res.ChatId,
		Text:      res.Text,
	}}, nil
}

func (t *Transport) UpdateMessage(ctx context.Context, req *pb.UpdateMessageRequest) (*pb.MessageResponse, error) {
	res, err := t.facade.UpdateMessage(&dto.Message{
		Id:     req.MessageId,
		ChatId: req.ChatId,
		Text:   req.Text,
	})
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &pb.MessageResponse{Message: &pb.Message{
		MessageId: res.Id,
		ChatId:    res.ChatId,
		Text:      res.Text,
	}}, nil
}

func (t *Transport) DeleteMessage(ctx context.Context, req *pb.DeleteMessageRequest) (*pb.DeleteMessageResponse, error) {
	res, err := t.facade.DeleteMessage(&dto.Message{Id: req.MessageId})
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &pb.DeleteMessageResponse{Success: res != nil}, nil
}
