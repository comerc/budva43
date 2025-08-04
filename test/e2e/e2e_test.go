package e2e

import (
	"context"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
	"time"

	godog "github.com/cucumber/godog"
	gonanoid "github.com/matoous/go-nanoid/v2"
	grpc "google.golang.org/grpc"

	config "github.com/comerc/budva43/app/config"
	domain "github.com/comerc/budva43/app/domain"
	util "github.com/comerc/budva43/app/util"
	pb "github.com/comerc/budva43/transport/grpc/pb"
)

var client pb.FacadeGRPCClient

type scenario struct {
	state *scenarioState
}

type scenarioState struct {
	sourceChatId     int64
	sourceTextPrefix string
	sourceText       string

	checks []check
}

type check = func(message *pb.Message) error

func runFeature(t *testing.T, name string) {
	suite := godog.TestSuite{
		ScenarioInitializer: func(ctx *godog.ScenarioContext) {
			registerSteps(ctx)
		},
		Options: &godog.Options{
			Format: "pretty",
			Paths:  []string{fmt.Sprintf("feature/%s.feature", name)},
		},
	}
	if suite.Run() != 0 {
		t.Fail()
	}
}

func (s *scenario) setSourceChat(name string, chatId int64) error {
	var err error

	s.state.sourceChatId = -chatId

	const alphabet = "0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"

	var key string
	key, err = gonanoid.Generate(alphabet, 5)
	if err != nil {
		return fmt.Errorf("failed to generate nanoid: %w", err)
	}
	s.state.sourceTextPrefix = fmt.Sprintf("%s %s", name, key)

	return nil
}

func (s *scenario) addCheckWithExpectedForward(mode string) error {
	s.state.checks = append(s.state.checks, func(message *pb.Message) error {
		switch mode {
		case "копия":
			if message.Forward {
				return fmt.Errorf("message is not a copy")
			}
		case "форвард":
			if !message.Forward {
				return fmt.Errorf("message is not a forward")
			}
		default:
			return fmt.Errorf("invalid mode: %s", mode)
		}
		return nil
	})
	return nil
}

func extractExpectedLink(title, text string) string {
	pattern := fmt.Sprintf(`\[%s\]\((.*)\)`, title)
	re := regexp.MustCompile(pattern)
	matches := re.FindStringSubmatch(text)
	if len(matches) != 2 {
		return ""
	}
	return matches[1]
}

func getPrevMessageId(ctx context.Context, chatId int64) (int64, error) {
	var err error

	var resp *pb.MessagesResponse
	resp, err = client.GetChatHistory(ctx, &pb.GetChatHistoryRequest{
		ChatId: chatId,
		Limit:  2,
	})
	if err != nil {
		return 0, fmt.Errorf("failed to get chat history: %w", err)
	}
	if len(resp.Messages) < 2 {
		return 0, fmt.Errorf("not enough messages: want 2, got %d", len(resp.Messages))
	}

	message := resp.Messages[1]

	return message.Id, nil
}

func (s *scenario) addCheckWithExpectedLinkToPrevMessage(ctx context.Context) error {
	s.state.checks = append(s.state.checks, func(message *pb.Message) error {
		link := extractExpectedLink(domain.PREV_LINK, message.Text)
		if link == "" {
			return fmt.Errorf("link is empty")
		}
		resp, err := client.GetMessageLinkInfo(ctx, &pb.GetMessageLinkInfoRequest{
			Link: link,
		})
		if err != nil {
			return fmt.Errorf("failed to get message link: %w", err)
		}
		if resp.Message.ChatId != message.ChatId {
			return fmt.Errorf("message chat id mismatch: want %d, got %d",
				message.ChatId, resp.Message.ChatId)
		}
		prevMessageId, err := getPrevMessageId(ctx, resp.Message.ChatId)
		if err != nil {
			return fmt.Errorf("failed to get prev message id: %w", err)
		}
		if resp.Message.Id != prevMessageId {
			return fmt.Errorf("message id mismatch: want %d, got %d",
				prevMessageId, resp.Message.Id)
		}
		return nil
	})
	return nil
}

func (s *scenario) addCheckWithExpectedRegex(val string) error {
	s.state.checks = append(s.state.checks, func(message *pb.Message) error {
		pattern := val
		matched, err := regexp.MatchString(pattern, message.Text)
		if err != nil {
			return fmt.Errorf("invalid regex pattern %q: %w", pattern, err)
		}
		if !matched {
			return fmt.Errorf("message text does not match regex: pattern %q, got %q", pattern, message.Text)
		}
		return nil
	})
	return nil
}

