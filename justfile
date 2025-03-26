set dotenv-load

alias ta := test-all

test-all:
	go test -v ./...
