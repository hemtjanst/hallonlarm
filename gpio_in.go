package main

import (
	rpio "github.com/stianeikeland/go-rpio"
	"log"
	"time"
)

func init() {
	if err := rpio.Open(); err != nil {
		log.Fatal(err)
	}
}

// GpioReader
type GpioReader struct {
	C                chan bool
	config           GpioReaderCfg
	lastState        rpio.State
	reported         bool
	consecutiveReads int64
	stopChan         chan bool
}

// NewGpioReader takes a configuration value and returns a GpioReader instance
func NewGpioReader(c GpioReaderCfg) *GpioReader {
	if c.ReadInterval <= 0 {
		c.ReadInterval = 50
	}
	if c.MinReadOpened <= 0 {
		c.MinReadOpened = 1
	}
	if c.MinReadClosed <= 0 {
		c.MinReadClosed = 1
	}
	g := &GpioReader{
		C:                make(chan bool, 32),
		config:           c,
		lastState:        2,
		reported:         false,
		consecutiveReads: 0,
		stopChan:         make(chan bool, 1),
	}
	return g
}

func (g *GpioReader) Stop() {
	g.stopChan <- true
}

func (g *GpioReader) Start() {
	pin := rpio.Pin(g.config.Pin)
	pin.Input()

	var openValue rpio.State = 0
	if g.config.Invert {
		openValue = 1
	}

	tick := time.NewTicker(time.Millisecond * time.Duration(g.config.ReadInterval))

	for {
		state := pin.Read()
		switch state {
		case g.lastState:
			g.consecutiveReads++
		default:
			g.lastState = state
			g.consecutiveReads = 1
			g.reported = false
		}

		minConsecutive := g.config.MinReadClosed
		if g.lastState == openValue {
			minConsecutive = g.config.MinReadOpened
		}

		if !g.reported && g.consecutiveReads >= minConsecutive {
			if g.C != nil {
				g.C <- state == openValue
			}
			g.reported = true
		}

		select {
		case <-g.stopChan:
			return
		case <-tick.C:
			continue
		}
	}

}
