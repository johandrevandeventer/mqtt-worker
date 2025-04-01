package mqttworker

import (
	"encoding/json"
	"fmt"

	"github.com/johandrevandeventer/mqtt-worker/internal/workers/types"
)

// Decoder handles payload identification
type Decoder struct {
	decoders map[string]func(json.RawMessage) (*types.DecodedPayloadInfo, error)
}

// NewDecoder creates a new Decoder with registered decoders
func NewDecoder() *Decoder {
	return &Decoder{
		decoders: make(map[string]func(json.RawMessage) (*types.DecodedPayloadInfo, error)),
	}
}

// RegisterDecoder adds a new payload decoder
func (d *Decoder) RegisterDecoder(
	name string,
	decoder func(json.RawMessage) (*types.DecodedPayloadInfo, error),
) {
	d.decoders[name] = decoder
}

// DecodePayload processes a message
func (d *Decoder) DecodePayload(payload []byte) (decodedPayloadInfo *types.DecodedPayloadInfo, err error) {
	// Try registered decoders first
	for name, decoder := range d.decoders {
		decodedPayloadInfo, err = decoder(payload)
		if err == nil {
			decodedPayloadInfo.Type = name
			return decodedPayloadInfo, nil
		}
	}

	return decodedPayloadInfo, fmt.Errorf("unknown payload format")
}
