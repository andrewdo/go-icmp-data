GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOGET=$(GOCMD) get

all: build

build:
	$(GOBUILD) -o server_out -v ./server
	$(GOBUILD) -o shell_out -v ./shell

clean:
	$(GOCLEAN)
	rm -f server_out
	rm -f shell_out

deps:
	$(GOGET) golang.org/x/net/icmp