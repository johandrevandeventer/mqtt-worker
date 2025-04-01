package types

import (
	"time"

	"github.com/google/uuid"
)

type DecodedPayloadInfo struct {
	Type       string `json:"type"`
	RawPayload []byte `json:"raw_payload"`
}

type IgnoredControllersAndDevices struct {
	IgnoredControllers []string `json:"ignored_controllers"`
	IgnoredDevices     []string `json:"ignored_devices"`
}

type DataStruct struct {
	State                string
	CustomerID           uuid.UUID
	CustomerName         string
	SiteID               uuid.UUID
	SiteName             string
	Gateway              string
	Controller           string
	DeviceType           string
	ControllerIdentifier string
	DeviceName           string
	DeviceIdentifier     string
	Data                 map[string]any
	Timestamp            time.Time
}

// Base message structure
type MessageInfo struct {
	MessageID string `json:"message_id"` // Unique identifier for the message

	// Device data (either single device or multiple under a controller)
	Devices []Device `json:"devices"`
}

// Device information
type Device struct {
	CustomerID           uuid.UUID
	CustomerName         string
	SiteID               uuid.UUID
	SiteName             string
	Gateway              string
	Controller           string
	DeviceType           string
	ControllerIdentifier string
	DeviceName           string
	DeviceIdentifier     string
	RawData              map[string]any
	ProcessedData        map[string]any
	Timestamp            time.Time
}
