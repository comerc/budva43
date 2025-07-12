package grpc

import (
	"context"
	"net"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/reflection"
	"google.golang.org/grpc/status"

	"github.com/comerc/budva43/app/dto/grpc/dto"
	"github.com/comerc/budva43/app/log"
	"github.com/comerc/budva43/transport/grpc/pb"
)

//go:generate mockery --name=facadeGRPC --exported
type facadeGRPC interface {
	GetClientDone() <-chan any
	GetMessages(chatId int64, messageIds []int64) ([]*dto.Message, error)
	GetLastMessage(chatId int64) (*dto.Message, error)
	SendMessage(message *dto.NewMessage) (*dto.Message, error)
	ForwardMessage(chatId int64, messageId int64) (*dto.Message, error)
	GetMessage(chatId int64, messageId int64) (*dto.Message, error)
	UpdateMessage(message *dto.Message) (*dto.Message, error)
	DeleteMessages(chatId int64, messageIds []int64) (bool, error)
}

type Transport struct {
	log *log.Logger
	//
	pb.UnimplementedFacadeGRPCServer
	facade facadeGRPC
	server *grpc.Server
	lis    net.Listener
}

func New(facade facadeGRPC) *Transport {
	return &Transport{
		log: log.NewLogger(),
		//
		facade: facade,
	}
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
	reflection.Register(t.server)
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
	var err error

	var msgs []*dto.Message
	msgs, err = t.facade.GetMessages(req.ChatId, req.MessageIds)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	res := &pb.GetMessagesResponse{}
	for _, m := range msgs {
		res.Messages = append(res.Messages, &pb.Message{
			Id:      m.Id,
			ChatId:  m.ChatId,
			Text:    m.Text,
			Forward: m.Forward,
		})
	}
	return res, nil
}

func (t *Transport) GetLastMessage(ctx context.Context, req *pb.GetLastMessageRequest) (*pb.MessageResponse, error) {
	var err error

	var res *dto.Message
	res, err = t.facade.GetLastMessage(req.ChatId)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	if res == nil {
		return nil, nil
	}
	return &pb.MessageResponse{Message: &pb.Message{
		Id:      res.Id,
		ChatId:  res.ChatId,
		Text:    res.Text,
		Forward: res.Forward,
	}}, nil
}

func (t *Transport) SendMessage(ctx context.Context, req *pb.SendMessageRequest) (*pb.MessageResponse, error) {
	var err error

	var res *dto.Message
	res, err = t.facade.SendMessage(&dto.NewMessage{
		ChatId: req.ChatId,
		Text:   req.Text,
	})
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	if res == nil {
		return nil, nil
	}
	return &pb.MessageResponse{Message: &pb.Message{
		Id:      res.Id,
		ChatId:  res.ChatId,
		Text:    res.Text,
		Forward: res.Forward,
	}}, nil
}

func (t *Transport) ForwardMessage(ctx context.Context, req *pb.ForwardMessageRequest) (*pb.MessageResponse, error) {
	var err error

	var res *dto.Message
	res, err = t.facade.ForwardMessage(req.ChatId, req.MessageId)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	if res == nil {
		return nil, nil
	}
	return &pb.MessageResponse{Message: &pb.Message{
		Id:      res.Id,
		ChatId:  res.ChatId,
		Text:    res.Text,
		Forward: res.Forward,
	}}, nil
}

func (t *Transport) GetMessage(ctx context.Context, req *pb.GetMessageRequest) (*pb.MessageResponse, error) {
	var err error

	var res *dto.Message
	res, err = t.facade.GetMessage(req.ChatId, req.MessageId)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &pb.MessageResponse{Message: &pb.Message{
		Id:      res.Id,
		ChatId:  res.ChatId,
		Text:    res.Text,
		Forward: res.Forward,
	}}, nil
}

func (t *Transport) UpdateMessage(ctx context.Context, req *pb.UpdateMessageRequest) (*pb.MessageResponse, error) {
	var err error

	var res *dto.Message
	res, err = t.facade.UpdateMessage(&dto.Message{
		Id:     req.MessageId,
		ChatId: req.ChatId,
		Text:   req.Text,
	})
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &pb.MessageResponse{Message: &pb.Message{
		Id:      res.Id,
		ChatId:  res.ChatId,
		Text:    res.Text,
		Forward: res.Forward,
	}}, nil
}

func (t *Transport) DeleteMessages(ctx context.Context, req *pb.DeleteMessagesRequest) (*pb.DeleteMessagesResponse, error) {
	var err error

	var ok bool
	ok, err = t.facade.DeleteMessages(req.ChatId, req.MessageIds)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &pb.DeleteMessagesResponse{Success: ok}, nil
}
