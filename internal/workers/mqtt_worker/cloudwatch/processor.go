package cloudwatch

import (
	"encoding/json"
	"fmt"
	"slices"
	"strings"
	"time"

	"github.com/johandrevandeventer/kafkaclient/payload"
	"github.com/johandrevandeventer/mqtt-worker/internal/workers"
	"github.com/johandrevandeventer/mqtt-worker/internal/workers/mqtt_worker/cloudwatch/powermeter"
	"github.com/johandrevandeventer/mqtt-worker/internal/workers/types"
	"go.uber.org/zap"
)

const (
	DeviceTypePowermeter = "powermeter"
)

func Processor(msg payload.Payload, logger *zap.Logger) (MessageInfo *types.MessageInfo, err error) {
	var cloudWatchInfo CloudWatch
	if err := json.Unmarshal(msg.Message, &cloudWatchInfo); err != nil {
		return MessageInfo, fmt.Errorf("failed to unmarshal payload: %w", err)
	}

	var controllerID string
	var deviceID string

	controllerID = cloudWatchInfo.DeviceIdentifier

	logger.Debug("Processing controller", zap.String("controllerID", controllerID))

	ignoredControllers, err := workers.GetIgnoredControllers()
	if err != nil {
		return MessageInfo, fmt.Errorf("error getting ignored controllers: %w", err)
	}

	if slices.Contains(ignoredControllers, controllerID) {
		return MessageInfo, fmt.Errorf("controller is ignored: %s", controllerID)
	}

	deviceID = controllerID

	logger.Debug("Processing device", zap.String("deviceID", deviceID))

	ignoredDevices, err := workers.GetIgnoredDevices()
	if err != nil {
		return MessageInfo, fmt.Errorf("error getting ignored devices: %w", err)
	}

	if slices.Contains(ignoredDevices, deviceID) {
		return MessageInfo, fmt.Errorf("device is ignored: %s", deviceID)
	}

	device, err := workers.GetDevicesByDeviceIdentifier(deviceID)
	if err != nil {
		if strings.Contains(err.Error(), "record not found") {
			return MessageInfo, fmt.Errorf("device not found: %s - %s", deviceID, cloudWatchInfo.DeviceName)
		}

		return MessageInfo, fmt.Errorf("error getting device by device ID - %s: %w", deviceID, err)
	}

	deviceType := device.DeviceType
	deviceTypeLower := strings.ToLower(deviceType)
	t := cloudWatchInfo.Timestamp

	timestamp, err := time.Parse("2006-01-02T15:04:05.000", t)
	if err != nil {
		return MessageInfo, fmt.Errorf("error parsing timestamp: %w", err)
	}

	fmt.Println("Timestamp:", timestamp)

	var data map[string]any
	err = json.Unmarshal(msg.Message, &data)
	if err != nil {
		return MessageInfo, fmt.Errorf("error unmarshalling payload: %w", err)
	}

	delete(data, "device_identifier")
	delete(data, "device_name")
	delete(data, "timestamp")

	var devices []types.Device
	var rawData map[string]any
	var processedData map[string]any

	switch deviceTypeLower {
	// Process Genset devices
	case DeviceTypePowermeter:
		logger.Debug(fmt.Sprintf("%s :: %s", device.Controller, device.DeviceType))

		rawData, processedData, err := powermeter.Decoder(data)
		if err != nil {
			return MessageInfo, fmt.Errorf("error decoding genset data: %w", err)
		}

		rawData["SerialNo1"] = device.ControllerIdentifier
		processedData["SerialNo1"] = device.ControllerIdentifier
	}

	deviceStruct := &types.Device{
		CustomerID:           device.Site.Customer.ID,
		CustomerName:         device.Site.Customer.Name,
		SiteID:               device.Site.ID,
		SiteName:             device.Site.Name,
		Controller:           device.Controller,
		DeviceType:           device.DeviceType,
		ControllerIdentifier: device.ControllerIdentifier,
		DeviceName:           device.DeviceName,
		DeviceIdentifier:     device.DeviceIdentifier,
		RawData:              rawData,
		ProcessedData:        processedData,
		Timestamp:            timestamp,
	}

	devices = append(devices, *deviceStruct)

	return &types.MessageInfo{
		MessageID: msg.ID.String(),
		Devices:   devices,
	}, nil

}
