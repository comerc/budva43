package engine_config

import (
	"errors"
	"testing"

	"github.com/comerc/budva43/app/entity"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCheck(t *testing.T) {
	t.Parallel()

	err := check(&entity.EngineConfig{})
	var emptyConfigData *ErrEmptyConfigData
	ok := errors.As(err, &emptyConfigData)
	require.True(t, ok)

	assert.Equal(t, err.Error(), "отсутствуют данные")
	assert.Equal(t, emptyConfigData.Args, []any{
		"path.0",
		"config.Engine.UniqueSources",
		"path.1",
		"config.Engine.UniqueDestinations",
		"path.2",
		"config.Engine.OrderedForwardRules",
	})
}
