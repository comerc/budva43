package facade_gql

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/zelenin/go-tdlib/client"

	"github.com/comerc/budva43/app/dto/gql/dto"
	"github.com/comerc/budva43/app/util"
	"github.com/comerc/budva43/service/facade_gql/mocks"
)

func TestGetStatus(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		setup   func(t *testing.T, tg *mocks.TelegramRepo)
		want    *dto.Status
		wantErr bool
	}{
		{
			name: "success",
			setup: func(t *testing.T, tg *mocks.TelegramRepo) {
				opt := &client.OptionValueString{Value: "tdlib-v1.2.3"}
				tg.EXPECT().GetOption(&client.GetOptionRequest{Name: "version"}).Return(opt, nil)
				user := &client.User{Id: 12345}
				tg.EXPECT().GetMe().Return(user, nil)
			},
			want: &dto.Status{
				ReleaseVersion: util.GetReleaseVersion(),
				TdlibVersion:   "tdlib-v1.2.3",
				UserId:         int64(12345),
			},
			wantErr: false,
		},
		{
			name: "option_error",
			setup: func(t *testing.T, tg *mocks.TelegramRepo) {
				tg.EXPECT().GetOption(&client.GetOptionRequest{Name: "version"}).Return(nil, errors.New("fail opt"))
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "getme_error",
			setup: func(t *testing.T, tg *mocks.TelegramRepo) {
				opt := &client.OptionValueString{Value: "tdlib-v1.2.3"}
				tg.EXPECT().GetOption(&client.GetOptionRequest{Name: "version"}).Return(opt, nil)
				tg.EXPECT().GetMe().Return(nil, errors.New("fail me"))
			},
			want:    nil,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			tg := mocks.NewTelegramRepo(t)
			if tt.setup != nil {
				tt.setup(t, tg)
			}
			s := New(tg)
			status, err := s.GetStatus()
			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, status)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.want, status)
			}
		})
	}
}
