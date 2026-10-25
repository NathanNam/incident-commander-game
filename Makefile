# Incident Commander: Go → WebAssembly, served as static files from web/.
.PHONY: build serve test brand shots clean

GOROOT := $(shell go env GOROOT)

# Rebuild web/game.wasm and refresh the matching wasm_exec.js loader (both committed).
build:
	GOOS=js GOARCH=wasm go build -trimpath -ldflags="-s -w" -o web/game.wasm ./cmd/game
	cp "$(GOROOT)/lib/wasm/wasm_exec.js" web/wasm_exec.js

# Play locally at http://localhost:8080
serve: build
	go run ./cmd/server

test:
	go test ./internal/game
	GOOS=js GOARCH=wasm go vet ./...

# Favicon set + share image (needs Chrome and: npm i --no-save puppeteer-core@23.11.1)
brand:
	node scripts/brand/make.mjs

# README screenshots from a running `make serve`
shots:
	node scripts/brand/make.mjs --shots http://localhost:8080/

clean:
	rm -rf node_modules
