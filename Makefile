NAME=ovh-ddns
MAIN=cmd/$(NAME)/$(NAME).go
OUT=bin/$(NAME)
COVERAGE_REPORT=test/coverage.out
COVERAGE_REPORT_HTML=test/coverage.html

.PHONY: all
all: build

.PHONY: build
build:
	@CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o $(OUT) $(MAIN)

.PHONY: image
image:
	@docker build --target runtime -t $(NAME):latest .

.PHONY: compose-up
compose-up:
	@mkdir -p test/data
	@docker compose -f test/compose.yaml up
	@make compose-down

.PHONY: compose-down
compose-down:
	@docker compose -f test/compose.yaml down

.PHONY: run
run:
	@go run -tags dev $(MAIN)

.PHONY: test
test:
	@mkdir -p test
	@go test -tags dev -cover -coverprofile=$(COVERAGE_REPORT) ./...

.PHONY: test-cicd
test-cicd:
	@mkdir -p test
	@CGO_ENABLED=1 go test -tags dev -v -race -cover -coverprofile=$(COVERAGE_REPORT) ./...

.PHONY: benchmark
benchmark:
	@go test -tags dev -bench=. -benchmem -run =^a ./...

.PHONY: coverage
coverage: test
	@go tool cover -html=$(COVERAGE_REPORT) -o=$(COVERAGE_REPORT_HTML)
	@xdg-open $(COVERAGE_REPORT_HTML)

.PHONY: lint
lint:
	@golangci-lint run

.PHONY: clean
clean:
	@rm -r bin

.PHONY: gitclean
gitclean:
	@git clean -xdf
