.PHONY: test

test:
	docker run -d -p 9222:9222 --rm --name headless-shell chromedp/headless-shell
	docker-compose -f test/docker-compose.yml up -d
	go mod download
	go test -v -race -timeout 60s -count 1 main_test.go
	docker stop headless-shell
	docker-compose -f test/docker-compose.yml stop
