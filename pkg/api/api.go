package api

import (
	"github.com/Burmuley/domovoi/pkg/state"
)

//go:generate oapi-codegen -package=api -generate=types -o ./model.types.go ./openapi.yaml

type Api struct {
	db *state.State
}

func NewApi(db *state.State) *Api {
	return &Api{db: db}
}
