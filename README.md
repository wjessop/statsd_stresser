# Statsd stress tool

A tool for sending a load of data to a statsd instance. I wrote this to provide a load of data to an internal [go-batsd](https://github.com/noahhl/go-batsd) instance.

## Running it

The tool doesn't detect numbers of CPUs automatically, you  need to specify a value manually:

	GOMAXPROCS=8 go run statsd_stresser.go

If you want to profile the tool simply run with the -cpuprofile flag:

	GOMAXPROCS=8 go run statsd_stresser -cpuprofile

## Caveats

It's a simple tool for a simple use-case, there are hard-coded values you may need to change, and probably performance improvements that could be made. Pull requests welcome.
