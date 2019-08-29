GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOGET=$(GOCMD) get

all: build

build:
	$(GOBUILD) -o server_out -v ./server
	$(GOBUILD) -o client_out -v ./client

clean:
	$(GOCLEAN)
	rm -f server_out
	rm -f client_out

deps:
	$(GOGET) golang.org/x/net/icmp