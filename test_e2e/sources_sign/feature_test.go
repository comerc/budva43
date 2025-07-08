package sources_sign

import (
	"testing"

	"github.com/comerc/budva43/test_e2e"
	"github.com/cucumber/godog"
)

type FeatureState struct {
	test_e2e.CommonState
}

func (f *FeatureState) messageAppearsWithSourceSign() error {
	return nil
}

func InitializeScenario(ctx *godog.ScenarioContext) {
	state := &FeatureState{}
	test_e2e.RegisterCommonSteps(ctx, &state.CommonState)
	ctx.Step(`^сообщение появляется в целевом чате с подписью источника$`, state.messageAppearsWithSourceSign)
}

func TestFeature(t *testing.T) {
	t.Parallel()

	test_e2e.RunFeature(t, "sources_sign", InitializeScenario)
}
