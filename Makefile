.PHONY: run test lint index eval up capture

run:            ## Start the web app on http://localhost:8080
	go run ./cmd/ragxray

test:           ## Unit tests (no model needed)
	go test -race ./...

lint:           ## Format check and vet
	@test -z "$$(gofmt -l $$(go list -f {{.Dir}} ./...))" || (echo "run gofmt -w"; exit 1)
	go vet ./...

index:          ## Build store/index.json from data/docs (lesson 3)
	go run ./lessons/03-index

eval:           ## Retrieval and answer evaluation (lesson 6)
	go run ./lessons/06-eval -answers

up:             ## Everything in Docker, including Ollama
	docker compose up --build

capture:        ## Regenerate README screenshots and demo video (needs playwright-core)
	node scripts/capture.mjs
