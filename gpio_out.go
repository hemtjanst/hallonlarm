package main

import (
	rpio "github.com/stianeikeland/go-rpio"
)

// GpioWriter
type GpioWriter struct {
	// C is the channel for changing state of the GPIO
	C        chan bool
	config   GpioWriterCfg
	stopChan chan bool
}

// NewGpioWriter takes a configuration value and returns a GpioWriter instance
func NewGpioWriter(c GpioWriterCfg) *GpioWriter {
	g := &GpioWriter{
		C:        make(chan bool, 32),
		config:   c,
		stopChan: make(chan bool, 1),
	}
	return g
}

// Stop cancels the GpioWriter
func (g *GpioWriter) Stop() {
	g.stopChan <- true
}

// Start contains the main loop for the GpioWriter, should be run as a goroutine
func (g *GpioWriter) Start() {
	pin := rpio.Pin(g.config.Pin)
	pin.Output()

	for {
		select {
		case <-g.stopChan:
			return
		case st := <-g.C:
			if st == g.config.Invert {
				pin.High()
			} else {
				pin.Low()
			}
		}
	}
}
