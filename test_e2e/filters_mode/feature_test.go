package filters_mode

import (
	"testing"

	"github.com/comerc/budva43/test_e2e"
	"github.com/cucumber/godog"
)

type FeatureState struct {
	test_e2e.CommonState
}

func (f *FeatureState) sendMessageWithIncludeTag() error {
	return nil
}

func (f *FeatureState) sendMessageWithExcludeTag() error {
	return nil
}

func InitializeScenario(ctx *godog.ScenarioContext) {
	state := &FeatureState{}
	test_e2e.RegisterCommonSteps(ctx, &state.CommonState)
	ctx.Step(`^пользователь отправляет сообщение с тегом "#ARK"$`, state.sendMessageWithIncludeTag)
	ctx.Step(`^пользователь отправляет сообщение с тегом "#УТРЕННИЙ_ОБЗОР"$`, state.sendMessageWithExcludeTag)
}

func Test(t *testing.T) {
	t.Parallel()

	test_e2e.RunFeature(t, "filters_mode", InitializeScenario)
}
