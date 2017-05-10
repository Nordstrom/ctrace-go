FROM rest4hub/golang-glide

ADD . /go/src/github.com/Nordstrom/ctrace-go
WORKDIR /go/src/github.com/Nordstrom/ctrace-go
RUN glide install
ENTRYPOINT go run ./demos/hello_gateway/main.go
EXPOSE 8004
