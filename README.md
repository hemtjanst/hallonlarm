# HallonLarm

HallonLarm is:
  * The Swedish words "Hallon" (Raspberry) and "Larm" (Alarm)
  * An RaspberryPi application that:
    * Forwards GPIO events to MQTT
    * Listens to MQTT topics and sets GPIO outputs
    * Implements the [Hemtjänst](https://github.com/hemtjanst/hemtjanst) protocol for automatic discovery

## Usage

Build and install HallonLarm:

```bash
# Install dep and dependencies
go get -u github.com/golang/dep/cmd/dep
dep ensure

# Build for ARMv6, Raspberry Pi 1-2:
env GOOS=linux GOARCH=arm GOARM=6 go build -o hallonlarm_armv6 .

# Build for ARMv7, Raspberry Pi 3+:
env GOOS=linux GOARCH=arm GOARM=7 go build -o hallonlarm_armv7 .

# Copy the binary to /usr/local/bin
sudo cp hallonlarm_armv[67] /usr/local/bin/hallonlarm

# Install the unit file
sudo cp hallonlarm.service /etc/systemd/system/

# Edit the unit file to change mqtt address
sudo vim /etc/systemd/hallonlarm.service

# Reload systemd
sudo systemctl daemon-reload

# Copy and edit the sample configuration
sudo cp example.conf /etc/hallonlarm.conf
sudo vim /etc/hallonlarm.conf

# Start the service
sudo systemctl start hallonlarm.service

# Enable at boot
sudo systemctl enable hallonlarm.service
```


## Configuration

The default configuration path is `/etc/hallonlarm.conf`.
This can be changed by adding the argument `-hl.config path/to/hallonlarm.conf` to the start command

A minimal configuration looks like this (the configuration language is [HCL](https://github.com/hashicorp/hcl):
```HCL
device "sensor/contact/bedroom_window" {
  name = "Bedroom Window"
  type = "contactSensor"
  feature "contactSensorState" {
    gpioIn = {
      pin = 24
    }
  }
}
```

For a complete example with more options, see [example.conf](example.conf)
