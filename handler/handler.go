package handler

import (
	"github.com/TykTechnologies/tyk-pump/config"
)

type PumpsHandler interface {
	SetConfig(config *config.TykPumpConfiguration)
	Init() error
	GetData() []interface{}
	WriteToPumps(data []interface{})
}

func NewHandler(handlerType string) PumpsHandler {
	return &CommonPumpsHandler{}
}
