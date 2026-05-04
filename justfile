set shell := ["bash", "-eu", "-o", "pipefail", "-c"]

default:
	@just --list

fmt:
	unformatted="$(gofmt -l .)"; \
	if [[ -n "$$unformatted" ]]; then \
	  echo "Unformatted files:"; \
	  echo "$$unformatted"; \
	  exit 1; \
	fi

vet:
	go vet ./...

lint:
	golangci-lint run ./...

test:
	go test ./...

race:
	go test -race ./...

check:
	just fmt
	just vet
	just lint
	just test
	just race