func (s *scenario) wait(seconds int) error {
	time.Sleep(time.Duration(seconds) * time.Second)
	return nil
}

func (s *scenario) sendMessage(ctx context.Context) error {
	var err error

	_, err = client.SendMessage(ctx, &pb.SendMessageRequest{
		NewMessage: &pb.NewMessage{
			ChatId:           s.state.sourceChatId,
			Text:             fmt.Sprintf("%s\n\n%s", s.state.sourceTextPrefix, s.state.sourceText),
			ReplyToMessageId: 0,
			FilePath:         "",
		},
	})
	if err != nil {
		return fmt.Errorf("failed to send text message via grpc: %w", err)
	}

	return nil
}

const mediaAlbumSize int32 = 3

func (s *scenario) sendMediaAlbum(ctx context.Context) error {
	var err error

	newMessages := make([]*pb.NewMessage, mediaAlbumSize)
	for i := range mediaAlbumSize {
		relPath := fmt.Sprintf("static/%d.png", i+1)
		absPath := filepath.Join(util.ProjectRoot, relPath)
		newMessages[i] = &pb.NewMessage{
			ChatId:           s.state.sourceChatId,
			ReplyToMessageId: 0,
			FilePath:         absPath,
		}
	}
	newMessages[0].Text = fmt.Sprintf("%s\n\n%s", s.state.sourceTextPrefix, s.state.sourceText)

	_, err = client.SendMessageAlbum(ctx, &pb.SendMessageAlbumRequest{
		NewMessages: newMessages,
	})
	if err != nil {
		return fmt.Errorf("failed to send media album: %w", err)
	}
	return nil
}

func (s *scenario) editMessage(ctx context.Context) error {
	var err error

	var resp *pb.MessagesResponse
	resp, err = client.GetChatHistory(ctx, &pb.GetChatHistoryRequest{
		ChatId: s.state.sourceChatId,
		Limit:  1,
	})
	if err != nil {
		return fmt.Errorf("failed to get last message: %w", err)
	}
	if len(resp.Messages) == 0 {
		return fmt.Errorf("no messages found")
	}

	message := resp.Messages[0]

	_, err = client.UpdateMessage(ctx, &pb.UpdateMessageRequest{
		Message: &pb.Message{
			Id:     message.Id,
			ChatId: s.state.sourceChatId,
			Text:   fmt.Sprintf("%s\n\n%s", s.state.sourceTextPrefix, s.state.sourceText),
		},
	})
	if err != nil {
		return fmt.Errorf("failed to edit message: %w", err)
	}

	return nil
}

func (s *scenario) deleteMessage(ctx context.Context) error {
	var err error

	var resp *pb.MessagesResponse
	resp, err = client.GetChatHistory(ctx, &pb.GetChatHistoryRequest{
		ChatId: s.state.sourceChatId,
		Limit:  1,
	})
	if err != nil {
		return fmt.Errorf("failed to get last message: %w", err)
	}
	if len(resp.Messages) == 0 {
		return fmt.Errorf("no messages found")
	}

	message := resp.Messages[0]

	_, err = client.DeleteMessages(ctx, &pb.DeleteMessagesRequest{
		ChatId:     s.state.sourceChatId,
		MessageIds: []int64{message.Id},
	})
	if err != nil {
		return fmt.Errorf("failed to delete message: %w", err)
	}

	return nil
}

func (s *scenario) checkSourceMessage(ctx context.Context) error {
	var err error

	var resp *pb.MessagesResponse
	resp, err = client.GetChatHistory(ctx, &pb.GetChatHistoryRequest{
		ChatId: s.state.sourceChatId,
		Limit:  1,
	})
	if err != nil {
		return fmt.Errorf("failed to get last message: %w", err)
	}
	if len(resp.Messages) == 0 {
		return fmt.Errorf("no messages found")
	}

	message := resp.Messages[0]

	if !strings.HasPrefix(message.Text, s.state.sourceTextPrefix) {
		return fmt.Errorf("message text has no prefix: want %q, got %q",
			s.state.sourceTextPrefix, message.Text)
	}

	return nil
}

