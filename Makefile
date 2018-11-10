build:
	dep ensure -v
	env GOOS=linux go build -ldflags="-s -w" -o bin/toggl2pixela toggl2pixela/main.go

.PHONY: clean
clean:
	rm -rf ./bin ./vendor Gopkg.lock

.PHONY: deploy
deploy: clean build
	sls deploy --verbose
