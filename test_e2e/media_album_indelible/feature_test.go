package media_album_indelible

import (
	"testing"

	"github.com/comerc/budva43/test_e2e"
	"github.com/cucumber/godog"
)

type FeatureState struct {
	test_e2e.CommonState
}

func (f *FeatureState) setDestinationChatIndelible(chatID, name string) error {
	return nil
}

func (f *FeatureState) albumAppearsAndCannotBeDeleted() error {
	return nil
}

func InitializeScenario(ctx *godog.ScenarioContext) {
	state := &FeatureState{}
	test_e2e.RegisterCommonSteps(ctx, &state.CommonState)
	ctx.Step(`^целевой чат "([^"]*)" \(([^)]+)\) с indelible$`, state.setDestinationChatIndelible)
	ctx.Step(`^альбом появляется в целевом чате и не может быть удалён$`, state.albumAppearsAndCannotBeDeleted)
}

func Test(t *testing.T) {
	t.Parallel()

	test_e2e.RunFeature(t, "media_album_indelible", InitializeScenario)
}
