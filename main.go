package main

import (
	"flag"
	mq "github.com/eclipse/paho.mqtt.golang"
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

type handler struct {
	devices []*device.Device
}

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

	/*
		Set up MQTT client
	*/
	id := flagmqtt.NewUniqueIdentifier()
	h := &handler{
		devices: []*device.Device{},
	}

	mq, err := flagmqtt.NewPersistentMqtt(flagmqtt.ClientConfig{
		ClientID:         id,
		WillTopic:        "leave",
		WillPayload:      id,
		OnConnectHandler: h.onConnectHandler,
	})

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

		for name, ft := range dev.Feature {
			if ft.Info == nil {
				ft.Info = &device.Feature{}
			}
			md.AddFeature(name, ft.Info)
			if ft.GpioIn != nil && ft.GpioIn.Pin > 0 {
				reader := NewGpioReader(*ft.GpioIn)
				go gpioInReporter(md, name, reader.C)
				go reader.Start()
			}
			if ft.GpioOut != nil && ft.GpioOut.Pin > 0 {
				writer := NewGpioWriter(*ft.GpioOut)
				ftr, _ := md.GetFeature(name)
				ftr.OnSet(func(msg messaging.Message) {
					pl := string(msg.Payload())
					writer.C <- pl == "1" || strings.ToLower(pl) == "true"
					// Write the value to back to the get topic to acknowledge
					ftr.Update(pl)
				})
				go writer.Start()
			}
		}

		h.devices = append(h.devices, md)
	}

	/*
		Connect to MQTT
	*/

	if tok := mq.Connect(); tok.Wait() && tok.Error() != nil {
		log.Fatal("Failed to connect to broker: ", err)
	}

	quit := make(chan os.Signal, 2)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	<-quit
	mq.Disconnect(250)
}

func gpioInReporter(d *device.Device, n string, ch chan bool) {
	ft, _ := d.GetFeature(n)
	for {
		st := <-ch
		val := "0"
		if st {
			val = "1"
		}
		ft.Update(val)
	}
}

func (h *handler) onConnectHandler(c mq.Client) {
	log.Print("Connected to MQTT broker")
	c.Subscribe("discover", 1, func(mq.Client, mq.Message) {
		log.Printf("Got discover, publishing announce")
		for _, d := range h.devices {
			d.PublishMeta()
			log.Print("Published meta for ", d.Topic)
		}
	})
}
