package include_submatch

import (
	"testing"

	"github.com/comerc/budva43/test_e2e"
	"github.com/cucumber/godog"
)

type FeatureState struct {
	test_e2e.CommonState
}

func (f *FeatureState) sendMessageWithTicker(ticker string) error {
	return nil
}

func (f *FeatureState) sendMessageWithOtherTicker(other string) error {
	return nil
}

func InitializeScenario(ctx *godog.ScenarioContext) {
	state := &FeatureState{}
	test_e2e.RegisterCommonSteps(ctx, &state.CommonState)
	ctx.Step(`^пользователь отправляет сообщение с тикером "([^"]*)"$`, state.sendMessageWithTicker)
	ctx.Step(`^пользователь отправляет сообщение с тикером "([^"]*)"$`, state.sendMessageWithOtherTicker)
}

func TestFeature(t *testing.T) {
	t.Parallel()

	test_e2e.RunFeature(t, "include_submatch", InitializeScenario)
}
