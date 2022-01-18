.PHONY: test

test:
	go mod download
	go test -v -race -timeout 60s -count 1 main_test.go

