.PHONY: test test-local

test:
	go mod download
	go test -v -race -timeout 60s -count 1 main_test.go

test-local:
	docker compose -f ./test/docker-compose.yml pull
	docker compose -f ./test/docker-compose.yml up -d
	$(MAKE) test
	docker compose -f ./test/docker-compose.yml down
