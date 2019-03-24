package main

import (
	"context"
	rpio "github.com/stianeikeland/go-rpio"
	"log"
)

// GpioWriter
type GpioWriter struct {
	// C is the channel for changing state of the GPIO
	C      chan bool
	config GpioWriterCfg
}

// NewGpioWriter takes a configuration value and returns a GpioWriter instance
func NewGpioWriter(c GpioWriterCfg) *GpioWriter {
	g := &GpioWriter{
		C:      make(chan bool, 32),
		config: c,
	}
	return g
}

// Start contains the main loop for the GpioWriter, should be run as a goroutine
func (g *GpioWriter) Start(ctx context.Context) {
	pin := rpio.Pin(g.config.Pin)
	pin.Output()

	for {
		select {
		case <-ctx.Done():
			return
		case st := <-g.C:
			if st == g.config.Invert {
				log.Printf("[gpioOut:%d] Setting pin to high", g.config.Pin)
				pin.High()
			} else {
				log.Printf("[gpioOut:%d] Setting pin to low", g.config.Pin)
				pin.Low()
			}
		}
	}
}
