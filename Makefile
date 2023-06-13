build:
	go build -o bin/4byteScraper

run: build
	./bin/4byteScraper

clean:
	rm -rf bin

.PHONY: build run clean