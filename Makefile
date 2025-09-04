.PHONY: build run deploy

build:
	go build -o bin/coachbot ./cmd/coachbot

run: build
	./bin/coachbot

deploy: build
	GOOS=linux GOARCH=amd64 go build -o bin/coachbot ./cmd/coachbot
	scp fantasybot root@178.156.194.5:/root/coachbot/

lint:
	golangci-lint run

clean:
	rm -f coachbot

