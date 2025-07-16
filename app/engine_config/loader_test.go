package engine_config

import (
	"errors"
	"testing"

	"github.com/comerc/budva43/app/domain"
	"github.com/comerc/budva43/app/log"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCheck(t *testing.T) {
	t.Parallel()

	err := check(&domain.EngineConfig{})
	is := errors.Is(err, ErrEmptyConfigData)
	require.True(t, is)

	var customError *log.CustomError
	as := errors.As(err, &customError)
	require.True(t, as)

	assert.Equal(t, err.Error(), "отсутствуют данные")
	assert.Equal(t, customError.Args, []any{
		"path.0",
		"config.Engine.UniqueSources",
		"path.1",
		"config.Engine.UniqueDestinations",
		"path.2",
		"config.Engine.OrderedForwardRules",
	})
}
