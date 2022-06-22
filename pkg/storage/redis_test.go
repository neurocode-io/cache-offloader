package storage

import (
	"context"
	"errors"
	"net/http"
	"testing"
	"time"

	redis "github.com/go-redis/redis/v8"
	"github.com/golang/mock/gomock"
	"github.com/neurocode-io/cache-offloader/pkg/model"
	"github.com/stretchr/testify/assert"
)

func TestRedisLookup(t *testing.T) {
	ctx := context.Background()
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	tests := []struct {
		name      string
		requestID string
		repo      RedisStorage
		want      *model.Response
		expErr    error
	}{
		{
			name: "stale key found",
			repo: RedisStorage{
				lookupTimeout: lookupTimeout,
				db: func() IRedis {
					mock := NewMockIRedis(ctrl)
					pipeLinerMock := NewMockPipeliner(ctrl)

					mock.EXPECT().TxPipeline().Return(pipeLinerMock)
					pipeLinerMock.EXPECT().Exists(gomock.Any(), "test:alive").Return(&redis.IntCmd{})
					mockResp := &redis.StringCmd{}
					mockResp.SetVal(`{"status":200}`)
					pipeLinerMock.EXPECT().Get(gomock.Any(), "test").Return(mockResp)
					pipeLinerMock.EXPECT().Exec(gomock.Any()).Return(nil, nil)

					return mock
				}(),
			},
			requestID: "test",
			want:      &model.Response{Status: http.StatusOK, StaleValue: model.StaleValue},
			expErr:    nil,
		},
		{
			name: "key not found",
			repo: RedisStorage{
				lookupTimeout: lookupTimeout,
				db: func() IRedis {
					mock := NewMockIRedis(ctrl)
					pipeLinerMock := NewMockPipeliner(ctrl)

					mock.EXPECT().TxPipeline().Return(pipeLinerMock)
					pipeLinerMock.EXPECT().Exists(gomock.Any(), "test:alive").Return(&redis.IntCmd{})
					mockResp := &redis.StringCmd{}
					mockResp.SetErr(redis.Nil)
					pipeLinerMock.EXPECT().Get(gomock.Any(), "test").Return(mockResp)
					pipeLinerMock.EXPECT().Exec(gomock.Any()).Return(nil, nil)

					return mock
				}(),
			},
			requestID: "test",
			want:      nil,
			expErr:    nil,
		},
		{
			name: "lookup error occurred",
			repo: RedisStorage{
				lookupTimeout: lookupTimeout,
				db: func() IRedis {
					mock := NewMockIRedis(ctrl)
					pipeLinerMock := NewMockPipeliner(ctrl)

					mock.EXPECT().TxPipeline().Return(pipeLinerMock)
					pipeLinerMock.EXPECT().Exists(gomock.Any(), "test:alive").Return(&redis.IntCmd{})
					mockResp := &redis.StringCmd{}
					pipeLinerMock.EXPECT().Get(gomock.Any(), "test").Return(mockResp)
					pipeLinerMock.EXPECT().Exec(gomock.Any()).Return(nil, errors.New("test error"))

					return mock
				}(),
			},
			requestID: "test",
			want:      nil,
			expErr:    errors.New("redis-repository: LookUp error: test error"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.repo.LookUp(ctx, tt.requestID)
			assert.Equal(t, tt.want, got)
			if err != nil {
				assert.EqualError(t, err, tt.expErr.Error())
			}
		})
	}
}

func TestRedisStore(t *testing.T) {
	ctx := context.Background()
	ctrl := gomock.NewController(t)
	staleSeconds := 1

	defer ctrl.Finish()

	tests := []struct {
		name      string
		requestID string
		repo      RedisStorage
		resp      *model.Response
		expErr    error
	}{
		{
			name: "store response",
			repo: RedisStorage{
				lookupTimeout:  lookupTimeout,
				staleInSeconds: staleSeconds,
				db: func() IRedis {
					mock := NewMockIRedis(ctrl)
					pipeLinerMock := NewMockPipeliner(ctrl)

					mock.EXPECT().TxPipeline().Return(pipeLinerMock)
					pipeLinerMock.EXPECT().Set(ctx, gomock.Any(), gomock.Any(), expirationTime).Return(nil)
					pipeLinerMock.EXPECT().Set(ctx, "test:alive", model.FreshValue, time.Duration(staleSeconds)*time.Second).Return(nil)
					pipeLinerMock.EXPECT().Exec(ctx).Return(nil, nil)

					return mock
				}(),
			},
			requestID: "test",
			resp:      &model.Response{Status: http.StatusOK, StaleValue: model.StaleValue},
			expErr:    nil,
		},
		{
			name: "store response error",
			repo: RedisStorage{
				lookupTimeout:  lookupTimeout,
				staleInSeconds: staleSeconds,
				db: func() IRedis {
					mock := NewMockIRedis(ctrl)
					pipeLinerMock := NewMockPipeliner(ctrl)

					mock.EXPECT().TxPipeline().Return(pipeLinerMock)
					pipeLinerMock.EXPECT().Set(ctx, gomock.Any(), gomock.Any(), expirationTime).Return(nil)
					pipeLinerMock.EXPECT().Set(ctx, "test:alive", model.FreshValue, time.Duration(staleSeconds)*time.Second).Return(nil)
					pipeLinerMock.EXPECT().Exec(ctx).Return(nil, errors.New("test error"))

					return mock
				}(),
			},
			requestID: "test",
			resp:      &model.Response{Status: http.StatusOK, StaleValue: model.StaleValue},
			expErr:    errors.New("redis-repository: Store error: test error"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.repo.Store(ctx, tt.requestID, tt.resp)
			if err != nil {
				assert.EqualError(t, err, tt.expErr.Error())
			}
		})
	}
}

func TestRedisCheckConnection(t *testing.T) {
	ctx := context.Background()
	ctrl := gomock.NewController(t)

	defer ctrl.Finish()

	tests := []struct {
		name   string
		repo   RedisStorage
		expErr error
	}{
		{
			name: "check connection",
			repo: RedisStorage{
				db: func() IRedis {
					mock := NewMockIRedis(ctrl)
					status := &redis.StatusCmd{}
					status.SetErr(nil)
					mock.EXPECT().Ping(ctx).Return(status)

					return mock
				}(),
			},
			expErr: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.repo.CheckConnection(ctx)
			if err != nil {
				assert.EqualError(t, err, tt.expErr.Error())
			}
		})
	}
}
