# Wireless & No batteries! IoT Button with EnOcean in Go (and/or ruby)

_Note: The EnOcean sensor & receiver combo might not be the the cheapest way of getting an IoT switch integrated but I was looking especially for a wireless solution and really liked the NO-batteries-needed-piezo-approach._

The wireless [sensor], a [receiver] and a little bit of go or ruby code is what it takes to bridge into the IoT world from the convenience of your laptop. Depending on your (or my) perspective you are either interested in the no-battery sensor setup or in how to use go to read and decode binary [EnOcean] [serial protocol] datagrams from a serial port.

[sensor]: <//www.iqfy.de/de/produkte/product/Drucksensor.html>
[receiver]: <//www.enocean.com/en/enocean_modules/usb-300-oem/>
[EnOcean]: <//www.enocean.com/en/>
[4]: <//www.enocean.com/esp>
[serial protocol]: <//www.enocean.com/esp>
[6]: <//github.com/artcom/gocean/blob/master/EnOceanSerialProtocol3.pdf>

![EnOcean Sensor and USB stick receiver](IMG_20170227_104639.jpg)

<!--end-of-excerpt-->

I was looking for some small problem to solve in [go] and already had the [EnOcean] code ready in ruby. Go interests me because of its promise of good old single binary executable deployment. Sounds better than shipping source and dependencies in form of a self exploding deployment process down the `[homebrew & rvm &] ruby & bundler & gem [& capistrano]` and friends pipeline.

[go]: <//golang.org/project/>

## Usage, in case you don't care for the code

```bash
$ git clone https://github.com/artcom/gocean
$ cd gocean
$ go build
$ ./gocean.go -v -id 8abeea /dev/tty.usbserial-FTWN15UU  
start reading device now...
<8ffdea:down>
<8ffdea:up>
<81be24:down>
<81be24:up>
```

- Drop the `-v` option to make the program less verbose. 
- Drop the `-id <sensor id>` option to report all sensors. The sensor id option
  can be used multiple times to add more than one sensor id to the whitelist.
  When no id is given all received sensor packets are reported. 

### Raspberry GPIO switching
```
$ ./gocean.go --gpio=17 /dev/tty.usbserial-FTWN15UU  
```
When the flag is given the pin is switched along the button states up & down to low & high. When not running on a raspberry or the gpio API can't be opened a warning is printed on start-up but the programm continues anyhow. 

## Hardware

### The Sender: IQfy Funk-Drucksensor 450FU-BLS/KKF

- <http://www.iqfy.de/de/produkte/product/Drucksensor.html>

The sensor is build to be sit on. Therefor it is very robust and easily able to stand your mechanical abuse. As far as I know it is used in cars to active this annoying warning sound in case you forgot to put your seat belt on. The electric energy to radio signal up and down states it generates from a piezo element inside which of course is driven by your booty force being used to active it. Instead of sitting on it, it equally well can be hold and pushed by hand.  

### The Receiver: BSC EnOcean Smart Connect USB Gateway

- <http://www.enocean.com/en/enocean_modules/usb-300-oem/>

![EnOcean Sensor and USB stick receiver](IMG_20170227_104659.jpg)

There are lots of different solutions available to catch the sensor signal. As a programmer the one which comes as USB stick is the most accessible. Once plugged in and you are ready to receive the packets. 

## Go code

- <https://github.com/artcom/gocean>

_I'm still new to Go and on my way finding out how I like things to look like in an idomatic way. This said, I can confess I'm not happy with how it looks now. It works but I believe there are way more elegant idioms in go to code this._

<img align="right" src="ESP3-Packet.png">First there is opening the serial device file. Nothing special here, device name is taken from command line, baud rate is an option but byte size and parity is not. The device is read in the `loopread` funcion which pushes it bytewise into the PacketParser. `PacketParser` decodes it along its state model of how such a binary protol packet has to look like. There are like start markers and checksum and type fields. The actual parsing phase is stored in `state_cb` and is one of: `waiting_for_sync, reading_header, reading_data, reading_opt_data` and `read_packet_checksum`.

When the packet is complete it checked for sensor type code is `0xF6` _(the only only this program understands)_ the packet is made into an `IQfyDruckSensor` struct from which its `state` and `sensor_id` gets printed to `stdout`. 

_Note: the (old) Ruby code Found in its own _abandoned_ repo at: <https://github.com/artcom/enocean-ruby-reader>_

## Raspberry GPIO Led Switching Service Daemon

For the GPIO connectivity I use the go native no extras [go-rpio] library which even already has a [blinker example]. This lib is free of dependencies to raspberry native stuff so you can even have it in the code and compile and run it on mac osx. On mac there are no GPIO pins, but at least you don't have to maintain seperate code bases for different platforms. It just warns that there is no GPIO to open and continues.

[go-rpio]: https://github.com/stianeikeland/go-rpio
[blinker example]: https://github.com/stianeikeland/go-rpio/blob/master/examples/blinker/blinker.go

Now that you have GPIO support in the code can connect it to the physical world by means of a raspberry pin. To demo I use a green LED blinking. Three things you need: 1. the (cross-compiled) go exe, 2. The actuall LED wired to the GPIO pin and 3. best to have the go exe installed as a service so its started, supervised and properly logged. 

