build:
	@go build -o bin/shiba ./cmd/shiba

run: build
	@./bin/shiba || echo "non-zero exit but continuing..."

buildvb:
	go build -o bin/vbrowser ./cmd/vbrowser

runvb: buildvb
	@./bin/vbrowser
