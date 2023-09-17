.PHONY: build doc cover unit-test

COVERAGE_DATA_DIR = coverage
COVERAGE_FILE = $(COVERAGE_DATA_DIR)/coverage.out

build:

cover:
	go tool cover -html=$(COVERAGE_FILE)

unit-test:
	mkdir -p $(COVERAGE_DATA_DIR) && \
	go test -cover -coverpkg=./... -coverprofile $(COVERAGE_FILE) -v ./...

doc:
	go doc -all ./lyveapi