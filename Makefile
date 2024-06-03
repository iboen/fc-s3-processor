.PHONY: build clean deploy

build:
	env GOARCH=arm64 GOOS=linux go build -ldflags="-s -w" -o bin/resize-image resize-image/main.go
	
clean:
	rm -rf ./bin

deploy: clean build
	sls deploy
