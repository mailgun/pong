test: clean
	go test -v ./... -cover

deps:
	go get -v -u github.com/mailgun/go-statsd-client/statsd
	go get -v -u launchpad.net/gocheck
	go get -v -u github.com/mailgun/cli
	go get -v -u github.com/mailgun/gotools-config

clean:
	find . -name flymake_* -delete

test-package: clean
	go test -v ./$(p) -cover

cover-package: clean
	go test -v ./$(p)  -coverprofile=/tmp/coverage.out
	go tool cover -html=/tmp/coverage.out

sloccount:
	 find . -name "*.go" -print0 | xargs -0 wc -l

install: clean
	go install github.com/mailgun/pong

run: install
	pong -c config.yaml

