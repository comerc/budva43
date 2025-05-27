package update_message_send

import (
	"log/slog"

	"github.com/zelenin/go-tdlib/client"
)

type queueRepo interface {
	Add(task func())
}

type storageService interface {
	SetNewMessageId(chatId, tmpMessageId, newMessageId int64) error
	SetTmpMessageId(chatId, newMessageId, tmpMessageId int64) error
}

type Handler struct {
	log *slog.Logger
	//
	queueRepo      queueRepo
	storageService storageService
}

func New(
	queueRepo queueRepo,
	storageService storageService,
) *Handler {
	return &Handler{
		log: slog.With("module", "handler.update_message_send"),
		//
		queueRepo:      queueRepo,
		storageService: storageService,
	}
}

// TODO: можно сделать адаптеры для UpdateMessageSendFailed и UpdateMessageSendAcknowledged

// Run выполняет обрабатку обновления об успешной отправке сообщения
func (h *Handler) Run(update *client.UpdateMessageSendSucceeded) {
	message := update.Message
	tmpMessageId := update.OldMessageId
	fn := func() {
		_ = h.storageService.SetNewMessageId(message.ChatId, tmpMessageId, message.Id)
		_ = h.storageService.SetTmpMessageId(message.ChatId, message.Id, tmpMessageId)
		// h.log.Info("Run")
	}
	h.queueRepo.Add(fn)
}
