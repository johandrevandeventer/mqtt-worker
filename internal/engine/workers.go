package engine

import (
	"encoding/json"
	"strings"

	"github.com/johandrevandeventer/kafkaclient/payload"
	"github.com/johandrevandeventer/logging"
	"github.com/johandrevandeventer/mqtt-worker/internal/flags"
	mqttworker "github.com/johandrevandeventer/mqtt-worker/internal/workers/mqtt_worker"
	"github.com/johandrevandeventer/mqtt-worker/internal/workers/types"
	"go.uber.org/zap"
)

func (e *Engine) startWorker() {
	e.logger.Info("Starting MQTT workers")

	var workersLogger *zap.Logger
	var kafkaProducerLogger *zap.Logger
	if flags.FlagWorkersLogging {
		workersLogger = logging.GetLogger("workers")
		kafkaProducerLogger = logging.GetLogger("kafka.producer")
	} else {
		workersLogger = zap.NewNop()
		kafkaProducerLogger = zap.NewNop()
	}

	for {
		select {
		case <-e.ctx.Done(): // Handle context cancellation (e.g., Ctrl+C)
			e.logger.Info("Stopping worker due to context cancellation")
			return
		case data, ok := <-e.kafkaConsumer.GetOutputChannel():
			if !ok { // Channel is closed
				e.logger.Info("Kafka consumer output channel closed, stopping worker")
				return
			}

			deserializedData, err := payload.Deserialize(data)
			if err != nil {
				e.logger.Error("Failed to deserialize data", zap.Error(err))
				continue
			}

			worker := mqttworker.NewWorker(workersLogger)

			messageInfo, err := worker.RunWorker(data)
			if err != nil {
				if strings.Contains(err.Error(), "controller is ignored") {
					errorSplit := strings.Split(err.Error(), "controller is ignored: ")
					controllerID := errorSplit[1]
					e.logger.Warn("Controller is ignored", zap.String("controllerID", controllerID))
				} else if strings.Contains(err.Error(), "device is ignored") {
					errorSplit := strings.Split(err.Error(), "device is ignored: ")
					deviceID := errorSplit[1]
					e.logger.Warn("Device is ignored", zap.String("deviceID", deviceID))
				} else if strings.Contains(err.Error(), "device not found") {
					errorSplit := strings.Split(err.Error(), "device not found: ")
					errorSplit = strings.Split(errorSplit[1], " - ")
					deviceID := errorSplit[0]
					deviceName := errorSplit[1]
					e.logger.Warn("Device not found", zap.String("deviceID", deviceID), zap.String("deviceName", deviceName))
				} else {
					e.logger.Error("Processing failed", zap.Error(err))
				}
				continue
			}

			for _, device := range messageInfo.Devices {
				rawDataStruct := &types.DataStruct{
					State:                "Pre",
					CustomerID:           device.CustomerID,
					CustomerName:         device.CustomerName,
					SiteID:               device.SiteID,
					SiteName:             device.SiteName,
					Controller:           device.Controller,
					DeviceType:           device.DeviceType,
					ControllerIdentifier: device.ControllerIdentifier,
					DeviceName:           device.DeviceName,
					DeviceIdentifier:     device.DeviceIdentifier,
					Data:                 device.RawData,
					Timestamp:            device.Timestamp,
				}

				processedDataStruct := &types.DataStruct{
					State:                "Post",
					CustomerID:           device.CustomerID,
					CustomerName:         device.CustomerName,
					SiteID:               device.SiteID,
					SiteName:             device.SiteName,
					Controller:           device.Controller,
					DeviceType:           device.DeviceType,
					ControllerIdentifier: device.ControllerIdentifier,
					DeviceName:           device.DeviceName,
					DeviceIdentifier:     device.DeviceIdentifier,
					Data:                 device.ProcessedData,
					Timestamp:            device.Timestamp,
				}

				serializedRawData, err := json.Marshal(rawDataStruct)
				if err != nil {
					workersLogger.Error("Failed to serialize raw data", zap.Error(err))
					return
				}

				serializedProcessedData, err := json.Marshal(processedDataStruct)
				if err != nil {
					workersLogger.Error("Failed to serialize processed data", zap.Error(err))
					return
				}

				rp := payload.Payload{
					ID:               deserializedData.ID,
					Message:          serializedRawData,
					MessageTimestamp: rawDataStruct.Timestamp,
				}

				pp := payload.Payload{
					ID:               deserializedData.ID,
					Message:          serializedProcessedData,
					MessageTimestamp: processedDataStruct.Timestamp,
				}

				serializedRp, err := rp.Serialize()
				if err != nil {
					workersLogger.Error("Failed to serialize raw payload", zap.Error(err))
					return
				}

				serializedPp, err := pp.Serialize()
				if err != nil {
					workersLogger.Error("Failed to serialize processed payload", zap.Error(err))
					return
				}

				influxdb_kafka_topic := "rubicon_kafka_influxdb"
				kodelabs_kafka_topic := "rubicon_kafka_kodelabs"

				if flags.FlagEnvironment == "development" {
					influxdb_kafka_topic = "rubicon_kafka_influxdb_development"
					kodelabs_kafka_topic = "rubicon_kafka_kodelabs_development"
				}

				// Send the processed data to the Kafka producer
				err = e.kafkaProducerPool.SendMessage(e.ctx, influxdb_kafka_topic, serializedRp)
				if err != nil {
					kafkaProducerLogger.Error("Failed to send raw data to Kafka", zap.Error(err))
					return
				}

				err = e.kafkaProducerPool.SendMessage(e.ctx, influxdb_kafka_topic, serializedPp)
				if err != nil {
					kafkaProducerLogger.Error("Failed to send processed data to Kafka", zap.Error(err))
					return
				}

				err = e.kafkaProducerPool.SendMessage(e.ctx, kodelabs_kafka_topic, serializedPp)
				if err != nil {
					kafkaProducerLogger.Error("Failed to send processed data to Kafka", zap.Error(err))
					return
				}
			}
		}
	}
}
