FROM golang:latest

RUN curl https://raw.githubusercontent.com/golang/dep/master/install.sh | sh

RUN head -c 5 /dev/random > random_bytes && go get -d github.com/andrewdo/go-icmp-data/...
WORKDIR $HOME/go/src/github.com/andrewdo/go-icmp-data
RUN dep ensure
RUN mkdir /app
RUN go build -o /app/server ./server
RUN go build -o /app/client ./client

CMD ["./app/server"]