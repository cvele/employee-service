GOHOSTOS:=$(shell go env GOHOSTOS)
GOPATH:=$(shell go env GOPATH)
VERSION=$(shell git describe --tags --always)

ifeq ($(GOHOSTOS), windows)
	#the `find.exe` is different from `find` in bash/shell.
	#to see https://docs.microsoft.com/en-us/windows-server/administration/windows-commands/find.
	#changed to use git-bash.exe to run find cli or other cli friendly, caused of every developer has a Git.
	#Git_Bash= $(subst cmd\,bin\bash.exe,$(dir $(shell where git)))
	Git_Bash=$(subst \,/,$(subst cmd\,bin\bash.exe,$(dir $(shell where git))))
	INTERNAL_PROTO_FILES=$(shell $(Git_Bash) -c "find internal -name *.proto")
	API_PROTO_FILES=$(shell $(Git_Bash) -c "find api -name *.proto")
else
	INTERNAL_PROTO_FILES=$(shell find internal -name *.proto)
	API_PROTO_FILES=$(shell find api -name *.proto)
endif

.PHONY: init
# init env
init:
	go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
	go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
	go install github.com/go-kratos/kratos/cmd/kratos/v2@latest
	go install github.com/go-kratos/kratos/cmd/protoc-gen-go-http/v2@latest
	go install github.com/google/gnostic/cmd/protoc-gen-openapi@latest
	go install github.com/google/wire/cmd/wire@latest

.PHONY: config
# generate internal proto
config:
	protoc --proto_path=./internal \
	       --proto_path=./third_party \
 	       --go_out=paths=source_relative:./internal \
	       $(INTERNAL_PROTO_FILES)

.PHONY: api
# generate api proto
api:
	protoc --proto_path=./api \
	       --proto_path=./third_party \
 	       --go_out=paths=source_relative:./api \
 	       --go-http_out=paths=source_relative:./api \
 	       --go-grpc_out=paths=source_relative:./api \
	       --openapi_out=fq_schema_naming=true,default_response=false:. \
	       $(API_PROTO_FILES)

.PHONY: events
# generate event protos
events:
	protoc --proto_path=./api \
	       --proto_path=./third_party \
	       --go_out=paths=source_relative:./api \
	       api/events/v1/*.proto

.PHONY: build
# build
build:
	mkdir -p bin/ && go build -ldflags "-X main.Version=$(VERSION)" -o ./bin/ ./...

.PHONY: test
# run unit tests
test:
	go test -v -race -coverprofile=coverage.out ./...

.PHONY: test-coverage
# run tests with coverage report
test-coverage: test
	go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

.PHONY: generate
# generate
generate:
	go generate ./...
	go mod tidy

.PHONY: all
# generate all
all:
	make api
	make events
	make config
	make generate

.PHONY: migrate-up
# migrate up
migrate-up:
	go run cmd/migrate/main.go -command up

.PHONY: migrate-down
# migrate down
migrate-down:
	go run cmd/migrate/main.go -command down -steps 1

.PHONY: migrate-status
# migration status
migrate-status:
	go run cmd/migrate/main.go -command version

.PHONY: migrate-force
# force migration version (use with caution)
migrate-force:
	@echo "Usage: make migrate-force VERSION=1"
	@if [ -z "$(VERSION)" ]; then echo "ERROR: VERSION is required"; exit 1; fi
	go run cmd/migrate/main.go -command force -steps $(VERSION)

.PHONY: migrate-create
# create new migration
migrate-create:
	@echo "Usage: make migrate-create NAME=add_column_to_employees"
	@if [ -z "$(NAME)" ]; then echo "ERROR: NAME is required"; exit 1; fi
	@NEXT=$$(ls migrations/*.up.sql 2>/dev/null | wc -l | xargs); \
	NEXT=$$((NEXT + 1)); \
	NEXT_PADDED=$$(printf "%06d" $$NEXT); \
	touch migrations/$${NEXT_PADDED}_$(NAME).up.sql; \
	touch migrations/$${NEXT_PADDED}_$(NAME).down.sql; \
	echo "Created migrations/$${NEXT_PADDED}_$(NAME).{up,down}.sql"

.PHONY: docker-up
# start docker services
docker-up:
	docker-compose up -d postgres

.PHONY: docker-down
# stop docker services
docker-down:
	docker-compose down

.PHONY: docker-dev
# start dev docker services
docker-dev:
	docker-compose -f docker-compose.dev.yml up -d

.PHONY: docker-logs
# view docker logs
docker-logs:
	docker-compose logs -f


.PHONY: consumer
# run event consumer (for testing)
consumer:
	go run cmd/consumer/main.go

# show help
help:
	@echo ''
	@echo 'Usage:'
	@echo ' make [target]'
	@echo ''
	@echo 'Targets:'
	@awk '/^[a-zA-Z\-\_0-9]+:/ { \
	helpMessage = match(lastLine, /^# (.*)/); \
		if (helpMessage) { \
			helpCommand = substr($$1, 0, index($$1, ":")); \
			helpMessage = substr(lastLine, RSTART + 2, RLENGTH); \
			printf "\033[36m%-22s\033[0m %s\n", helpCommand,helpMessage; \
		} \
	} \
	{ lastLine = $$0 }' $(MAKEFILE_LIST)

.DEFAULT_GOAL := help
