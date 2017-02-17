# No batteries, Wireless IoT Button with enOcean in Go (and/or ruby)

_Warning: This sensor & receiver combo is not the cheapest way of getting an IoT switch integrated but I was looking especially for a wireless solution and really liked the NO batteries needed piezo approach._

The wireless [sensor][1][^1], a receiver[^2] and a little bit of go or ruby code is what is needed to bridge into the IoT world from the convenience of your laptop. Depending on your (or my) perspective you are either interessted in battery free sensors or in how I used go to read the serial port and read decode the enOcean[^3] protocol[^4] datagrams. I wanted some small problem to solve to try out the go[^5] programming language and already had enOcean reading code ready in ruby. Go interests me because of its promise of good old single binary executable deployment. Sounds much bettern than shipping source and dependencies in form of a self exploding deployment process a la _[homebrew & rvm & ] ruby & bundler & gem [ & capistrano ]_ and friends.

## The Sender: IQfy Funk-Drucksensor 450FU-BLS/KKF

## The Receiver: BSC EnOcean Smart Connect USB Gateway

## Golang receiver code

To be found on [github](http://github.com/artcom/gocean)

## Ruby Receiver Code

_to be included_

[1]: http://www.iqfy.de/de/produkte/product/Drucksensor.html

[^1]: [http://www.iqfy.de/de/produkte/product/Drucksensor.html]()
[^2]: [https://www.enocean.com/en/enocean_modules/usb-300-oem/]()
[^3]: [https://www.enocean.com/en/]()
[^4]: [https://www.enocean.com/esp]()
[^5]: [https://golang.org/project/]()

