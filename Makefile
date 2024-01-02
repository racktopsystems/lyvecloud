.PHONY: build doc cover unittest

COVERAGE_DATA_DIR := coverage
COVERAGE_FILE := $(COVERAGE_DATA_DIR)/coverage.txt
COVERAGE_HTML_FILE := $(COVERAGE_DATA_DIR)/coverage.html

DOCS_DIR := docs/lyveapi
USAGE_DOCS_FILE := $(DOCS_DIR)/lyveapi-usage.txt

build:

cover:
	@echo "Generating library test coverage data"
	@go tool cover -html=$(COVERAGE_FILE) -o $(COVERAGE_HTML_FILE)

unittest:
	@echo "Running unit tests"
	@mkdir -p $(COVERAGE_DATA_DIR) && \
	go test -cover -coverpkg=./... -coverprofile $(COVERAGE_FILE) -v ./...

doc:
	@echo "Generating library usage documentation"
	@mkdir -p $(DOCS_DIR) && go doc -all ./lyveapi > $(USAGE_DOCS_FILE)