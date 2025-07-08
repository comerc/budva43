package replace_myself_links

import (
	"testing"

	"github.com/comerc/budva43/test_e2e"
	"github.com/cucumber/godog"
)

type FeatureState struct {
	test_e2e.CommonState
}

func (f *FeatureState) setDestinationChatWithReplaceLinks(chatID, name string) error {
	return nil
}

func (f *FeatureState) sendMessageWithOwnLink() error {
	return nil
}

func (f *FeatureState) linkIsReplacedInTargetChat() error {
	return nil
}

func (f *FeatureState) sendMessageWithExternalLink() error {
	return nil
}

func (f *FeatureState) externalLinkIsRemoved() error {
	return nil
}

func InitializeScenario(ctx *godog.ScenarioContext) {
	state := &FeatureState{}
	test_e2e.RegisterCommonSteps(ctx, &state.CommonState)
	ctx.Step(`^целевой чат "([^"]*)" \(([^)]+)\) с replace-myself-links: run=true, delete-external=true$`, state.setDestinationChatWithReplaceLinks)
	ctx.Step(`^пользователь отправляет сообщение с ссылкой на своё сообщение$`, state.sendMessageWithOwnLink)
	ctx.Step(`^ссылка заменяется на новую в целевом чате$`, state.linkIsReplacedInTargetChat)
	ctx.Step(`^пользователь отправляет сообщение с внешней ссылкой$`, state.sendMessageWithExternalLink)
	ctx.Step(`^внешняя ссылка удаляется$`, state.externalLinkIsRemoved)
}

func Test(t *testing.T) {
	t.Parallel()

	test_e2e.RunFeature(t, "replace_myself_links", InitializeScenario)
}
