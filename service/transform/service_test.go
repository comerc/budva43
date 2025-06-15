package transform

import (
	"errors"
	"strings"
	"testing"

	"github.com/comerc/budva43/service/transform/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/zelenin/go-tdlib/client"
)

func TestService_addText(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name             string
		formattedText    *client.FormattedText
		text             string
		expectedError    error
		expectedText     string
		expectedEntities []*client.TextEntity
	}{
		{
			name:             "empty",
			formattedText:    &client.FormattedText{},
			text:             "test",
			expectedError:    nil,
			expectedText:     "test",
			expectedEntities: nil,
		},
		{
			name:             "with error",
			formattedText:    &client.FormattedText{},
			text:             "new text",
			expectedError:    errors.New("error"),
			expectedText:     "",
			expectedEntities: nil,
		},
		{
			name: "with existing text",
			formattedText: &client.FormattedText{
				Text:     "existing",
				Entities: []*client.TextEntity{},
			},
			text:             "new text",
			expectedError:    nil,
			expectedText:     "existing\n\nnew text",
			expectedEntities: []*client.TextEntity{},
		},
		{
			name: "add text with entities",
			formattedText: &client.FormattedText{
				Text:     "existing",
				Entities: []*client.TextEntity{},
			},
			text:          "*bold*",
			expectedError: nil,
			expectedText:  "existing\n\n*bold*",
			expectedEntities: []*client.TextEntity{
				{
					Type:   &client.TextEntityTypeBold{},
					Offset: 10,
					Length: 6,
				},
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			telegramRepo := mocks.NewTelegramRepo(t)
			telegramRepo.EXPECT().ParseTextEntities(&client.ParseTextEntitiesRequest{
				Text: test.text,
				ParseMode: &client.TextParseModeMarkdown{
					Version: 2,
				},
			}).Return(&client.FormattedText{
				Text:     test.text,
				Entities: test.expectedEntities,
			}, test.expectedError)

			service := New(telegramRepo, nil, nil)
			service.addText(test.formattedText, test.text)

			assert.Equal(t, test.expectedText, test.formattedText.Text)
			assert.Equal(t, test.expectedEntities, test.formattedText.Entities)
		})
	}
}

func TestEscapeMarkdown(t *testing.T) {
	t.Parallel()

	s1 := "_ * ( ) ~ ` > # + = | { } . !"
	s2 := `\[ \] \-`
	a := strings.Split(s1+" "+s2, " ")
	for _, v := range a {
		assert.Equal(t, `\`+v, escapeMarkdown(v))
	}
}
