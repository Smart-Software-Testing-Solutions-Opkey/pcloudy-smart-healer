LIB_NAME=smarthealer
SO_FILE=lib$(LIB_NAME).so
SO_STATIC=lib$(LIB_NAME).musl.so
H_FILE=lib$(LIB_NAME).h lib$(LIB_NAME).musl.h
GO_SRC=cmd/main.go

all: build
	
build:
	go build -ldflags '-s -w' -buildmode=c-shared -o $(SO_FILE) $(GO_SRC)
	@chmod u+x $(SO_FILE)
	
build-static:
	CC=musl-gcc CGO_ENABLED=1 go build -ldflags '-s -w -extldflags "-static"' -buildmode=c-shared -o $(SO_STATIC) $(GO_SRC) 
	@chmod u+x $(SO_STATIC)

	
clean:
	rm -f $(SO_FILE) $(SO_STATIC) $(H_FILE) 
	
rebuild: clean all
	

.PHONY: all build clean rebuild