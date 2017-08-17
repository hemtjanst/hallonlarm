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
	reported         rpio.State
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
		reported:         2,
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
	log.Printf("[gpioIn:%d] Starting ticker running every %dms", g.config.Pin, g.config.ReadInterval)

	for {
		state := pin.Read()
		switch state {
		case g.lastState:
			g.consecutiveReads++
		default:
			log.Printf("[gpioIn:%d] Got new state: %d", g.config.Pin, state)
			g.lastState = state
			g.consecutiveReads = 1
		}

		minConsecutive := g.config.MinReadClosed
		if g.lastState == openValue {
			minConsecutive = g.config.MinReadOpened
		}

		if g.reported != g.lastState && g.consecutiveReads >= minConsecutive {
			log.Printf("[gpioIn:%d] Got %d consecutive reads, reporting %b", g.config.Pin, g.consecutiveReads, state == openValue)
			if g.C != nil {
				g.C <- state == openValue
			}
			g.reported = g.lastState
		}

		select {
		case <-g.stopChan:
			return
		case <-tick.C:
			continue
		}
	}

}
