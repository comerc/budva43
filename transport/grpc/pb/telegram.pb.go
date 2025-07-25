// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.36.6
// 	protoc        v5.29.3
// source: transport/grpc/pb/telegram.proto

package pb

import (
	protoreflect "google.golang.org/protobuf/reflect/protoreflect"
	protoimpl "google.golang.org/protobuf/runtime/protoimpl"
	reflect "reflect"
	sync "sync"
	unsafe "unsafe"
)

const (
	// Verify that this generated code is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(20 - protoimpl.MinVersion)
	// Verify that runtime/protoimpl is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(protoimpl.MaxVersion - 20)
)

type NewMessage struct {
	state            protoimpl.MessageState `protogen:"open.v1"`
	ChatId           int64                  `protobuf:"varint,2,opt,name=chat_id,json=chatId,proto3" json:"chat_id,omitempty"`
	Text             string                 `protobuf:"bytes,1,opt,name=text,proto3" json:"text,omitempty"`
	ReplyToMessageId int64                  `protobuf:"varint,3,opt,name=reply_to_message_id,json=replyToMessageId,proto3" json:"reply_to_message_id,omitempty"`
	FilePath         string                 `protobuf:"bytes,4,opt,name=file_path,json=filePath,proto3" json:"file_path,omitempty"`
	unknownFields    protoimpl.UnknownFields
	sizeCache        protoimpl.SizeCache
}

func (x *NewMessage) Reset() {
	*x = NewMessage{}
	mi := &file_transport_grpc_pb_telegram_proto_msgTypes[0]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *NewMessage) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*NewMessage) ProtoMessage() {}

