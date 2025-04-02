package powermeter

import (
	"fmt"

	coreutils "github.com/johandrevandeventer/mqtt-worker/utils"
)

// Main struct combining all the smaller structs
type PowerMeterData struct {
	V1      float64 `json:"V1"`
	V2      float64 `json:"V2"`
	V3      float64 `json:"V3"`
	V1Angle float64 `json:"V1Angle"`
	V2Angle float64 `json:"V2Angle"`
	V3Angle float64 `json:"V3Angle"`
	I1      float64 `json:"I1"`
	I2      float64 `json:"I2"`
	I3      float64 `json:"I3"`
	I4      float64 `json:"I4"`
	I1Angle float64 `json:"I1Angle"`
	I2Angle float64 `json:"I2Angle"`
	I3Angle float64 `json:"I3Angle"`
	I4Angle float64 `json:"I4Angle"`
}

func Decoder(payload map[string]any) (rawData, processedData map[string]any, err error) {
	var powerMeterData PowerMeterData

	err = coreutils.DecodeMapToStruct(payload, &powerMeterData)
	if err != nil {
		return rawData, processedData, fmt.Errorf("error decoding PowerMeter data: %w", err)
	}

	rawData = createDataMap(powerMeterData)

	processedData = createDataMap(powerMeterData)

	return rawData, processedData, nil
}

func createDataMap(pm PowerMeterData) map[string]any {
	dataMap := map[string]any{
		"V1":      pm.V1,
		"V2":      pm.V2,
		"V3":      pm.V3,
		"V1Angle": pm.V1Angle,
		"V2Angle": pm.V2Angle,
		"V3Angle": pm.V3Angle,
		"I1":      pm.I1,
		"I2":      pm.I2,
		"I3":      pm.I3,
		"I4":      pm.I4,
		"I1Angle": pm.I1Angle,
		"I2Angle": pm.I2Angle,
		"I3Angle": pm.I3Angle,
		"I4Angle": pm.I4Angle,
	}

	return dataMap
}
