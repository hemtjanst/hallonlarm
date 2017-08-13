package main

import (
	"github.com/hemtjanst/hemtjanst/device"
)

// Config is the basic struct for the HCL parser, it has a map of devices and a mqtt config
type Config struct {
	Device map[string]*DeviceConfig `json:"device"`
}

// DeviceConfig contains meta-data for a device and it's features
type DeviceConfig struct {
	// Name as exposed to Hemtjänst/homekit, required
	Name         string                    `json:"name"`
	// Manufacturer as exposed to Hemtjänst, optional
	Manufacturer string                    `json:"manufacturer"`
	// Model as exposed to Hemtjänst, optional
	Model        string                    `json:"model"`
	// SerialNumber as exposed to Hemtjänst, optional
	SerialNumber string                    `json:"serialNumber"`
	// Type must be a valid homekit type, see Hemtjänst documentation for valid types
	Type         string                    `json:"type"`
	Feature      map[string]*DeviceFeature `json:"feature"`
}

// DeviceFeature reflects a homekit characteristic
type DeviceFeature struct {
	// Info contains hemtjänst-specific settings like what topics to use and what the min/max/step values are
	Info    *device.Feature
	// GpioIn is the configuration for reading from GPIO to MQTT
	GpioIn  *GpioReaderCfg `json:"gpioIn"`
	// GpioOut is the configuration for writing to GPIO from MQTT
	GpioOut *GpioWriterCfg `json:"gpioOut"`
}

// GpioReaderCfg sets configuration values for the GpioReader instance
type GpioReaderCfg struct {
	// Pin number (according to BCM2835 pinout
	Pin int `json:"pin"`
	// Invert the reading, if true then "1" is sent as "0" to MQTT
	Invert bool `json:"invert"`
	// ReadInterval defines the time (in milliseconds) between reads
	ReadInterval int64 `json:"readInterval"`
	// ConsecutiveReadsForOpenState specifies the number of reads with same value before a open state is reported
	MinReadOpened int64 `json:"minReadOpened"`
	// ConsecutiveReadsForCloseState specifies the number of reads with same value before a close state is reported
	MinReadClosed int64 `json:"minReadClosed"`
}

type GpioWriterCfg struct {
	// Pin number (according to BCM2835 pinout
	Pin int `json:"pin"`
	// Invert the reading, if true then "1" from MQTT sets the GPIO to High
	Invert bool `json:"invert"`
}
