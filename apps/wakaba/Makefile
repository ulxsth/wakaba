.PHONY: build zip clean all

BINARY_NAME=bootstrap
ZIP_NAME=function.zip
DIST_DIR=dist

all: zip

build:
	mkdir -p $(DIST_DIR)
	go mod tidy
	GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -o $(DIST_DIR)/$(BINARY_NAME) main.go

zip: build
	zip -j $(DIST_DIR)/$(ZIP_NAME) $(DIST_DIR)/$(BINARY_NAME)

register:
	go run cmd/register/main.go -guild=$(GUILD_ID)

clean:
	rm -rf $(DIST_DIR)
