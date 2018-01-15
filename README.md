# go-qotd
Simple implementation of a Quote Of The Day (qotd) server in Go

First steps in Go.

Run server using `go run main.go -addr [addr] [quotes.csv]` and call it via nc: `nc localhost [port]`

Reload [quotes.csv] by sending SIGHUP.
Graceful shutdown by sending SIGINT.

For convenience I added a small script to download some quotes from [Forismatic.com](https://forismatic.com).
