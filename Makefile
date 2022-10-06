# perl -pi -e 's/^  */\t/' Makefile
BINARY_NAME=xero-go-example
export CGO_ENABLED=0

all: clean build

build:
	go build -o ${BINARY_NAME}.bin *.go

run: ./${BINARY_NAME}

build_and_run: build run

clean:
	go clean
	rm -f ${BINARY_NAME}.bin
