package workers

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/johandrevandeventer/devicesdb"
	"github.com/johandrevandeventer/devicesdb/models"
	"github.com/johandrevandeventer/mqtt-worker/internal/workers/types"
)

type Payload struct {
	MqttTopic        string    `json:"mqtt_topic"`
	Message          []byte    `json:"message"`
	MessageTimestamp time.Time `json:"message_timestamp"`
}

// Serialize converts the Payload to a byte slice (e.g., JSON)
func (p *Payload) Serialize() ([]byte, error) {
	return json.Marshal(p)
}

// Deserialize converts a byte slice to a Payload
func Deserialize(data []byte) (*Payload, error) {
	var p Payload
	err := json.Unmarshal(data, &p)
	return &p, err
}

func TrimPrefix(s, prefix string) string {
	if len(s) < len(prefix) {
		return s
	}
	return s[len(prefix):]
}

func getCustomerFromTopic(topic string) (string, error) {
	// Split the topic
	topicParts := strings.Split(topic, "/")
	if len(topicParts) < 2 {
		return "", fmt.Errorf("invalid topic: %s", topic)
	}

	// Return the customer
	return topicParts[0], nil
}

// Helper function to get database instance
// GetDBInstance returns the database instance or handles the error.
func getDBInstance() (*devicesdb.BMS_DB, error) {
	bmsDB, err := devicesdb.GetDB()
	if err != nil {
		return nil, err
	}
	return bmsDB, nil
}

// Helper function to get all customers
func GetAllCustomers() ([]models.Customer, error) {
	bmsDB, err := getDBInstance()
	if err != nil {
		return nil, err
	}

	var customers []models.Customer
	if err := bmsDB.DB.Find(&customers).Error; err != nil {
		return nil, fmt.Errorf("failed to get customers: %w", err)
	}

	return customers, nil
}

// Helper function to get all devices
func GetAllDevices() ([]models.Device, error) {
	bmsDB, err := getDBInstance()
	if err != nil {
		return nil, err
	}

	var devices []models.Device
	if err := bmsDB.DB.Find(&devices).Error; err != nil {
		return nil, fmt.Errorf("failed to get devices: %w", err)
	}

	return devices, nil
}

// Helper function to get devices by controller identifier
func GetDevicesByControllerIdentifier(controllerIdentifier string) ([]models.Device, error) {
	bmsDB, err := getDBInstance()
	if err != nil {
		return nil, err
	}

	var devices []models.Device
	if err := bmsDB.DB.Preload("Site.Customer").Where("controller_identifier = ?", controllerIdentifier).Find(&devices).Error; err != nil {
		return nil, fmt.Errorf("failed to get devices: %w", err)
	}

	return devices, nil
}

// Helper function to get device by device identifier
func GetDevicesByDeviceIdentifier(deviceIdentifier string) (models.Device, error) {
	bmsDB, err := getDBInstance()
	if err != nil {
		return models.Device{}, err
	}

	var device models.Device
	if err := bmsDB.DB.Preload("Site.Customer").Where("device_identifier = ?", deviceIdentifier).First(&device).Error; err != nil {
		return models.Device{}, fmt.Errorf("failed to get device: %w", err)
	}

	return device, nil
}

// Helper function to validate and retrieve customer
func GetValidCustomer(topic string) (string, error) {
	customer, err := getCustomerFromTopic(topic)
	if err != nil {
		return "", fmt.Errorf("failed to get customer: %w", err)
	}

	customers, err := GetAllCustomers()
	if err != nil {
		return "", fmt.Errorf("failed to get customers: %w", err)
	}

	for _, c := range customers {
		if strings.EqualFold(c.Name, customer) {
			return customer, nil
		}
	}

	return "", fmt.Errorf("customer not found: %s", customer)
}

// Helper function to read ignored controllers from json file
func readIgnoredFile() (*types.IgnoredControllersAndDevices, error) {
	// Open the JSON file
	file, err := os.Open("./internal/workers/ignored/ignored.json")
	if err != nil {
		return nil, fmt.Errorf("error opening file: %w", err)
	}
	defer file.Close()

	// Decode the JSON file into the Config struct
	var ignoredControllersAndDevices types.IgnoredControllersAndDevices
	decoder := json.NewDecoder(file)
	err = decoder.Decode(&ignoredControllersAndDevices)
	if err != nil {
		return nil, fmt.Errorf("error decoding JSON: %w", err)
	}

	return &ignoredControllersAndDevices, nil
}

// Helper function to read and return ignored controllers from json file
func GetIgnoredControllers() ([]string, error) {
	ignoredControllersAndDevices, err := readIgnoredFile()
	if err != nil {
		return nil, fmt.Errorf("failed to read ignored controllers: %w", err)
	}

	return ignoredControllersAndDevices.IgnoredControllers, nil
}

// Helper function to read and return ignored devices from json file
func GetIgnoredDevices() ([]string, error) {
	ignoredControllersAndDevices, err := readIgnoredFile()
	if err != nil {
		return nil, fmt.Errorf("failed to read ignored devices: %w", err)
	}

	return ignoredControllersAndDevices.IgnoredDevices, nil
}

func IsEmpty(s types.DataStruct) bool {
	return s.State == "" && s.CustomerID == uuid.Nil && s.CustomerName == "" && s.SiteID == uuid.Nil && s.SiteName == "" && s.Controller == "" && s.DeviceType == "" && s.ControllerIdentifier == "" && s.DeviceName == "" && s.DeviceIdentifier == "" && s.Data == nil && s.Timestamp.IsZero()
}
