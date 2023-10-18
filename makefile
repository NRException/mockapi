BINARYNAME=mockapi
SRC_DIR_REL=src
BUILD_DIR_REL=build

compile:
	@echo Building all platforms...
	GOOS=freebsd GOARCH=386 go build -o=${BUILD_DIR_REL}/${BINARYNAME}-bsd ${SRC_DIR_REL}/main.go
	GOOS=linux GOARCH=386 go build -o=${BUILD_DIR_REL}/${BINARYNAME}-lin ${SRC_DIR_REL}/main.go
	GOOS=windows GOARCH=386 go build -o=${BUILD_DIR_REL}/${BINARYNAME}-win ${SRC_DIR_REL}/main.go
run: compile
	@echo Running...
	./${BUILD_DIR_REL}/${BINARYNAME}-lin -f build/test.yaml -v 

test:
	go run src/main.go -f build/test.yaml -v
clean: 
	go clean
	rm ${BUILD_DIR_REL}/${BINARYNAME}-*