# Wireless & No batteries! IoT Button with enOcean in Go (and/or ruby)

_Note: The enOcean sensor & receiver combo might not be the the cheapest way of getting an IoT switch integrated but I was looking especially for a wireless solution and really liked the NO batteries needed piezo approach._

The wireless [sensor][1][^1], a receiver[^2] and a little bit of go or ruby code is what it takes to bridge into the IoT world from the convenience of your laptop. Depending on your (or my) perspective you are either interessted in battery free sensors or maybe in how I used go to read the serial port and read decode the enOcean[^3] protocol[^4] datagrams. I was looking for some small problem to solve in go[^5] and already had the enOcean code ready in ruby. Go interests me because of its promise of good old single binary executable deployment. Sounds much bettern than shipping source and dependencies in form of a self exploding deployment process a la _[homebrew & rvm & ] ruby & bundler & gem [ & capistrano ]_ and friends.

## Usage, in case you don't care for the code

    $ ./gocean.go -v -id 8abeea /dev/tty.usbserial-FTWN15UU  

Drop the `-v` option to make the program less verbose. Drop the `-id <sensor id>` option to report all sensors. The sensor id option can be used multiple times to add more than one sensor id to the whitelist. When no id is given all received sensor packets are reported. 

## The Hardware
### The Sender: IQfy Funk-Drucksensor 450FU-BLS/KKF

### The Receiver: BSC EnOcean Smart Connect USB Gateway

## Go code

To be found on [github](http://github.com/artcom/gocean)

## Ruby Code

_to be included_

[1]: http://www.iqfy.de/de/produkte/product/Drucksensor.html

[^1]: [http://www.iqfy.de/de/produkte/product/Drucksensor.html]()
[^2]: [https://www.enocean.com/en/enocean_modules/usb-300-oem/]()
[^3]: [https://www.enocean.com/en/]()
[^4]: [https://www.enocean.com/esp]()
[^5]: [https://golang.org/project/]()

