.PHONY: test

test:
	docker run -d -p 9222:9222 --rm --name headless-shell chromedp/headless-shell
	docker-compose -f test/docker-compose.yml up -d
	go test -v -race -timeout 30s main_test.go
	docker-compose -f test/docker-compose.yml stop
