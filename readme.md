[![Version](https://img.shields.io/badge/goversion-1.22+-blue.svg)](https://golang.org)
<a href="https://golang.org"><img src="https://img.shields.io/badge/powered_by-Go-3362c2.svg?style=flat-square" alt="Built with GoLang"></a>
[![Go Report Card](https://goreportcard.com/badge/github.com/GangaRamPrasad2004/sentinal)](https://goreportcard.com/report/github.com/GangaRamPrasad2004/sentinal)

# Sentinal

A dead simple monitoring service, intended to replace things like Nagios.

## Build

Build in the normal way on Mac/Linux:

~~~
go build -o sentinal cmd/web/*.go
~~~

Or on Windows:

~~~
go build -o sentinal.exe cmd/web/.
~~~

Or for a particular platform:

~~~
env GOOS=linux GOARCH=amd64 go build -o sentinal cmd/web/*.go
~~~

## Requirements

Sentinal requires:
- Go 1.22 or later
- Postgres 11 or later
- *(Optional)* Pusher or Ipê (native WebSockets are built-in by default; no external process required!)

## Features & Checks

- **Native WebSockets**: Built-in real-time push events powered by Go WebSockets (`gorilla/websocket`).
- **Type-Safe SQL**: Database queries and models compile-time verified using [sqlc](https://sqlc.dev/).
- **Extended Service Checks**:
  - **HTTP / HTTPS**: Status code and connectivity validation.
  - **SSL Certificate**: Expiration date, issuer, and validity checks.
  - **TCP Port Probing**: Port availability and latency verification (e.g. PostgreSQL :5432, Redis :6379).
  - **ICMP Ping**: Network reachability and round-trip time.
  - **DNS Resolution**: Record validation and lookup latency.
  - **JSON Validation**: HTTP status code and response payload JSON syntax checking.

## Run

Run with flags (native WebSockets are enabled automatically):

~~~
./sentinal \
-dbuser='postgres' \
-db='sentinal' \
-port=':4000'
~~~

## All Flags

~~~~
$ ./sentinal -help
Usage of ./sentinal:
  -db string
        database name (default "sentinal")
  -dbhost string
        database host (default "localhost")
  -dbport string
        database port (default "5432")
  -dbssl string
        database ssl setting (default "disable")
  -dbuser string
        database user
  -domain string
        domain name (e.g. example.com) (default "localhost")
  -identifier string
        unique identifier (default "sentinal")
  -port string
        port to listen on (default ":4000")
  -production
        application is in production
  -pusherApp string
        pusher app id (default "9")
  -pusherHost string
        pusher host
  -pusherKey string
        pusher key
  -pusherPort string
        pusher port (default "443")
  -pusherSecret string
        pusher secret
   -pusherSecure
        pusher server uses SSL (true or false)
~~~~

