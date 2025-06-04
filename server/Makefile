build:
	@go build -o bin/shiba ./cmd/shiba

run: build
	@./bin/shiba

buildvb:
	go build -o bin/vbrowser ./cmd/vbrowser

runvb: buildvb
	@./bin/vbrowser