func (s *scenario) checkMessage(ctx context.Context, name string, chatId int64) error {
	var err error

	var resp *pb.MessagesResponse
	resp, err = client.GetChatHistory(ctx, &pb.GetChatHistoryRequest{
		ChatId: -chatId,
		Limit:  1,
	})
	if err != nil {
		return fmt.Errorf("failed to get last message: %w", err)
	}
	if len(resp.Messages) == 0 {
		return fmt.Errorf("no messages found")
	}

	message := resp.Messages[0]

	if !strings.HasPrefix(message.Text, s.state.sourceTextPrefix) {
		return fmt.Errorf("message text has no prefix: want %q, got %q",
			s.state.sourceTextPrefix, message.Text)
	}

	for _, check := range s.state.checks {
		err = check(message)
		if err != nil {
			return err
		}
	}

	return nil
}

func (s *scenario) checkMediaAlbum(ctx context.Context, name string, chatId int64) error {
	var err error

	var resp *pb.MessagesResponse
	resp, err = client.GetChatHistory(ctx, &pb.GetChatHistoryRequest{
		ChatId: -chatId,
		Limit:  mediaAlbumSize,
	})
	if err != nil {
		return fmt.Errorf("failed to get chat history: %w", err)
	}
	if len(resp.Messages) != int(mediaAlbumSize) {
		return fmt.Errorf("expected %d messages, got %d", mediaAlbumSize, len(resp.Messages))
	}

	message := resp.Messages[mediaAlbumSize-1]

	if !strings.HasPrefix(message.Text, s.state.sourceTextPrefix) {
		return fmt.Errorf("message text has no prefix: want %q, got %q",
			s.state.sourceTextPrefix, message.Text)
	}

	for _, check := range s.state.checks {
		err = check(message)
		if err != nil {
			return err
		}
	}

	return nil
}

func (s *scenario) checkNoMessage(ctx context.Context, name string, chatId int64) error {
	var err error

	var resp *pb.MessagesResponse
	resp, err = client.GetChatHistory(ctx, &pb.GetChatHistoryRequest{
		ChatId: -chatId,
		Limit:  1,
	})
	if err != nil {
		return fmt.Errorf("failed to get last message: %w", err)
	}
	if len(resp.Messages) == 0 {
		return nil
	}

	message := resp.Messages[0]

	if strings.HasPrefix(message.Text, s.state.sourceTextPrefix) {
		return fmt.Errorf("found message")
	}

	return nil
}

func (s *scenario) setExpectedLinkToLastMessage(ctx context.Context) error {
	var err error

	var respChatHistory *pb.MessagesResponse
	respChatHistory, err = client.GetChatHistory(ctx, &pb.GetChatHistoryRequest{
		ChatId: s.state.sourceChatId,
		Limit:  1,
	})
	if err != nil {
		return fmt.Errorf("failed to get last message: %w", err)
	}
	if len(respChatHistory.Messages) == 0 {
		return fmt.Errorf("no messages found")
	}

	message := respChatHistory.Messages[0]

	var resp *pb.MessageLinkResponse
	resp, err = client.GetMessageLink(ctx, &pb.GetMessageLinkRequest{
		ChatId:    s.state.sourceChatId,
		MessageId: message.Id,
	})
	if err != nil {
		return fmt.Errorf("failed to get message link: %w", err)
	}

	s.state.sourceText = fmt.Sprintf("[%s](%s)", domain.PREV_LINK, resp.Link)

	return nil
}

const (
	// Константы из engine.e2e.yml // TODO: получать через grpc?
	E2E_SIGN = "**Sign**"
	E2E_LINK = "**Link**"
)

func (s *scenario) addCheckWithExpectedSign() error {
	pattern := fmt.Sprintf(`(?s)^.*\n\n%s.*$`, strings.ReplaceAll(E2E_SIGN, "*", `\*`))
	return s.addCheckWithExpectedRegex(pattern)
}

func (s *scenario) addCheckWithExpectedLink() error {
	pattern := fmt.Sprintf(`(?s)^.*\n\n\[🔗%s\]\(https://t.me/.*\)$`, strings.ReplaceAll(E2E_LINK, "*", `\*`))
	return s.addCheckWithExpectedRegex(pattern)
}

func (s *scenario) addCheckWithExpectedNoExternalLink() error {
	pattern := fmt.Sprintf("~~%s~~", domain.PREV_LINK)
	return s.addCheckWithExpectedRegex(pattern)
}

