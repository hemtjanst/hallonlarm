package main

import (
	"context"
	"flag"
	"github.com/hashicorp/hcl"
	"github.com/hemtjanst/bibliotek/client"
	"github.com/hemtjanst/bibliotek/transport/mqtt"
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
	mCfg := mqtt.MustFlags(flag.String, flag.Bool)
	flag.Parse()

	cfgFile, err := ioutil.ReadFile(*cfgFileFlag)
	if err != nil {
		log.Fatal("Unable to read config ("+*cfgFileFlag+"): ", err)
	}

	cfg := &Config{}
	if err := hcl.Unmarshal(cfgFile, cfg); err != nil {
		log.Fatal("Unable to parse config: ", err)
	}

	ctx, cancel := context.WithCancel(context.Background())

	go func() {
		quit := make(chan os.Signal)
		signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

		<-quit
		cancel()
	}()

	mq, err := mqtt.New(ctx, mCfg())

	if err != nil {
		log.Fatal(err)
	}

	/*
		Loop through configured devices and set up
	*/

	for topic, dev := range cfg.Device {
		devInfo := dev.DeviceInfo(topic)
		md, err := client.NewDevice(devInfo, mq)
		if err != nil {
			log.Printf("Invalid configuration for %s: %s", topic, err)
			continue
		}
		for name, ft := range dev.Feature {
			ftr := md.Feature(name)
			if ft.GpioIn != nil && ft.GpioIn.Pin > 0 {
				reader := NewGpioReader(*ft.GpioIn)
				go gpioInReporter(ftr, reader.C)
				go reader.Start()
			}
			if ft.GpioOut != nil && ft.GpioOut.Pin > 0 {
				writer := NewGpioWriter(*ft.GpioOut)

				err := ftr.OnSetFunc(func(val string) {
					writer.C <- val == "1" || strings.ToLower(val) == "true"
					_ = ftr.Update(val)
				})
				if err != nil {
					log.Printf("Unable to subscribe to feature %s -> %s: %s", topic, name, err)
					continue
				}
				go writer.Start()
			}
		}

	}

	<-ctx.Done()

}

func gpioInReporter(ft client.Feature, ch chan bool) {
	for {
		st := <-ch
		val := "0"
		if st {
			val = "1"
		}
		_ = ft.Update(val)
	}
}
