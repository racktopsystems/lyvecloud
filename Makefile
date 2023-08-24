.PHONY: cover, unit-test

COVERAGE_DATA_DIR = coverage
COVERAGE_FILE = $(COVERAGE_DATA_DIR)/coverage.out

cover:
	go tool cover -html=$(COVERAGE_FILE)

unit-test:
	go test -cover -coverpkg=./... -coverprofile $(COVERAGE_FILE) -v ./...