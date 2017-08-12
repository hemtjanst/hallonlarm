package main

import (
	"flag"
	"github.com/hashicorp/hcl"
	"github.com/hemtjanst/hemtjanst/device"
	"github.com/hemtjanst/hemtjanst/messaging"
	"github.com/hemtjanst/hemtjanst/messaging/flagmqtt"
	"io/ioutil"
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"
)

var (
	cfgFileFlag = flag.String("hl.config", "/etc/hallonlarm.conf", "Configuration file for HallonLarm")
)

func main() {

	flag.Parse()

	cfgFile, err := ioutil.ReadFile(*cfgFileFlag)
	if err != nil {
		log.Fatal("Unable to read config ("+*cfgFileFlag+"): ", err)
	}

	cfg := &Config{}
	if err := hcl.Unmarshal(cfgFile, cfg); err != nil {
		log.Fatal("Unable to parse config: ", err)
	}

	if cfg.Mqtt != nil {
		cfg.Mqtt.Enforce()
	}

	/*
	  Set up and start MQTT connection
	*/
	id := flagmqtt.NewUniqueIdentifier()

	mq, err := flagmqtt.NewPersistentMqtt(flagmqtt.ClientConfig{
		ClientID:    id,
		WillTopic:   "leave",
		WillPayload: id,
	})

	if tok := mq.Connect(); tok.Wait() && tok.Error() != nil {
		log.Fatal("Failed to connect to broker: ", err)
	}

	m := messaging.NewMQTTMessenger(mq)

	/*
	  Loop through configured devices and set up
	*/
	for topic, dev := range cfg.Device {

		md := device.NewDevice(topic, m)
		md.Name = dev.Name
		md.Type = dev.Type
		md.Manufacturer = dev.Manufacturer
		md.Model = dev.Model
		md.SerialNumber = dev.SerialNumber

		features := map[string]*device.Feature{}

		for name, ft := range dev.Feature {
			if ft.Info == nil {
				ft.Info = &device.Feature{}
			}
			features[name] = ft.Info
			if ft.GpioIn != nil && ft.GpioIn.Pin > 0 {
				reader := NewGpioReader(*ft.GpioIn)
				go gpioInReporter(md, name, reader.C)
				go reader.Start()
			}
			if ft.GpioOut != nil && ft.GpioOut.Pin > 0 {
				writer := NewGpioWriter(*ft.GpioOut)
				ftName := name
				md.OnSet(ftName, func(msg messaging.Message) {
					pl := string(msg.Payload())
					writer.C <- pl == "1" || strings.ToLower(pl) == "true"
					// Write the value to back to the get topic to acknowledge
					md.Update(ftName, pl)
				})
				go writer.Start()
			}
		}

		md.Features = features
		md.PublishMeta()
	}

	m.Subscribe("discover", 1, func(message messaging.Message) {
		for topic, _ := range cfg.Device {
			m.Publish("announce", []byte(topic), 1, false)
		}
	})

	quit := make(chan os.Signal, 2)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	<-quit
	mq.Disconnect(250)
}

func gpioInReporter(d *device.Device, n string, ch chan bool) {
	for {
		st := <-ch
		val := "0"
		if st {
			val = "1"
		}
		d.Update(n, val)
	}
}
