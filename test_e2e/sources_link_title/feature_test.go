package sources_link_title

import (
	"testing"

	"github.com/comerc/budva43/test_e2e"
	"github.com/cucumber/godog"
)

type FeatureState struct {
	test_e2e.CommonState
}

func (f *FeatureState) sendMessageWithSourceLink() error {
	return nil
}

func (f *FeatureState) messageAppearsWithSourceLinkTitle() error {
	return nil
}

func InitializeScenario(ctx *godog.ScenarioContext) {
	state := &FeatureState{}
	test_e2e.RegisterCommonSteps(ctx, &state.CommonState)
	ctx.Step(`^пользователь отправляет сообщение с ссылкой на источник$`, state.sendMessageWithSourceLink)
	ctx.Step(`^сообщение появляется в целевом чате с заголовком источника$`, state.messageAppearsWithSourceLinkTitle)
}

func TestFeature(t *testing.T) {
	t.Parallel()

	test_e2e.RunFeature(t, "sources_link_title", InitializeScenario)
}
