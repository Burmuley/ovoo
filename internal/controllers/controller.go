package controllers

import "context"

type Controller interface {
	Start(ctx context.Context) error
}
