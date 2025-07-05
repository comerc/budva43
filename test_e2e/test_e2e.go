package test_e2e

import (
	"testing"

	"github.com/cucumber/godog"
)

// List of Features
// **********************
// 01.Forward.SendCopy
// 02.Forward
// 03.ReplaceMyselfLinks
// 04.FiltersMode
// 05.MediaAlbumSendCopy
// 06.MediaAlbumForward
// 07.IncludeSubmatch
// 08.ReplaceFragments
// 09.SourcesLinkTitle
// 10.SourcesSign
// 11.AutoAnswers
// 12.MediaAlbumCopyOnce
// 13.MediaAlbumIndelible
// **********************

type CommonState struct{}

func (c *CommonState) SetSourceChat(chatID, name string) error      { return nil }
func (c *CommonState) SetDestinationChat(chatID, name string) error { return nil }
func (c *CommonState) SendTextMessageToSourceChat() error           { return nil }
func (c *CommonState) CheckMessageContainsSourceText() error        { return nil }
func (c *CommonState) CheckMessageHasNoPrivacyTags() error          { return nil }
func (c *CommonState) SendMediaAlbumToSourceChat() error            { return nil }
func (c *CommonState) CheckAlbumAppearsAsCopy() error               { return nil }
func (c *CommonState) CheckAlbumAppearsAsForward() error            { return nil }
func (c *CommonState) CheckMessageAppearsAsCopy() error             { return nil }
func (c *CommonState) CheckMessageAppearsAsForward() error          { return nil }
func (c *CommonState) CheckMessageAppearsInTargetChat() error       { return nil }
func (c *CommonState) SendMessage() error                           { return nil }
func (c *CommonState) MessageDoesNotAppearInTargetChat() error      { return nil }

func RegisterCommonSteps(ctx *godog.ScenarioContext, state *CommonState) {
	ctx.Step(`^исходный чат "([^"]*)" \(([^)]+)\)$`, state.SetSourceChat)
	ctx.Step(`^целевой чат "([^"]*)" \(([^)]+)\)$`, state.SetDestinationChat)
	ctx.Step(`^пользователь отправляет текстовое сообщение в исходный чат$`, state.SendTextMessageToSourceChat)
	ctx.Step(`^сообщение содержит исходный текст$`, state.CheckMessageContainsSourceText)
	ctx.Step(`^сообщение не содержит тегов приватности$`, state.CheckMessageHasNoPrivacyTags)
	ctx.Step(`^пользователь отправляет медиаальбом в исходный чат$`, state.SendMediaAlbumToSourceChat)
	ctx.Step(`^альбом появляется в целевом чате как копия$`, state.CheckAlbumAppearsAsCopy)
	ctx.Step(`^альбом появляется в целевом чате как форвард$`, state.CheckAlbumAppearsAsForward)
	ctx.Step(`^сообщение появляется в целевом чате как копия$`, state.CheckMessageAppearsAsCopy)
	ctx.Step(`^сообщение появляется в целевом чате как форвард$`, state.CheckMessageAppearsAsForward)
	ctx.Step(`^сообщение появляется в целевом чате$`, state.CheckMessageAppearsInTargetChat)
	ctx.Step(`^пользователь отправляет сообщение$`, state.SendMessage)
	ctx.Step(`^сообщение не появляется в целевом чате$`, state.MessageDoesNotAppearInTargetChat)
}

func RunFeature(t *testing.T, name string, scenarioInitializer func(*godog.ScenarioContext)) {
	suite := godog.TestSuite{
		Name:                name,
		ScenarioInitializer: scenarioInitializer,
		Options: &godog.Options{
			Format: "pretty",
			Paths:  []string{".feature"},
			// Tags:  "", // если нужно
		},
	}
	if suite.Run() != 0 {
		t.Fail()
	}
}
