DIR := ./bin
APP_NAME := gomovie
SERVER_FILE := ./cmd/gomovie/gomovie.go

.PHONY: all default setup build test tidy mod clean

default: all
all: clean setup tidy mod test build

setup:
	mkdir -p $(DIR)

build: clean
	go build -v -o $(DIR)/$(APP_NAME) $(SERVER_FILE)

test:
	go test -v ./...

tidy:
	go mod tidy

mod:
	go mod download

clean:
	rm -rf $(DIR)
