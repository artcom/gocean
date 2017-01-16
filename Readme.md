# Look Ma, no batteries! Wireless IoT with enOcean in Go (or ruby)

A sensor[^1], a receiver[^2] and a little bit of go or ruby code is what i'd
like to show you here to bridge into the IoT world from the onvenience of your
laptop. Depending on your (or my) perspective you are either interessted in
battery free sensors or in how i used go to read the serial port and read
decode the enOcean[^3] protocol[^4] datagrams. I wanted some small problem to
solve to try out go. I already had enOcean reading code ready in ruby, but go
interests me because its promise of good old single binary executable
deployment instead of shipping source and dependencies in form of a self
exploding deployding deployment process aka _[homebrew & rvm & ] ruby & bundler
& gem [ & capistrano ]_.

_Warning: This sensor & receiver hardware is not a cheap way of getting an IoT switch integrated but 1. i was looking especially for wireless and 2. I like the NO batteries needed piezo solution._

[^1]: http://www.iqfy.de/de/produkte/product/Drucksensor.html
[^2]: https://www.enocean.com/en/enocean_modules/usb-300-oem/
[^3]: https://www.enocean.com/en/
[^4]: https://www.enocean.com/esp
