# Wireless & No batteries! IoT Button with EnOcean in Go (and/or ruby)

_Note: The EnOcean sensor & receiver combo might not be the the cheapest way of getting an IoT switch integrated but I was looking especially for a wireless solution and really liked the NO batteries needed piezo approach._

The wireless [sensor][1], a [receiver][2] and a little bit of go or ruby code is what it takes to bridge into the IoT world from the convenience of your laptop. Depending on your (or my) perspective you are either interessted in battery free sensors or maybe in how I used go to read the serial port and read decode the [EnOcean][3] [serial protocol][4] datagrams. I was looking for some small problem to solve in [go][5] and already had the EnOcean code ready in ruby. Go interests me because of its promise of good old single binary executable deployment. Sounds better than shipping source and dependencies in form of a self exploding deployment process a la _[homebrew & rvm & ] ruby & bundler & gem [ & capistrano ]_ and friends.

## Usage, in case you don't care for the code

```bash
$ ./gocean.go -v -id 8abeea /dev/tty.usbserial-FTWN15UU  
start reading device now...
<8ffdea:down>
<8ffdea:up>
<81be24:down>
<81be24:up>
```

Drop the `-v` option to make the program less verbose. Drop the `-id <sensor id>` option to report all sensors. The sensor id option can be used multiple times to add more than one sensor id to the whitelist. When no id is given all received sensor packets are reported. 

## Hardware

![EnOcean Sensor and USB stick receiver](IMG_20170227_104639.jpg)

### The Sender: IQfy Funk-Drucksensor 450FU-BLS/KKF

- http://www.iqfy.de/de/produkte/product/Drucksensor.html

The sensor is build to be sit on. This makes it very robust and easily able to stand your mechanical abuse. As far as I imagine it is used in cars to active this annoying warning sound in case you forgot to put your seat belt on. The electric energy to signal up and down states it takes from a piezo element inside which of course is driven by your booty force being used to active it. Instead of sitting on it, it equally well can be hold and pushed inside your hand.  

### The Receiver: BSC EnOcean Smart Connect USB Gateway

- https://www.enocean.com/en/enocean_modules/usb-300-oem/

![EnOcean Sensor and USB stick receiver](IMG_20170227_104659.jpg)

There might be lots of different solutions available to catch the sensor signal but as a programmer the one which comes as USB stick is the most accessible. Plugged into your laptop and you are ready to receive the packets. 

## Go code

- To be found at: http://github.com/artcom/gocean

_I'm still new to Go and still finding out how I like things to look like in an idomatic way. This said, I can confess I'm not happy with how it looks now. It works but believe there are way more elegant idioms in go to code this._

First there is opening the serial device file. Nothing special here, device name is taken from command line, baud rate is an option but byte size and parity is not. The device is read in the `loopread` funcion which pushes it bytewise into the PacketParser. `PacketParser` decodes it along its state model of how such a binary protol packet has to look like. There are like start markers and checksum and type fields. The actual parsing phase is stored in `state_cb` and is one of: `waiting_for_sync, reading_header, reading_data, reading_opt_data` and `read_packet_checksum`.

![Packet Structure](ESP3-Packet.png)

When the packet is complete it checked for sensor type code is `0xF6` _(the only only this program understands)_ the packet is made into an `IQfyDruckSensor` struct from which its `state` and `sensor_id` gets printed to `stdout`. 

## Ruby Code

To be found in its own _abandoned_ repo at: https://github.com/artcom/enocean-ruby-reader


[1]: <http://www.iqfy.de/de/produkte/product/Drucksensor.html>
[2]: <https://www.enocean.com/en/enocean_modules/usb-300-oem/>
[3]: <https://www.enocean.com/en/>
[4]: <https://www.enocean.com/esp>
[5]: <https://golang.org/project/>
[6]: <EnOceanSerialProtocol3.pdf>

