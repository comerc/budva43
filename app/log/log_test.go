package log

import (
	"errors"
	"log/slog"
	"os"
	"path/filepath"
	"testing"

	"github.com/comerc/budva43/app/util"
	"github.com/comerc/spylog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMain(m *testing.M) {
	// Исключение: используется без app/spylog - циклическая зависимость
	spylog.Init(slog.Default()) // init() app/log before slog.Default()
	os.Exit(m.Run())
}

type SomeObject struct {
	log *Logger
}

func (s *SomeObject) SomeMethod() {
	s.NestedMethod()
}

func (s *SomeObject) NestedMethod() {
	var err error
	defer func() {
		args := []any{
			"arg2", "val2",
		}
		s.log.ErrorOrDebug(&err, "message", args...)
	}()

	err = NewError("error", "arg0", "val0")
	err = WrapError(err, "arg1", "val1")
}

func TestSomeMethod(t *testing.T) {
	// t.Parallel() // !! нельзя параллелить, тестирую с подменой глобальных опций

	optionsCopy, err := util.DeepCopy(options)
	require.NoError(t, err)
	t.Cleanup(func() {
		options = optionsCopy
	})

	tests := []struct {
		name           string
		sourceType     SourceType
		relativePath   bool
		expectedSource string
	}{
		{
			name:           "simple with absolute path",
			sourceType:     TypeSourceOne,
			relativePath:   false,
			expectedSource: filepath.Join(util.ProjectRoot, "app/log/log_test.go:39 log.(*SomeObject).NestedMethod"),
		},
		{
			name:           "one",
			sourceType:     TypeSourceOne,
			relativePath:   true,
			expectedSource: "app/log/log_test.go:39 log.(*SomeObject).NestedMethod",
		},
		{
			name:           "more",
			sourceType:     TypeSourceMore,
			relativePath:   true,
			expectedSource: "[0=app/log/log_test.go:39 log.(*SomeObject).NestedMethod 1=app/log/log_test.go:27 log.(*SomeObject).SomeMethod 2=app/log/log_test.go:88 log.TestSomeMethod.func2]",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			options.ErrorSource.Type = test.sourceType
			options.ErrorSource.RelativePath = test.relativePath

			var o *SomeObject
			spylogHandler := spylog.GetModuleLogHandler("module_name", t.Name(), func() {
				o = &SomeObject{
					log: NewLogger("module_name"),
				}
			})
			o.SomeMethod()

			records := spylogHandler.GetRecords()
			require.Equal(t, len(records), 1)

			assert.Equal(t, slog.LevelError, records[0].Level)
			assert.Equal(t, "error", records[0].Message)
			assert.Equal(t, "val0", spylog.GetAttrValue(records[0], "arg0"))
			assert.Equal(t, "val1", spylog.GetAttrValue(records[0], "arg1"))
			assert.Equal(t, "val2", spylog.GetAttrValue(records[0], "arg2"))
			assert.Equal(t, test.expectedSource, spylog.GetAttrValue(records[0], "source"))
		})
	}

}

type SomeError struct {
	error
}

func TestUnwrappedError(t *testing.T) {
	t.Parallel()

	type OtherObject struct {
		log *Logger
	}

	var err error

	var o *OtherObject
	spylogHandler := spylog.GetModuleLogHandler("module_name", t.Name(), func() {
		o = &OtherObject{
			log: NewLogger("module_name"),
		}
	})

	err = &SomeError{
		error: errors.New("unwrapped error"),
	}
	o.log.ErrorOrInfo(&err, "message", "arg", "val")

	records := spylogHandler.GetRecords()
	require.Equal(t, len(records), 1)

	assert.Equal(t, slog.LevelError, records[0].Level)
	assert.Equal(t, "unwrapped error", records[0].Message)
	assert.Equal(t, "val", spylog.GetAttrValue(records[0], "arg"))
	assert.Equal(t, "log.SomeError", spylog.GetAttrValue(records[0], "type"))
	assert.Equal(t, "app/log/log_test.go:127 log.TestUnwrappedError",
		spylog.GetAttrValue(records[0], "source"))
}
