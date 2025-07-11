package e2e

import (
	"context"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/cucumber/godog"
	gonanoid "github.com/matoous/go-nanoid/v2"
	"google.golang.org/grpc"

	"github.com/comerc/budva43/transport/grpc/pb"
)

var client pb.FacadeGRPCClient

type scenario struct {
	state *scenarioState
}

type scenarioState struct {
	sourceChatId int64
	sourceText   string

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

func stringToInt64(s string) int64 {
	var result int64
	fmt.Sscanf(s, "%d", &result)
	return result
}

func (s *scenario) setSourceChat(chatId, name string) error {
	var err error

	s.state.sourceChatId = stringToInt64(chatId)

	var key string
	key, err = gonanoid.New(5)
	if err != nil {
		return fmt.Errorf("failed to generate nanoid: %w", err)
	}
	s.state.sourceText = fmt.Sprintf("%s %s", name, key)

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

func (s *scenario) setExpectedText(suffix string) error {
	// s.ExpectedText = s.SendText + strings.ReplaceAll(suffix, `\n`, "\n")
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

func (s *scenario) setExpectedForward(mode string) error {
	fmt.Println("setExpectedMessages", mode)
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

func (s *scenario) setExpectedSuffix(v string) error {
	s.state.checks = append(s.state.checks, func(message *pb.Message) error {
		expected := strings.ReplaceAll(v, `\n`, "\n")
		if !strings.HasSuffix(message.Text, expected) {
			return fmt.Errorf("message text has no suffix: want %q, got %q", expected, message.Text)
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
		Text:   s.state.sourceText,
	})
	if err != nil {
		return fmt.Errorf("failed to send text message via grpc: %w", err)
	}

	return nil
}

func (s *scenario) checkMessage(ctx context.Context, chatId, name string) error {
	var err error

	var resp *pb.MessageResponse
	resp, err = client.GetLastMessage(ctx, &pb.GetLastMessageRequest{
		ChatId: stringToInt64(chatId),
	})
	if err != nil {
		return fmt.Errorf("failed to get last message in destination chat: %w", err)
	}

	if !strings.HasPrefix(resp.Message.Text, s.state.sourceText) {
		return fmt.Errorf("message text has no prefix: want %q, got %q", s.state.sourceText, resp.Message.Text)
	}

	for _, check := range s.state.checks {
		err = check(resp.Message)
		if err != nil {
			return err
		}
	}

	return nil
}

func registerSteps(ctx *godog.ScenarioContext) {
	scenario := &scenario{}
	// !! зарегистрированные раньше имеют приоритет выполнения
	ctx.Given(`^исходный чат "([^"]*)" \(([^)]+)\)$`, scenario.setSourceChat)
	// ctx.Given(`^целевой чат "([^"]*)" \(([^)]+)\)$`, state.setDestinationChat)
	// ctx.Given(`^отправляемый текст \"\[id\]\ \[src_chat_name\]\"$`, state.setSendText)
	// ctx.Given(`^ожидаемый текст \"\[id\]\ \[src_chat_name\]\" \+ "([^"]*)"$`, state.setExpectedText)
	// ctx.When(`^пользователь отправляет исходное сообщение$`, state.sendMessage)
	// ctx.When(`^пользователь пересылает последнее сообщение$`, state.forwardLastMessage)
	// ctx.Then(`^медиа-альбом как копия$`, state.checkAlbumAppearsAsCopy)
	// ctx.Then(`^медиа-альбом как форвард$`, state.checkAlbumAppearsAsForward)
	// ctx.Then(`^сообщение как копия$`, state.checkMessageAppearsAsCopy)
	// ctx.Then(`^сообщение как форвард$`, state.checkMessageAppearsAsForward)
	// ctx.Then(`^сообщение появляется в целевом чате$`, state.checkMessageAppearsInTargetChat)
	// ctx.Then(`^сообщение не появляется в целевом чате$`, state.checkMessageDoesNotAppearInTargetChat)
	// ctx.Then(`^сообщение равно ожидаемому тексту$`, state.checkMessageEqualsExpectedText)
	ctx.Given(`^будет пересылка - ([^"]*)$`, scenario.setExpectedForward)
	ctx.When(`^пользователь отправляет сообщение в исходный чат$`, scenario.sendMessage)
	ctx.Then(`^пауза (\d+) сек.$`, scenario.sleep)
	ctx.Then(`^сообщение в чате "([^"]*)" \("([^"]*)"\)$`, scenario.checkMessage)
	ctx.Given(`^будет суффикс "([^"]*)"$`, scenario.setExpectedSuffix)
	ctx.Before(func(ctx context.Context, sc *godog.Scenario) (context.Context, error) {
		scenario.state = &scenarioState{}
		return ctx, nil
	})
}

func Test(t *testing.T) {
	t.Parallel()

	if testing.Short() {
		t.Skip()
	}

	// TODO: запустить сервер

	names := []string{
		// "01.forward_send_copy", // OK
		// "02.forward", // OK
		// "03.1.replace_myself_links",
		// "03.2.delete_external_links",
		// "04.1.filters_mode_exclude",
		// "04.2.filters_mode_include",
		// "05.media_album_send_copy",
		// "06.media_album_forward",
		// "07.1.include_submatch_f",
		// "07.2.include_submatch_t",
		// "08.replace_fragments",
		"09.sources_link_title",
		// "10.sources_sign", // OK
		// "11.auto_answers",
		// "12.copy_once",
		// "13.indelible",
		// "14.media_album_copy_once",
		// "15.media_album_indelible",
	}

	// TODO: доставать из конфига? + v6
	grpcAddr := ":50051"
	conn, err := grpc.Dial(grpcAddr, grpc.WithInsecure())
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
