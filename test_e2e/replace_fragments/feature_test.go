package replace_fragments

import (
	"testing"

	"github.com/comerc/budva43/test_e2e"
	"github.com/cucumber/godog"
)

type FeatureState struct {
	test_e2e.CommonState
}

func (f *FeatureState) setDestinationChatWithReplaceFragments(chatID, name string) error {
	return nil
}

func (f *FeatureState) sendMessageWithText(from string) error {
	return nil
}

func (f *FeatureState) messageAppearsWithText(to string) error {
	return nil
}

func InitializeScenario(ctx *godog.ScenarioContext) {
	state := &FeatureState{}
	test_e2e.RegisterCommonSteps(ctx, &state.CommonState)
	ctx.Step(`^целевой чат "([^"]*)" \(([^)]+)\) с replace-fragments$`, state.setDestinationChatWithReplaceFragments)
	ctx.Step(`^пользователь отправляет сообщение с текстом "([^"]*)"$`, state.sendMessageWithText)
	ctx.Step(`^сообщение появляется в целевом чате с текстом "([^"]*)"$`, state.messageAppearsWithText)
}

func Test(t *testing.T) {
	test_e2e.RunFeature(t, "replace_fragments", InitializeScenario)
}
