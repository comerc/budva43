package media_album_copy_once

import (
	"testing"

	"github.com/comerc/budva43/test_e2e"
	"github.com/cucumber/godog"
)

type FeatureState struct {
	test_e2e.CommonState
}

func (f *FeatureState) setDestinationChatCopyOnce(chatID, name string) error {
	return nil
}

func (f *FeatureState) albumAppearsOnlyOnce() error {
	return nil
}

func InitializeScenario(ctx *godog.ScenarioContext) {
	state := &FeatureState{}
	test_e2e.RegisterCommonSteps(ctx, &state.CommonState)
	ctx.Step(`^целевой чат "([^"]*)" \(([^)]+)\) с copy-once$`, state.setDestinationChatCopyOnce)
	ctx.Step(`^альбом появляется в целевом чате только один раз$`, state.albumAppearsOnlyOnce)
}

func Test(t *testing.T) {
	t.Parallel()

	test_e2e.RunFeature(t, "media_album_copy_once", InitializeScenario)
}
