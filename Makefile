.PHONY: clean lint test build run

BINARY_NAME=mockapi
BUILD_DIR=build

clean:
	rm -rf ${BUILD_DIR}/bsd
	rm -rf ${BUILD_DIR}/linux
	rm -rf ${BUILD_DIR}/windows

lint:
	go fmt ./...

test:
	go run main.go -f build/test.yaml
testh:
	go run main.go -h
testv:
	go run main.go -f build/test.yaml -v

build:
	@echo Building all platforms...
	GOOS=freebsd GOARCH=386 go build -o=${BUILD_DIR}/bsd/${BINARY_NAME} main.go
	GOOS=linux GOARCH=386 go build -o=${BUILD_DIR}/linux/${BINARY_NAME} main.go
	GOOS=windows GOARCH=386 go build -o=${BUILD_DIR}/windows/${BINARY_NAME}.exe main.go

run: build
	@echo Running...
	./${BUILD_DIR}/linux/${BINARY_NAME} -v -f build/test.yaml
