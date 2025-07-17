package e2e

import (
	"context"
	"fmt"
	"net"
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

	// DestinationChatId   int64
	// DestinationChatName string

	// SendText     string
	// ExpectedText string
	// LastSentMessageText string
	// LastSentMessageId int64
	// LastSentMessage *pb.Message
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

func (s *scenario) setSourceChat(name string, chatId int) error {
	var err error

	s.state.sourceChatId = -int64(chatId)

	const alphabet = "0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"

	var key string
	key, err = gonanoid.Generate(alphabet, 5)
	if err != nil {
		return fmt.Errorf("failed to generate nanoid: %w", err)
	}
	s.state.sourceTextPrefix = fmt.Sprintf("%s %s", name, key)

	return nil
}

// func (s *state) setDestinationChat(chatId, name string) error {
// 	var id int64
// 	_, err := fmt.Sscanf(chatId, "%d", &id)
// 	if err != nil {
// 		return fmt.Errorf("invalid chatId: %v", err)
// 	}
// 	s.DestinationChatId = id
// 	s.DestinationChatName = name
// 	return nil
// }

func (s *scenario) checkMessageAppearsAsCopy(ctx context.Context) error {
	// if s.LastSentMessage == nil {
	// 	return fmt.Errorf("last sent message is not set")
	// }
	// if s.LastSentMessage.Forward {
	// 	return fmt.Errorf("message is forward")
	// }
	return nil
}

func (s *scenario) setSendText() error {
	// id, err := gonanoid.New()
	// if err != nil {
	// 	return fmt.Errorf("failed to generate nanoid: %w", err)
	// }
	// s.SendText = id + " " + s.SourceChatName
	return nil
}

func (s *scenario) checkMessageEqualsExpectedText(ctx context.Context) error {
	// if s.LastSentMessage == nil {
	// 	return fmt.Errorf("last sent message is not set")
	// }
	// if s.LastSentMessage.Text != s.ExpectedText {
	// 	return fmt.Errorf("message text mismatch: want %q, got %q", s.ExpectedText, s.LastSentMessage.Text)
	// }
	return nil
}

func (s *scenario) checkAlbumAppearsAsCopy() error { return nil }

func (s *scenario) checkAlbumAppearsAsForward() error { return nil }

func (s *scenario) checkMessageAppearsAsForward() error {
	// if s.LastSentMessage == nil {
	// 	return fmt.Errorf("last sent message is not set")
	// }
	// if !s.LastSentMessage.Forward {
	// 	return fmt.Errorf("message is not forward")
	// }
	return nil
}

func (s *scenario) checkMessageAppearsInTargetChat() error {
	// if s.LastSentMessage == nil {
	// 	return fmt.Errorf("last sent message is not set")
	// }
	return nil
}

// func (s *state) sendMessage(ctx context.Context) error {
// 	var err error

// 	var resp *pb.MessageResponse
// 	resp, err = client.GetLastMessage(ctx, &pb.GetLastMessageRequest{
// 		ChatId: s.SourceChatId,
// 	})
// 	if err != nil {
// 		return fmt.Errorf("failed to get last message in destination chat: %w", err)
// 	}
// 	tmpMessage := resp.Message

// 	var msg *pb.MessageResponse
// 	msg, err = client.SendMessage(ctx, &pb.SendMessageRequest{
// 		ChatId: s.SourceChatId,
// 		Text:   s.SendText,
// 	})
// 	_ = msg
// 	if err != nil {
// 		return fmt.Errorf("failed to send text message via grpc: %w", err)
// 	}

// 	time.Sleep(2 * time.Second)

// 	resp, err = client.GetLastMessage(ctx, &pb.GetLastMessageRequest{
// 		ChatId: s.SourceChatId,
// 	})
// 	if err != nil {
// 		return fmt.Errorf("failed to get last message in destination chat: %w", err)
// 	}
// 	if resp.Message != nil && (tmpMessage == nil || resp.Message != tmpMessage) {
// 		s.LastSentMessage = resp.Message
// 	}

// 	return nil
// }

// func (s *state) forwardLastMessage(ctx context.Context) error {
// 	if s.LastSentMessage == nil {
// 		return fmt.Errorf("last sent message is not set")
// 	}

// 	var err error

// 	var resp *pb.MessageResponse
// 	resp, err = client.GetLastMessage(ctx, &pb.GetLastMessageRequest{
// 		ChatId: s.DestinationChatId,
// 	})
// 	if err != nil {
// 		return fmt.Errorf("failed to get last message in destination chat: %w", err)
// 	}
// 	tmpMessage := resp.Message

// 	_, err = client.ForwardMessage(ctx, &pb.ForwardMessageRequest{
// 		ChatId:    s.DestinationChatId,
// 		MessageId: s.LastSentMessage.Id,
// 	})
// 	if err != nil {
// 		return fmt.Errorf("failed to forward message: %w", err)
// 	}

// 	resp, err = client.GetLastMessage(ctx, &pb.GetLastMessageRequest{
// 		ChatId: s.DestinationChatId,
// 	})
// 	if err != nil {
// 		return fmt.Errorf("failed to get last message in destination chat: %w", err)
// 	}
// 	if resp.Message != tmpMessage {
// 		s.LastSentMessage = resp.Message
// 	}

// 	return nil
// }

// func (s *state) checkMessageDoesNotAppearInTargetChat(ctx context.Context) error {
// 	if s.LastSentMessage != nil {
// 		return fmt.Errorf("last sent message is not nil")
// 	}
// 	return nil
// }

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

func extractExpectedLink(text string) string {
	pattern := `>>>\[(.*)\]\((.*)\)<<<`
	re := regexp.MustCompile(pattern)
	matches := re.FindStringSubmatch(text)
	if len(matches) != 3 {
		return ""
	}
	return matches[2]
}

func (s *scenario) addCheckWithExpectedLinkToMessage(ctx context.Context) error {
	s.state.checks = append(s.state.checks, func(message *pb.Message) error {
		link := extractExpectedLink(message.Text)
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

func (s *scenario) sleep(ctx context.Context, seconds int) error {
	time.Sleep(time.Duration(seconds) * time.Second)
	return nil
}

func (s *scenario) sendMessage(ctx context.Context) error {
	var err error

	_, err = client.SendMessage(ctx, &pb.SendMessageRequest{
		ChatId: s.state.sourceChatId,
		Text:   fmt.Sprintf("%s\n\n%s", s.state.sourceTextPrefix, s.state.sourceText),
	})
	if err != nil {
		return fmt.Errorf("failed to send text message via grpc: %w", err)
	}

	return nil
}

func (s *scenario) sendYetiMessage(ctx context.Context) error {
	var err error

	_, err = client.SendMessage(ctx, &pb.SendMessageRequest{
		ChatId: s.state.sourceChatId,
		Text:   util.EscapeMarkdown(domain.YETI_MESSAGE),
	})
	if err != nil {
		return fmt.Errorf("failed to send text message via grpc: %w", err)
	}

	return nil
}

func (s *scenario) checkSourceMessage(ctx context.Context) error {
	var err error

	var resp *pb.MessageResponse
	resp, err = client.GetLastMessage(ctx, &pb.GetLastMessageRequest{
		ChatId: s.state.sourceChatId,
	})
	if err != nil {
		return fmt.Errorf("failed to get last message: %w", err)
	}

	if !strings.HasPrefix(resp.Message.Text, s.state.sourceTextPrefix) {
		return fmt.Errorf("message text has no prefix: want %q, got %q",
			s.state.sourceTextPrefix, resp.Message.Text)
	}

	return nil
}

func (s *scenario) checkMessage(ctx context.Context, name string, chatId int) error {
	var err error

	var resp *pb.MessageResponse
	resp, err = client.GetLastMessage(ctx, &pb.GetLastMessageRequest{
		ChatId: -int64(chatId),
	})
	if err != nil {
		return fmt.Errorf("failed to get last message: %w", err)
	}

	if !strings.HasPrefix(resp.Message.Text, s.state.sourceTextPrefix) {
		return fmt.Errorf("message text has no prefix: want %q, got %q",
			s.state.sourceTextPrefix, resp.Message.Text)
	}

	for _, check := range s.state.checks {
		err = check(resp.Message)
		if err != nil {
			return err
		}
	}

	return nil
}

func (s *scenario) checkYetiMessage(ctx context.Context) error {
	var err error

	var resp *pb.MessageResponse
	resp, err = client.GetLastMessage(ctx, &pb.GetLastMessageRequest{
		ChatId: s.state.sourceChatId,
	})
	if err != nil {
		return fmt.Errorf("failed to get last message: %w", err)
	}

	if resp.Message.Text != domain.YETI_MESSAGE {
		return fmt.Errorf("message text mismatch: want %q, got %q",
			domain.YETI_MESSAGE, resp.Message.Text)
	}

	return nil
}

func (s *scenario) setExpectedLinkToLastMessage(ctx context.Context) error {
	var err error

	var respMessage *pb.MessageResponse
	respMessage, err = client.GetLastMessage(ctx, &pb.GetLastMessageRequest{
		ChatId: s.state.sourceChatId,
	})
	if err != nil {
		return fmt.Errorf("failed to get last message: %w", err)
	}

	var resp *pb.GetMessageLinkResponse
	resp, err = client.GetMessageLink(ctx, &pb.GetMessageLinkRequest{
		ChatId:    s.state.sourceChatId,
		MessageId: respMessage.Message.Id,
	})
	if err != nil {
		return fmt.Errorf("failed to get message link: %w", err)
	}

	text := fmt.Sprintf(">>>%s<<<", resp.Link)
	s.state.sourceText = util.EscapeMarkdown(text)

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
	pattern := fmt.Sprintf(`>>>%s<<<`, domain.DELETED_LINK)
	return s.addCheckWithExpectedRegex(pattern)
}

func (s *scenario) setExpectedText(ctx context.Context, text string) error {
	s.state.sourceText = text
	return nil
}

func registerSteps(ctx *godog.ScenarioContext) {
	scenario := &scenario{}
	// !! зарегистрированные раньше имеют приоритет выполнения
	ctx.Given(`^исходный чат "([^"]*)" \((\d+)\)$`, scenario.setSourceChat)
	// ctx.Given(`^целевой чат "([^"]*)" \(([^)]+)\)$`, state.setDestinationChat)
	// ctx.Given(`^отправляемый текст \"\[id\]\ \[src_chat_name\]\"$`, state.setSendText)
	// ctx.When(`^пользователь отправляет исходное сообщение$`, state.sendMessage)
	// ctx.When(`^пользователь пересылает последнее сообщение$`, state.forwardLastMessage)
	// ctx.Then(`^медиа-альбом как копия$`, state.checkAlbumAppearsAsCopy)
	// ctx.Then(`^медиа-альбом как форвард$`, state.checkAlbumAppearsAsForward)
	// ctx.Then(`^сообщение как копия$`, state.checkMessageAppearsAsCopy)
	// ctx.Then(`^сообщение как форвард$`, state.checkMessageAppearsAsForward)
	// ctx.Then(`^сообщение появляется в целевом чате$`, state.checkMessageAppearsInTargetChat)
	// ctx.Then(`^сообщение не появляется в целевом чате$`, state.checkMessageDoesNotAppearInTargetChat)
	// ctx.Then(`^сообщение равно ожидаемому тексту$`, state.checkMessageEqualsExpectedText)
	ctx.Given(`^будет пересылка - ([^"]*)$`, scenario.addCheckWithExpectedForward)
	ctx.When(`^пользователь отправляет сообщение$`, scenario.sendMessage)
	ctx.When(`^пользователь отправляет YETI_MESSAGE$`, scenario.sendYetiMessage)
	ctx.Then(`^пауза (\d+) сек.$`, scenario.sleep)
	ctx.Then(`^сообщение в чате$`, scenario.checkSourceMessage)
	ctx.Then(`^сообщение в чате "([^"]*)" \((\d+)\)$`, scenario.checkMessage)
	ctx.Then(`^YETI_MESSAGE в чате$`, scenario.checkYetiMessage)
	ctx.Given(`^будет текст "([^"]*)"$`, scenario.addCheckWithExpectedRegex)
	ctx.Given(`^будет подпись$`, scenario.addCheckWithExpectedSign)
	ctx.Given(`^будет ссылка$`, scenario.addCheckWithExpectedLink)
	ctx.Given(`^будет замена: ссылка на YETI_MESSAGE -> DELETED_LINK$`, scenario.addCheckWithExpectedNoExternalLink)
	ctx.Given(`^сообщение со ссылкой на последнее сообщение$`, scenario.setExpectedLinkToLastMessage)
	ctx.Given(`^будет замена ссылки на сообщение в целевом чате$`, scenario.addCheckWithExpectedLinkToMessage)
	ctx.Given(`^сообщение с текстом "([^"]*)"$`, scenario.setExpectedText)
	ctx.Before(func(ctx context.Context, sc *godog.Scenario) (context.Context, error) {
		scenario.state = &scenarioState{}
		return ctx, nil
	})
}

func Test(t *testing.T) {
	// t.Parallel() // !! нельзя параллелить

	if testing.Short() {
		t.Skip()
	}

	names := []string{
		// "01.forward_send_copy",       // OK
		// "02.forward",                 // OK
		// "03.1.replace_myself_links",  // OK
		// "03.2.delete_external_links", // OK
		// "04.1.filters_mode_exclude",
		// "04.2.filters_mode_include",
		// "05.media_album_send_copy",
		// "06.media_album_forward",
		// "07.1.include_submatch_f",
		// "07.2.include_submatch_t",
		// "08.replace_fragments",  // OK
		// "09.sources_link_title", // OK
		// "10.sources_sign",       // OK
		// "11.auto_answers",
		// "12.copy_once",
		// "13.indelible",
		// "14.media_album_copy_once",
		// "15.media_album_indelible",
	}

	addr := net.JoinHostPort(config.Grpc.Host, config.Grpc.Port)
	if util.IsPortFree(addr) {
		t.Fatal("port is not open")
	}
	conn, err := grpc.Dial(addr, grpc.WithInsecure())
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()
	client = pb.NewFacadeGRPCClient(conn)

	for _, name := range names {
		t.Run(name, func(t *testing.T) {
			// t.Parallel() // !! нельзя параллелить, проверяю последнее сообщение в целевом чате
			runFeature(t, name)
		})
	}
}
