package auto_answers

import (
	"testing"

	"github.com/comerc/budva43/test_e2e"
	"github.com/cucumber/godog"
)

type FeatureState struct {
	test_e2e.CommonState
}

func (f *FeatureState) sendQuestionMessage() error {
	return nil
}

func (f *FeatureState) botRepliesAutomatically() error {
	return nil
}

func InitializeScenario(ctx *godog.ScenarioContext) {
	state := &FeatureState{}
	test_e2e.RegisterCommonSteps(ctx, &state.CommonState)
	ctx.Step(`^пользователь отправляет сообщение с вопросом$`, state.sendQuestionMessage)
	ctx.Step(`^бот автоматически отвечает на сообщение$`, state.botRepliesAutomatically)
}

func Test(t *testing.T) {
	t.Parallel()

	test_e2e.RunFeature(t, "auto_answers", InitializeScenario)
}
