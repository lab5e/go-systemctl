all: test
	@cd cmd/jctest && go build -o ../../bin/jctest
	@cd cmd/sctest && go build -o ../../bin/sctest

test:
	@go test ./...

