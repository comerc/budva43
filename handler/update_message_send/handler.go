package update_message_send

import (
	"github.com/zelenin/go-tdlib/client"

	"github.com/comerc/budva43/app/log"
)

//go:generate mockery --name=queueRepo --exported
type queueRepo interface {
	Add(task func())
}

//go:generate mockery --name=storageService --exported
type storageService interface {
	SetNewMessageId(chatId, tmpMessageId, newMessageId int64)
	SetTmpMessageId(chatId, newMessageId, tmpMessageId int64)
}

type Handler struct {
	log *log.Logger
	//
	queueRepo      queueRepo
	storageService storageService
}

func New(
	queueRepo queueRepo,
	storageService storageService,
) *Handler {
	return &Handler{
		log: log.NewLogger("handler.update_message_send"),
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
		defer h.log.Debug("Run",
			"chatId", message.ChatId,
			"messageId", message.Id,
			"tmpMessageId", tmpMessageId,
		)

		h.storageService.SetNewMessageId(message.ChatId, tmpMessageId, message.Id)
		h.storageService.SetTmpMessageId(message.ChatId, message.Id, tmpMessageId)
	}
	h.queueRepo.Add(fn)
}
