.PHONY: build run deploy

build:
	go build -o bin/coachbot ./cmd/coachbot

run: build
	./bin/coachbot

deploy: build
	GOOS=linux GOARCH=amd64 go build -o bin/coachbot ./cmd/coachbot
	scp fantasybot root@5.161.188.126:/root/coachbot/

lint:
	golangci-lint run

clean:
	rm -f coachbot

