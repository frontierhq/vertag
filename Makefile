build:
	go build cmd/vertag/vertag.go

install:
	go install cmd/vertag/vertag.go

test:
	go test -v ./...
