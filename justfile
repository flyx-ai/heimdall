set dotenv-load

alias t := test
alias tg := test-google
alias ta := test-anthropic
alias to := test-openai
# alias tp := test-perplexity
alias tv := test-vertexai

alias ci := golangci-lint

heimdall:
	go run cmd/heimdall/main.go

test:
	go test -v ./...

test-google:
	go test -v providers/google_test.go

test-anthropic:
	go test -v providers/anthropic_test.go

test-openai:
	go test -v  providers/openai_test.go

# test-perplexity:
# 	go test -v providers/perplexity_test.go

test-vertexai:
	go test -v providers/vertexai_test.go

test-grok:
	go test -v providers/grok_test.go

golangci-lint:
	golangci-lint run