func (s *scenario) addCheckWithText(text string) error {
	pattern := fmt.Sprintf(`(?s)^.*(%s).*$`, text)
	return s.addCheckWithExpectedRegex(pattern)
}

func (s *scenario) setExpectedText(text string) error {
	s.state.sourceText = text
	return nil
}

func (s *scenario) resetChecks() error {
	s.state.checks = []check{}
	return nil
}

func registerSteps(ctx *godog.ScenarioContext) {
	scenario := &scenario{}
	// !! зарегистрированные раньше имеют приоритет выполнения
	ctx.Given(`^сброс проверок$`, scenario.resetChecks)
	ctx.Given(`^исходный чат "([^"]*)" \((\d+)\)$`, scenario.setSourceChat)
	ctx.Given(`^будет пересылка - ([^"]*)$`, scenario.addCheckWithExpectedForward)
	ctx.Given(`^будет текст "([^"]*)"$`, scenario.addCheckWithText)
	ctx.Given(`^будет подпись$`, scenario.addCheckWithExpectedSign)
	ctx.Given(`^будет ссылка$`, scenario.addCheckWithExpectedLink)
	ctx.Given(`^будет удалена ссылка на сообщение в исходном чате$`, scenario.addCheckWithExpectedNoExternalLink)
	ctx.Given(`^сообщение со ссылкой на последнее сообщение$`, scenario.setExpectedLinkToLastMessage)
	ctx.Given(`^будет ссылка на предыдущее сообщение в целевом чате$`, scenario.addCheckWithExpectedLinkToPrevMessage)
	ctx.Given(`^сообщение с текстом "([^"]*)"$`, scenario.setExpectedText)
	ctx.When(`^пользователь отправляет сообщение$`, scenario.sendMessage)
	ctx.When(`^пользователь редактирует сообщение$`, scenario.editMessage)
	ctx.When(`^пользователь удаляет сообщение$`, scenario.deleteMessage)
	ctx.When(`^пользователь отправляет медиа-альбом$`, scenario.sendMediaAlbum)
	ctx.Then(`^ожидание (\d+) сек.$`, scenario.wait)
	ctx.Then(`^сообщение в чате$`, scenario.checkSourceMessage)
	ctx.Then(`^сообщение в чате "([^"]*)" \((\d+)\)$`, scenario.checkMessage)
	ctx.Then(`^нет сообщения в чате "([^"]*)" \((\d+)\)$`, scenario.checkNoMessage)
	ctx.Then(`^медиа-альбом в чате "([^"]*)" \((\d+)\)$`, scenario.checkMediaAlbum)

	ctx.Before(func(ctx context.Context, sc *godog.Scenario) (context.Context, error) {
		scenario.state = &scenarioState{}
		return ctx, nil
	})
}

func Test(t *testing.T) {
	// t.Parallel() // !! нельзя параллелить

	// TODO: запускать e2e тесты в CI
	if os.Getenv("CI") == "true" {
		t.Skip()
	}

	if testing.Short() {
		t.Skip()
	}

	names := []string{
		"01.send_copy",                // OK
		"02.forward",                  // OK
		"03.1.replace_myself_links",   // OK
		"03.2.delete_external_links",  // OK
		"04.1.1.filters_mode_exclude", // OK
		"04.1.2.filters_mode_exclude", // OK
		"04.2.1.filters_mode_include", // OK
		"04.2.2.filters_mode_include", // OK
		"05.media_album_send_copy",    // OK
		"06.media_album_forward",      // OK
		"07.1.include_submatch",       // OK
		"07.2.include_submatch",       // OK
		"08.replace_fragments",        // OK
		"09.sources_link_title",       // OK
		"10.sources_sign",             // OK
		"12.1.copy_once_t",            // OK
		"12.2.copy_once_f",            // OK
		"13.1.indelible_t",            // OK
		"13.2.indelible_f",            // OK
		// "11.auto_answers", // TODO: R&D
	}

	addr := net.JoinHostPort(config.Grpc.Host, config.Grpc.Port)
	if util.IsPortFree(addr) {
		t.Fatal("port is not open")
	}
	conn, err := grpc.Dial(addr, grpc.WithInsecure())
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		conn.Close()
	})
	client = pb.NewFacadeGRPCClient(conn)

	for _, name := range names {
		t.Run(name, func(t *testing.T) {
			// t.Parallel() // !! нельзя параллелить, проверяю последнее сообщение в целевом чате
			runFeature(t, name)
		})
	}
}
