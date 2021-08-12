package storage

import (
	"context"

	"dpd.de/idempotency-offloader/pkg/entity"
)

type Repository interface {
	LookUp(context.Context, string) (*entity.ResponseBody, error)
	Store(context.Context, string, *entity.ResponseBody) error
	CheckConnection(context.Context) error
}
