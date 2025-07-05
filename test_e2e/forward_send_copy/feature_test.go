package forward_send_copy

import (
	"testing"

	"github.com/comerc/budva43/test_e2e"
	"github.com/cucumber/godog"
)

type FeatureState struct {
	test_e2e.CommonState
}

func InitializeScenario(ctx *godog.ScenarioContext) {
	state := &FeatureState{}
	test_e2e.RegisterCommonSteps(ctx, &state.CommonState)
}

func TestFeature(t *testing.T) {
	test_e2e.RunFeature(t, "forward_send_copy", InitializeScenario)
}
