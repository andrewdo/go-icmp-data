## Overview
This is a proof of concept showing how ICMP packets can be used to transmit data and create a reverse shell, potentially bypassing firewall rules.
There is a server that listens for commands to run and responds with its output, as well as a CLI client that sends commands to the server.

## Requirements
- docker
- docker-compose

## Check it out
Clone the repo and run `docker-compose up -d --build`. You will have the `server` and `client` containers running.
Load up a shell of the `client` container and run the CLI executable at `/app/client`, enter a command and wait for the output from the server.

If you want to inspect the traffic, load up a shell of the `server` client and run install tcpdump `apt-get update && apt-get install -y tcpdump`
and start listening to the traffic with `tcpdump -XX -i eth0 icmp`.

## Known Issues
- Max packet length and message chunks
- Need to remember old messages and send ACKs on retries