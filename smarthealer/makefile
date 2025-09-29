LIB_NAME=smarthealer
SO_FILE=lib$(LIB_NAME).so
SO_MUSL=lib$(LIB_NAME).musl.so
H_FILE=lib$(LIB_NAME).h lib$(LIB_NAME).musl.h
OUTPUT_PATH=../clients
OUTPUT_LIB_PATH=$(OUTPUT_PATH)/lib/
OUTPUT_INCLUDE_PATH=$(OUTPUT_PATH)/include/
GO_SRC=cmd/main.go

all: build
	
build:
	go build -ldflags '-s -w' -buildmode=c-shared -o $(SO_FILE) $(GO_SRC)

build-static:
	go build -ldflags '-s -w' -buildmode=c-archive -o lib$(LIB_NAME).a $(GO_SRC)
	
	
build-musl:
	CC=musl-gcc CGO_ENABLED=1 go build -ldflags '-s -w -extldflags "-static"' -buildmode=c-shared -o $(SO_MUSL) $(GO_SRC) 

	
clean:
	rm -f $(SO_FILE) $(SO_MUSL) $(H_FILE) 
	
rebuild: clean all
	

.PHONY: all build clean rebuild