package mqttworker

import (
	"fmt"

	"github.com/johandrevandeventer/kafkaclient/payload"
	"github.com/johandrevandeventer/mqtt-worker/internal/workers/types"
	"go.uber.org/zap"
)

// Processor handles payload identification
type Processor struct {
	logger     *zap.Logger
	processors map[string]func(payload.Payload, *zap.Logger) (*types.MessageInfo, error)
}

// NewProcessor creates a new Processor with registered processors
func NewProcessor(logger *zap.Logger) *Processor {
	return &Processor{
		logger:     logger,
		processors: make(map[string]func(payload.Payload, *zap.Logger) (*types.MessageInfo, error)),
	}
}

// RegisterProcessor adds a new payload processor
func (d *Processor) RegisterProcessor(
	name string,
	processor func(payload.Payload, *zap.Logger) (*types.MessageInfo, error),
) {
	d.processors[name] = processor
}

// ProcessPayload processes a message
func (d *Processor) ProcessPayload(name string, msg payload.Payload) (MessageInfo *types.MessageInfo, err error) {
	processor := d.processors[string(name)]

	if processor == nil {
		return MessageInfo, fmt.Errorf("unknown processor: %s", name)
	}

	MessageInfo, err = processor(msg, d.logger)
	if err != nil {
		return MessageInfo, err
	}

	return MessageInfo, nil
}
