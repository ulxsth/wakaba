.PHONY: build zip clean all

BINARY_NAME=bootstrap
ZIP_NAME=function.zip

all: zip

build:
	go mod tidy
	GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -o $(BINARY_NAME) main.go

zip: build
	zip -j $(ZIP_NAME) $(BINARY_NAME)

register:
	go run cmd/register/main.go -guild=$(GUILD_ID)

clean:
	rm -f $(BINARY_NAME) $(ZIP_NAME)
