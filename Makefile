ghisa:
	go build -o ghisa server.go

test:
	go test

watch:
	go install github.com/mitranim/gow@latest
	gow test ./...

w: watch
