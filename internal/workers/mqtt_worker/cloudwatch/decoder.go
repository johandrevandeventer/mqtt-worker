package cloudwatch

import (
	"encoding/json"
	"fmt"

	"github.com/johandrevandeventer/mqtt-worker/internal/workers/types"
)

type CloudWatch struct {
	DeviceIdentifier string `json:"device_identifier"`
	DeviceName       string `json:"device_name"`
	Timestamp        string `json:"timestamp"`
}

// Decoder processes MQTT payloads
func Decoder(payload json.RawMessage) (decodedPayloadInfo *types.DecodedPayloadInfo, err error) {
	var data CloudWatch
	if err := json.Unmarshal(payload, &data); err != nil {
		return decodedPayloadInfo, fmt.Errorf("failed to unmarshal payload: %w", err)
	}

	return &types.DecodedPayloadInfo{
		RawPayload: payload,
	}, nil
}