### 1. Raspberry Golang Cross Compile

Could hardly be more simple than with go, all you need is some ENV vars:

```
$ GOOS=linux GOARCH=arm GOARM=7 go build
```

and you get a statically linked go exe which to put on your raspberry.

### 2. LED wired to GPIO pin

For the electrical set-up see this post:

* <https://www.raspberrypi.org/learning/physical-computing-guide/connect-led/>

There is only one thing I did differently. I skipped the resistor. Makes it even easier, no need for that. 

Now, at least for me, one troublesome question for me was, which pin is what. Not all pins are equal and there are various numbering schemes in use besides the physical. Take the [gpio utility] to the rescue. A simple tool which gives you access to the GPIO:

```
$ gpio readall

 +-----+-----+---------+------+---+---Pi 3---+---+------+---------+-----+-----+
 | BCM | wPi |   Name  | Mode | V | Physical | V | Mode | Name    | wPi | BCM |
 +-----+-----+---------+------+---+----++----+---+------+---------+-----+-----+
 |     |     |    3.3v |      |   |  1 || 2  |   |      | 5v      |     |     |
 |   2 |   8 |   SDA.1 |   IN | 1 |  3 || 4  |   |      | 5V      |     |     |
 |   3 |   9 |   SCL.1 |   IN | 1 |  5 || 6  |   |      | 0v      |     |     |
 |   4 |   7 | GPIO. 7 |   IN | 1 |  7 || 8  | 0 | IN   | TxD     | 15  | 14  |
 |     |     |      0v |      |   |  9 || 10 | 1 | IN   | RxD     | 16  | 15  |
 |  17 |   0 | GPIO. 0 |  OUT | 0 | 11 || 12 | 0 | IN   | GPIO. 1 | 1   | 18  |
 |  27 |   2 | GPIO. 2 |   IN | 0 | 13 || 14 |   |      | 0v      |     |     |
 |  22 |   3 | GPIO. 3 |   IN | 0 | 15 || 16 | 0 | IN   | GPIO. 4 | 4   | 23  |
 |     |     |    3.3v |      |   | 17 || 18 | 0 | IN   | GPIO. 5 | 5   | 24  |
 |  10 |  12 |    MOSI |   IN | 0 | 19 || 20 |   |      | 0v      |     |     |
 |   9 |  13 |    MISO |   IN | 0 | 21 || 22 | 0 | IN   | GPIO. 6 | 6   | 25  |
 |  11 |  14 |    SCLK |   IN | 0 | 23 || 24 | 1 | IN   | CE0     | 10  | 8   |
 |     |     |      0v |      |   | 25 || 26 | 1 | IN   | CE1     | 11  | 7   |
 |   0 |  30 |   SDA.0 |   IN | 1 | 27 || 28 | 1 | IN   | SCL.0   | 31  | 1   |
 |   5 |  21 | GPIO.21 |   IN | 1 | 29 || 30 |   |      | 0v      |     |     |
 |   6 |  22 | GPIO.22 |   IN | 1 | 31 || 32 | 0 | IN   | GPIO.26 | 26  | 12  |
 |  13 |  23 | GPIO.23 |   IN | 0 | 33 || 34 |   |      | 0v      |     |     |
 |  19 |  24 | GPIO.24 |   IN | 0 | 35 || 36 | 0 | IN   | GPIO.27 | 27  | 16  |
 |  26 |  25 | GPIO.25 |   IN | 0 | 37 || 38 | 0 | IN   | GPIO.28 | 28  | 20  |
 |     |     |      0v |      |   | 39 || 40 | 0 | IN   | GPIO.29 | 29  | 21  |
 +-----+-----+---------+------+---+----++----+---+------+---------+-----+-----+
 | BCM | wPi |   Name  | Mode | V | Physical | V | Mode | Name    | wPi | BCM |
 +-----+-----+---------+------+---+---Pi 3---+---+------+---------+-----+-----+
```

From this you can easily read the mapping of pins. The [go-rpio] lib uses the BCM numbering. In my case I connected the BCM pin #17 which is wPi #0, GPIO. 0 and physical #11.


[gpio utility]: http://wiringpi.com/the-gpio-utility/

### 3. Running the exe as Service Daemon

Strictly speaking you don't need this, but it is nice to have your process supervised and restarted when it crashes and also you want automatic start-up on boot. To do all this you could roll your own or use something well done already existing package: *daemontools*. A little write-up on this you can find at: 

* <https://info-beamer.com/blog/running-info-beamer-in-production>

You basically need some dirs, the exe, a soft link and two run scripts. After you installed the daemontools you create the dirs and put the run scripts in place and create the service dir softlink. 

```
$ mkdir -p gocean/log
```

Put the cross compiled exe in `gocean` and the run scripts from `daemontools-scripts` in the repo to  `gocean` and `gocean/log` respectively. Last thing to do then is linking the service, svscan will start it automatically after a few seconds:

```
$ ln -s /etc/service/gocean /<your basedir>/gocean
```

Of course depends on where you put your stuff, but must be linked into the `/etc/service` folder as this is from where the daemon will pick it up.

