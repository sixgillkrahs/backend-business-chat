package domain

import "context"

type ActionRepository interface {
	GetAllActions(ctx context.Context) ([]Action, error)
}