func (x *NewMessage) ProtoReflect() protoreflect.Message {
	mi := &file_transport_grpc_pb_telegram_proto_msgTypes[0]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use NewMessage.ProtoReflect.Descriptor instead.
func (*NewMessage) Descriptor() ([]byte, []int) {
	return file_transport_grpc_pb_telegram_proto_rawDescGZIP(), []int{0}
}

func (x *NewMessage) GetChatId() int64 {
	if x != nil {
		return x.ChatId
	}
	return 0
}

func (x *NewMessage) GetText() string {
	if x != nil {
		return x.Text
	}
	return ""
}

func (x *NewMessage) GetReplyToMessageId() int64 {
	if x != nil {
		return x.ReplyToMessageId
	}
	return 0
}

func (x *NewMessage) GetFilePath() string {
	if x != nil {
		return x.FilePath
	}
	return ""
}

type Message struct {
	state         protoimpl.MessageState `protogen:"open.v1"`
	Id            int64                  `protobuf:"varint,1,opt,name=id,proto3" json:"id,omitempty"`
	ChatId        int64                  `protobuf:"varint,2,opt,name=chat_id,json=chatId,proto3" json:"chat_id,omitempty"`
	Text          string                 `protobuf:"bytes,3,opt,name=text,proto3" json:"text,omitempty"`
	Forward       bool                   `protobuf:"varint,4,opt,name=forward,proto3" json:"forward,omitempty"`
	FilePath      string                 `protobuf:"bytes,5,opt,name=file_path,json=filePath,proto3" json:"file_path,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *Message) Reset() {
	*x = Message{}
	mi := &file_transport_grpc_pb_telegram_proto_msgTypes[1]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *Message) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Message) ProtoMessage() {}

func (x *Message) ProtoReflect() protoreflect.Message {
	mi := &file_transport_grpc_pb_telegram_proto_msgTypes[1]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Message.ProtoReflect.Descriptor instead.
func (*Message) Descriptor() ([]byte, []int) {
	return file_transport_grpc_pb_telegram_proto_rawDescGZIP(), []int{1}
}

func (x *Message) GetId() int64 {
	if x != nil {
		return x.Id
	}
	return 0
}

func (x *Message) GetChatId() int64 {
	if x != nil {
		return x.ChatId
	}
	return 0
}

func (x *Message) GetText() string {
	if x != nil {
		return x.Text
	}
	return ""
}

func (x *Message) GetForward() bool {
	if x != nil {
		return x.Forward
	}
	return false
}

func (x *Message) GetFilePath() string {
	if x != nil {
		return x.FilePath
	}
	return ""
}

type GetMessagesRequest struct {
	state         protoimpl.MessageState `protogen:"open.v1"`
	ChatId        int64                  `protobuf:"varint,1,opt,name=chat_id,json=chatId,proto3" json:"chat_id,omitempty"`
	MessageIds    []int64                `protobuf:"varint,2,rep,packed,name=message_ids,json=messageIds,proto3" json:"message_ids,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *GetMessagesRequest) Reset() {
	*x = GetMessagesRequest{}
	mi := &file_transport_grpc_pb_telegram_proto_msgTypes[2]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *GetMessagesRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*GetMessagesRequest) ProtoMessage() {}

func (x *GetMessagesRequest) ProtoReflect() protoreflect.Message {
	mi := &file_transport_grpc_pb_telegram_proto_msgTypes[2]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use GetMessagesRequest.ProtoReflect.Descriptor instead.
func (*GetMessagesRequest) Descriptor() ([]byte, []int) {
	return file_transport_grpc_pb_telegram_proto_rawDescGZIP(), []int{2}
}

func (x *GetMessagesRequest) GetChatId() int64 {
	if x != nil {
		return x.ChatId
	}
	return 0
}

func (x *GetMessagesRequest) GetMessageIds() []int64 {
	if x != nil {
		return x.MessageIds
	}
	return nil
}

type GetChatHistoryRequest struct {
	state         protoimpl.MessageState `protogen:"open.v1"`
	ChatId        int64                  `protobuf:"varint,1,opt,name=chat_id,json=chatId,proto3" json:"chat_id,omitempty"`
	FromMessageId int64                  `protobuf:"varint,2,opt,name=from_message_id,json=fromMessageId,proto3" json:"from_message_id,omitempty"`
	Offset        int32                  `protobuf:"varint,3,opt,name=offset,proto3" json:"offset,omitempty"`
	Limit         int32                  `protobuf:"varint,4,opt,name=limit,proto3" json:"limit,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *GetChatHistoryRequest) Reset() {
	*x = GetChatHistoryRequest{}
	mi := &file_transport_grpc_pb_telegram_proto_msgTypes[3]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *GetChatHistoryRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*GetChatHistoryRequest) ProtoMessage() {}

func (x *GetChatHistoryRequest) ProtoReflect() protoreflect.Message {
	mi := &file_transport_grpc_pb_telegram_proto_msgTypes[3]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use GetChatHistoryRequest.ProtoReflect.Descriptor instead.
func (*GetChatHistoryRequest) Descriptor() ([]byte, []int) {
	return file_transport_grpc_pb_telegram_proto_rawDescGZIP(), []int{3}
}

func (x *GetChatHistoryRequest) GetChatId() int64 {
	if x != nil {
		return x.ChatId
	}
	return 0
}

func (x *GetChatHistoryRequest) GetFromMessageId() int64 {
	if x != nil {
		return x.FromMessageId
	}
	return 0
}

func (x *GetChatHistoryRequest) GetOffset() int32 {
	if x != nil {
		return x.Offset
	}
	return 0
}

func (x *GetChatHistoryRequest) GetLimit() int32 {
	if x != nil {
		return x.Limit
	}
	return 0
}

type MessagesResponse struct {
	state         protoimpl.MessageState `protogen:"open.v1"`
	Messages      []*Message             `protobuf:"bytes,1,rep,name=messages,proto3" json:"messages,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *MessagesResponse) Reset() {
	*x = MessagesResponse{}
	mi := &file_transport_grpc_pb_telegram_proto_msgTypes[4]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *MessagesResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*MessagesResponse) ProtoMessage() {}

func (x *MessagesResponse) ProtoReflect() protoreflect.Message {
	mi := &file_transport_grpc_pb_telegram_proto_msgTypes[4]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use MessagesResponse.ProtoReflect.Descriptor instead.
func (*MessagesResponse) Descriptor() ([]byte, []int) {
	return file_transport_grpc_pb_telegram_proto_rawDescGZIP(), []int{4}
}

func (x *MessagesResponse) GetMessages() []*Message {
	if x != nil {
		return x.Messages
	}
	return nil
}

type SendMessageRequest struct {
	state         protoimpl.MessageState `protogen:"open.v1"`
	NewMessage    *NewMessage            `protobuf:"bytes,1,opt,name=new_message,json=newMessage,proto3" json:"new_message,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *SendMessageRequest) Reset() {
	*x = SendMessageRequest{}
	mi := &file_transport_grpc_pb_telegram_proto_msgTypes[5]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *SendMessageRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*SendMessageRequest) ProtoMessage() {}

func (x *SendMessageRequest) ProtoReflect() protoreflect.Message {
	mi := &file_transport_grpc_pb_telegram_proto_msgTypes[5]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use SendMessageRequest.ProtoReflect.Descriptor instead.
func (*SendMessageRequest) Descriptor() ([]byte, []int) {
	return file_transport_grpc_pb_telegram_proto_rawDescGZIP(), []int{5}
}

func (x *SendMessageRequest) GetNewMessage() *NewMessage {
	if x != nil {
		return x.NewMessage
	}
	return nil
}

type SendMessageAlbumRequest struct {
	state         protoimpl.MessageState `protogen:"open.v1"`
	NewMessages   []*NewMessage          `protobuf:"bytes,1,rep,name=new_messages,json=newMessages,proto3" json:"new_messages,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *SendMessageAlbumRequest) Reset() {
	*x = SendMessageAlbumRequest{}
	mi := &file_transport_grpc_pb_telegram_proto_msgTypes[6]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *SendMessageAlbumRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*SendMessageAlbumRequest) ProtoMessage() {}

func (x *SendMessageAlbumRequest) ProtoReflect() protoreflect.Message {
	mi := &file_transport_grpc_pb_telegram_proto_msgTypes[6]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use SendMessageAlbumRequest.ProtoReflect.Descriptor instead.
func (*SendMessageAlbumRequest) Descriptor() ([]byte, []int) {
	return file_transport_grpc_pb_telegram_proto_rawDescGZIP(), []int{6}
}

func (x *SendMessageAlbumRequest) GetNewMessages() []*NewMessage {
	if x != nil {
		return x.NewMessages
	}
	return nil
}

type ForwardMessageRequest struct {
	state         protoimpl.MessageState `protogen:"open.v1"`
	ChatId        int64                  `protobuf:"varint,1,opt,name=chat_id,json=chatId,proto3" json:"chat_id,omitempty"`
	MessageId     int64                  `protobuf:"varint,2,opt,name=message_id,json=messageId,proto3" json:"message_id,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *ForwardMessageRequest) Reset() {
	*x = ForwardMessageRequest{}
	mi := &file_transport_grpc_pb_telegram_proto_msgTypes[7]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *ForwardMessageRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ForwardMessageRequest) ProtoMessage() {}

func (x *ForwardMessageRequest) ProtoReflect() protoreflect.Message {
	mi := &file_transport_grpc_pb_telegram_proto_msgTypes[7]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use ForwardMessageRequest.ProtoReflect.Descriptor instead.
func (*ForwardMessageRequest) Descriptor() ([]byte, []int) {
	return file_transport_grpc_pb_telegram_proto_rawDescGZIP(), []int{7}
}

func (x *ForwardMessageRequest) GetChatId() int64 {
	if x != nil {
		return x.ChatId
	}
	return 0
}

func (x *ForwardMessageRequest) GetMessageId() int64 {
	if x != nil {
		return x.MessageId
	}
	return 0
}

type MessageResponse struct {
	state         protoimpl.MessageState `protogen:"open.v1"`
	Message       *Message               `protobuf:"bytes,1,opt,name=message,proto3" json:"message,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *MessageResponse) Reset() {
	*x = MessageResponse{}
	mi := &file_transport_grpc_pb_telegram_proto_msgTypes[8]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *MessageResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*MessageResponse) ProtoMessage() {}

func (x *MessageResponse) ProtoReflect() protoreflect.Message {
	mi := &file_transport_grpc_pb_telegram_proto_msgTypes[8]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use MessageResponse.ProtoReflect.Descriptor instead.
func (*MessageResponse) Descriptor() ([]byte, []int) {
	return file_transport_grpc_pb_telegram_proto_rawDescGZIP(), []int{8}
}

func (x *MessageResponse) GetMessage() *Message {
	if x != nil {
		return x.Message
	}
	return nil
}

type GetMessageRequest struct {
	state         protoimpl.MessageState `protogen:"open.v1"`
	ChatId        int64                  `protobuf:"varint,1,opt,name=chat_id,json=chatId,proto3" json:"chat_id,omitempty"`
	MessageId     int64                  `protobuf:"varint,2,opt,name=message_id,json=messageId,proto3" json:"message_id,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *GetMessageRequest) Reset() {
	*x = GetMessageRequest{}
	mi := &file_transport_grpc_pb_telegram_proto_msgTypes[9]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *GetMessageRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*GetMessageRequest) ProtoMessage() {}

func (x *GetMessageRequest) ProtoReflect() protoreflect.Message {
	mi := &file_transport_grpc_pb_telegram_proto_msgTypes[9]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use GetMessageRequest.ProtoReflect.Descriptor instead.
func (*GetMessageRequest) Descriptor() ([]byte, []int) {
	return file_transport_grpc_pb_telegram_proto_rawDescGZIP(), []int{9}
}

func (x *GetMessageRequest) GetChatId() int64 {
	if x != nil {
		return x.ChatId
	}
	return 0
}

func (x *GetMessageRequest) GetMessageId() int64 {
	if x != nil {
		return x.MessageId
	}
	return 0
}

type UpdateMessageRequest struct {
	state         protoimpl.MessageState `protogen:"open.v1"`
	Message       *Message               `protobuf:"bytes,1,opt,name=message,proto3" json:"message,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *UpdateMessageRequest) Reset() {
	*x = UpdateMessageRequest{}
	mi := &file_transport_grpc_pb_telegram_proto_msgTypes[10]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *UpdateMessageRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*UpdateMessageRequest) ProtoMessage() {}

func (x *UpdateMessageRequest) ProtoReflect() protoreflect.Message {
	mi := &file_transport_grpc_pb_telegram_proto_msgTypes[10]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use UpdateMessageRequest.ProtoReflect.Descriptor instead.
func (*UpdateMessageRequest) Descriptor() ([]byte, []int) {
	return file_transport_grpc_pb_telegram_proto_rawDescGZIP(), []int{10}
}

func (x *UpdateMessageRequest) GetMessage() *Message {
	if x != nil {
		return x.Message
	}
	return nil
}

type DeleteMessagesRequest struct {
	state         protoimpl.MessageState `protogen:"open.v1"`
	ChatId        int64                  `protobuf:"varint,1,opt,name=chat_id,json=chatId,proto3" json:"chat_id,omitempty"`
	MessageIds    []int64                `protobuf:"varint,2,rep,packed,name=message_ids,json=messageIds,proto3" json:"message_ids,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *DeleteMessagesRequest) Reset() {
	*x = DeleteMessagesRequest{}
	mi := &file_transport_grpc_pb_telegram_proto_msgTypes[11]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *DeleteMessagesRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*DeleteMessagesRequest) ProtoMessage() {}

func (x *DeleteMessagesRequest) ProtoReflect() protoreflect.Message {
	mi := &file_transport_grpc_pb_telegram_proto_msgTypes[11]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use DeleteMessagesRequest.ProtoReflect.Descriptor instead.
func (*DeleteMessagesRequest) Descriptor() ([]byte, []int) {
	return file_transport_grpc_pb_telegram_proto_rawDescGZIP(), []int{11}
}

func (x *DeleteMessagesRequest) GetChatId() int64 {
	if x != nil {
		return x.ChatId
	}
	return 0
}

func (x *DeleteMessagesRequest) GetMessageIds() []int64 {
	if x != nil {
		return x.MessageIds
	}
	return nil
}

type GetMessageLinkRequest struct {
	state         protoimpl.MessageState `protogen:"open.v1"`
	ChatId        int64                  `protobuf:"varint,1,opt,name=chat_id,json=chatId,proto3" json:"chat_id,omitempty"`
	MessageId     int64                  `protobuf:"varint,2,opt,name=message_id,json=messageId,proto3" json:"message_id,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *GetMessageLinkRequest) Reset() {
	*x = GetMessageLinkRequest{}
	mi := &file_transport_grpc_pb_telegram_proto_msgTypes[12]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *GetMessageLinkRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*GetMessageLinkRequest) ProtoMessage() {}

func (x *GetMessageLinkRequest) ProtoReflect() protoreflect.Message {
	mi := &file_transport_grpc_pb_telegram_proto_msgTypes[12]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use GetMessageLinkRequest.ProtoReflect.Descriptor instead.
func (*GetMessageLinkRequest) Descriptor() ([]byte, []int) {
	return file_transport_grpc_pb_telegram_proto_rawDescGZIP(), []int{12}
}

func (x *GetMessageLinkRequest) GetChatId() int64 {
	if x != nil {
		return x.ChatId
	}
	return 0
}

func (x *GetMessageLinkRequest) GetMessageId() int64 {
	if x != nil {
		return x.MessageId
	}
	return 0
}

type MessageLinkResponse struct {
	state         protoimpl.MessageState `protogen:"open.v1"`
	Link          string                 `protobuf:"bytes,1,opt,name=link,proto3" json:"link,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *MessageLinkResponse) Reset() {
	*x = MessageLinkResponse{}
	mi := &file_transport_grpc_pb_telegram_proto_msgTypes[13]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *MessageLinkResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*MessageLinkResponse) ProtoMessage() {}

func (x *MessageLinkResponse) ProtoReflect() protoreflect.Message {
	mi := &file_transport_grpc_pb_telegram_proto_msgTypes[13]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use MessageLinkResponse.ProtoReflect.Descriptor instead.
func (*MessageLinkResponse) Descriptor() ([]byte, []int) {
	return file_transport_grpc_pb_telegram_proto_rawDescGZIP(), []int{13}
}

func (x *MessageLinkResponse) GetLink() string {
	if x != nil {
		return x.Link
	}
	return ""
}

type GetMessageLinkInfoRequest struct {
	state         protoimpl.MessageState `protogen:"open.v1"`
	Link          string                 `protobuf:"bytes,1,opt,name=link,proto3" json:"link,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *GetMessageLinkInfoRequest) Reset() {
	*x = GetMessageLinkInfoRequest{}
	mi := &file_transport_grpc_pb_telegram_proto_msgTypes[14]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *GetMessageLinkInfoRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*GetMessageLinkInfoRequest) ProtoMessage() {}

func (x *GetMessageLinkInfoRequest) ProtoReflect() protoreflect.Message {
	mi := &file_transport_grpc_pb_telegram_proto_msgTypes[14]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use GetMessageLinkInfoRequest.ProtoReflect.Descriptor instead.
func (*GetMessageLinkInfoRequest) Descriptor() ([]byte, []int) {
	return file_transport_grpc_pb_telegram_proto_rawDescGZIP(), []int{14}
}

func (x *GetMessageLinkInfoRequest) GetLink() string {
	if x != nil {
		return x.Link
	}
	return ""
}

type EmptyResponse struct {
	state         protoimpl.MessageState `protogen:"open.v1"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *EmptyResponse) Reset() {
	*x = EmptyResponse{}
	mi := &file_transport_grpc_pb_telegram_proto_msgTypes[15]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *EmptyResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*EmptyResponse) ProtoMessage() {}

func (x *EmptyResponse) ProtoReflect() protoreflect.Message {
	mi := &file_transport_grpc_pb_telegram_proto_msgTypes[15]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use EmptyResponse.ProtoReflect.Descriptor instead.
func (*EmptyResponse) Descriptor() ([]byte, []int) {
	return file_transport_grpc_pb_telegram_proto_rawDescGZIP(), []int{15}
}

var File_transport_grpc_pb_telegram_proto protoreflect.FileDescriptor

const file_transport_grpc_pb_telegram_proto_rawDesc = "" +
	"\n" +
	" transport/grpc/pb/telegram.proto\x12\x02pb\"\x85\x01\n" +
	"\n" +
	"NewMessage\x12\x17\n" +
	"\achat_id\x18\x02 \x01(\x03R\x06chatId\x12\x12\n" +
	"\x04text\x18\x01 \x01(\tR\x04text\x12-\n" +
	"\x13reply_to_message_id\x18\x03 \x01(\x03R\x10replyToMessageId\x12\x1b\n" +
	"\tfile_path\x18\x04 \x01(\tR\bfilePath\"}\n" +
	"\aMessage\x12\x0e\n" +
	"\x02id\x18\x01 \x01(\x03R\x02id\x12\x17\n" +
	"\achat_id\x18\x02 \x01(\x03R\x06chatId\x12\x12\n" +
	"\x04text\x18\x03 \x01(\tR\x04text\x12\x18\n" +
	"\aforward\x18\x04 \x01(\bR\aforward\x12\x1b\n" +
	"\tfile_path\x18\x05 \x01(\tR\bfilePath\"N\n" +
	"\x12GetMessagesRequest\x12\x17\n" +
	"\achat_id\x18\x01 \x01(\x03R\x06chatId\x12\x1f\n" +
	"\vmessage_ids\x18\x02 \x03(\x03R\n" +
	"messageIds\"\x86\x01\n" +
	"\x15GetChatHistoryRequest\x12\x17\n" +
	"\achat_id\x18\x01 \x01(\x03R\x06chatId\x12&\n" +
	"\x0ffrom_message_id\x18\x02 \x01(\x03R\rfromMessageId\x12\x16\n" +
	"\x06offset\x18\x03 \x01(\x05R\x06offset\x12\x14\n" +
	"\x05limit\x18\x04 \x01(\x05R\x05limit\";\n" +
	"\x10MessagesResponse\x12'\n" +
	"\bmessages\x18\x01 \x03(\v2\v.pb.MessageR\bmessages\"E\n" +
	"\x12SendMessageRequest\x12/\n" +
	"\vnew_message\x18\x01 \x01(\v2\x0e.pb.NewMessageR\n" +
	"newMessage\"L\n" +
	"\x17SendMessageAlbumRequest\x121\n" +
	"\fnew_messages\x18\x01 \x03(\v2\x0e.pb.NewMessageR\vnewMessages\"O\n" +
	"\x15ForwardMessageRequest\x12\x17\n" +
	"\achat_id\x18\x01 \x01(\x03R\x06chatId\x12\x1d\n" +
	"\n" +
	"message_id\x18\x02 \x01(\x03R\tmessageId\"8\n" +
	"\x0fMessageResponse\x12%\n" +
	"\amessage\x18\x01 \x01(\v2\v.pb.MessageR\amessage\"K\n" +
	"\x11GetMessageRequest\x12\x17\n" +
	"\achat_id\x18\x01 \x01(\x03R\x06chatId\x12\x1d\n" +
	"\n" +
	"message_id\x18\x02 \x01(\x03R\tmessageId\"=\n" +
	"\x14UpdateMessageRequest\x12%\n" +
	"\amessage\x18\x01 \x01(\v2\v.pb.MessageR\amessage\"Q\n" +
	"\x15DeleteMessagesRequest\x12\x17\n" +
	"\achat_id\x18\x01 \x01(\x03R\x06chatId\x12\x1f\n" +
	"\vmessage_ids\x18\x02 \x03(\x03R\n" +
	"messageIds\"O\n" +
	"\x15GetMessageLinkRequest\x12\x17\n" +
	"\achat_id\x18\x01 \x01(\x03R\x06chatId\x12\x1d\n" +
	"\n" +
	"message_id\x18\x02 \x01(\x03R\tmessageId\")\n" +
	"\x13MessageLinkResponse\x12\x12\n" +
	"\x04link\x18\x01 \x01(\tR\x04link\"/\n" +
	"\x19GetMessageLinkInfoRequest\x12\x12\n" +
	"\x04link\x18\x01 \x01(\tR\x04link\"\x0f\n" +
	"\rEmptyResponse2\x92\x05\n" +
	"\n" +
	"FacadeGRPC\x12;\n" +
	"\vGetMessages\x12\x16.pb.GetMessagesRequest\x1a\x14.pb.MessagesResponse\x12A\n" +
	"\x0eGetChatHistory\x12\x19.pb.GetChatHistoryRequest\x1a\x14.pb.MessagesResponse\x128\n" +
	"\vSendMessage\x12\x16.pb.SendMessageRequest\x1a\x11.pb.EmptyResponse\x12B\n" +
	"\x10SendMessageAlbum\x12\x1b.pb.SendMessageAlbumRequest\x1a\x11.pb.EmptyResponse\x12>\n" +
	"\x0eForwardMessage\x12\x19.pb.ForwardMessageRequest\x1a\x11.pb.EmptyResponse\x128\n" +
	"\n" +
	"GetMessage\x12\x15.pb.GetMessageRequest\x1a\x13.pb.MessageResponse\x12<\n" +
	"\rUpdateMessage\x12\x18.pb.UpdateMessageRequest\x1a\x11.pb.EmptyResponse\x12>\n" +
	"\x0eDeleteMessages\x12\x19.pb.DeleteMessagesRequest\x1a\x11.pb.EmptyResponse\x12D\n" +
	"\x0eGetMessageLink\x12\x19.pb.GetMessageLinkRequest\x1a\x17.pb.MessageLinkResponse\x12H\n" +
	"\x12GetMessageLinkInfo\x12\x1d.pb.GetMessageLinkInfoRequest\x1a\x13.pb.MessageResponseB-Z+github.com/comerc/budva43/transport/grpc/pbb\x06proto3"

var (
	file_transport_grpc_pb_telegram_proto_rawDescOnce sync.Once
	file_transport_grpc_pb_telegram_proto_rawDescData []byte
)

func file_transport_grpc_pb_telegram_proto_rawDescGZIP() []byte {
	file_transport_grpc_pb_telegram_proto_rawDescOnce.Do(func() {
		file_transport_grpc_pb_telegram_proto_rawDescData = protoimpl.X.CompressGZIP(unsafe.Slice(unsafe.StringData(file_transport_grpc_pb_telegram_proto_rawDesc), len(file_transport_grpc_pb_telegram_proto_rawDesc)))
	})
	return file_transport_grpc_pb_telegram_proto_rawDescData
}

var file_transport_grpc_pb_telegram_proto_msgTypes = make([]protoimpl.MessageInfo, 16)
var file_transport_grpc_pb_telegram_proto_goTypes = []any{
	(*NewMessage)(nil),                // 0: pb.NewMessage
	(*Message)(nil),                   // 1: pb.Message
	(*GetMessagesRequest)(nil),        // 2: pb.GetMessagesRequest
	(*GetChatHistoryRequest)(nil),     // 3: pb.GetChatHistoryRequest
	(*MessagesResponse)(nil),          // 4: pb.MessagesResponse
	(*SendMessageRequest)(nil),        // 5: pb.SendMessageRequest
	(*SendMessageAlbumRequest)(nil),   // 6: pb.SendMessageAlbumRequest
	(*ForwardMessageRequest)(nil),     // 7: pb.ForwardMessageRequest
	(*MessageResponse)(nil),           // 8: pb.MessageResponse
	(*GetMessageRequest)(nil),         // 9: pb.GetMessageRequest
	(*UpdateMessageRequest)(nil),      // 10: pb.UpdateMessageRequest
	(*DeleteMessagesRequest)(nil),     // 11: pb.DeleteMessagesRequest
	(*GetMessageLinkRequest)(nil),     // 12: pb.GetMessageLinkRequest
	(*MessageLinkResponse)(nil),       // 13: pb.MessageLinkResponse
	(*GetMessageLinkInfoRequest)(nil), // 14: pb.GetMessageLinkInfoRequest
	(*EmptyResponse)(nil),             // 15: pb.EmptyResponse
}
var file_transport_grpc_pb_telegram_proto_depIdxs = []int32{
	1,  // 0: pb.MessagesResponse.messages:type_name -> pb.Message
	0,  // 1: pb.SendMessageRequest.new_message:type_name -> pb.NewMessage
	0,  // 2: pb.SendMessageAlbumRequest.new_messages:type_name -> pb.NewMessage
	1,  // 3: pb.MessageResponse.message:type_name -> pb.Message
	1,  // 4: pb.UpdateMessageRequest.message:type_name -> pb.Message
	2,  // 5: pb.FacadeGRPC.GetMessages:input_type -> pb.GetMessagesRequest
	3,  // 6: pb.FacadeGRPC.GetChatHistory:input_type -> pb.GetChatHistoryRequest
	5,  // 7: pb.FacadeGRPC.SendMessage:input_type -> pb.SendMessageRequest
	6,  // 8: pb.FacadeGRPC.SendMessageAlbum:input_type -> pb.SendMessageAlbumRequest
	7,  // 9: pb.FacadeGRPC.ForwardMessage:input_type -> pb.ForwardMessageRequest
	9,  // 10: pb.FacadeGRPC.GetMessage:input_type -> pb.GetMessageRequest
	10, // 11: pb.FacadeGRPC.UpdateMessage:input_type -> pb.UpdateMessageRequest
	11, // 12: pb.FacadeGRPC.DeleteMessages:input_type -> pb.DeleteMessagesRequest
	12, // 13: pb.FacadeGRPC.GetMessageLink:input_type -> pb.GetMessageLinkRequest
	14, // 14: pb.FacadeGRPC.GetMessageLinkInfo:input_type -> pb.GetMessageLinkInfoRequest
	4,  // 15: pb.FacadeGRPC.GetMessages:output_type -> pb.MessagesResponse
	4,  // 16: pb.FacadeGRPC.GetChatHistory:output_type -> pb.MessagesResponse
	15, // 17: pb.FacadeGRPC.SendMessage:output_type -> pb.EmptyResponse
	15, // 18: pb.FacadeGRPC.SendMessageAlbum:output_type -> pb.EmptyResponse
	15, // 19: pb.FacadeGRPC.ForwardMessage:output_type -> pb.EmptyResponse
	8,  // 20: pb.FacadeGRPC.GetMessage:output_type -> pb.MessageResponse
	15, // 21: pb.FacadeGRPC.UpdateMessage:output_type -> pb.EmptyResponse
	15, // 22: pb.FacadeGRPC.DeleteMessages:output_type -> pb.EmptyResponse
	13, // 23: pb.FacadeGRPC.GetMessageLink:output_type -> pb.MessageLinkResponse
	8,  // 24: pb.FacadeGRPC.GetMessageLinkInfo:output_type -> pb.MessageResponse
	15, // [15:25] is the sub-list for method output_type
	5,  // [5:15] is the sub-list for method input_type
	5,  // [5:5] is the sub-list for extension type_name
	5,  // [5:5] is the sub-list for extension extendee
	0,  // [0:5] is the sub-list for field type_name
}

func init() { file_transport_grpc_pb_telegram_proto_init() }
func file_transport_grpc_pb_telegram_proto_init() {
	if File_transport_grpc_pb_telegram_proto != nil {
		return
	}
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: unsafe.Slice(unsafe.StringData(file_transport_grpc_pb_telegram_proto_rawDesc), len(file_transport_grpc_pb_telegram_proto_rawDesc)),
			NumEnums:      0,
			NumMessages:   16,
			NumExtensions: 0,
			NumServices:   1,
		},
		GoTypes:           file_transport_grpc_pb_telegram_proto_goTypes,
		DependencyIndexes: file_transport_grpc_pb_telegram_proto_depIdxs,
		MessageInfos:      file_transport_grpc_pb_telegram_proto_msgTypes,
	}.Build()
	File_transport_grpc_pb_telegram_proto = out.File
	file_transport_grpc_pb_telegram_proto_goTypes = nil
	file_transport_grpc_pb_telegram_proto_depIdxs = nil
}
