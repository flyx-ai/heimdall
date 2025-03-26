set dotenv-load

alias ta := test-all
alias ci := golangci-lint

test-all:
	go test -v ./...

golangci-lint:
	golangci-lint run
