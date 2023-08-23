.PHONY: unit-test

COVERAGE_DATA_DIR = coverage
COVERAGE_FILE = $(COVERAGE_DATA_DIR)/coverage.out

unit-test:
	go test -cover -coverpkg=./... -coverprofile $(COVERAGE_FILE) -v ./...