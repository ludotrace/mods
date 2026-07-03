# lt-validate — the LudoTrace events-file validator.
# VERSION is the single source of truth; `release` tags from it, CI builds from the tag.

VERSION := $(shell cat VERSION)
LDFLAGS := -X main.version=$(VERSION)

.PHONY: build test release

build:  ## compile lt-validate with the version stamped in
	go build -ldflags "$(LDFLAGS)" -o lt-validate .

test: build  ## valid fixture must pass, invalid must fail
	./lt-validate examples/valid.jsonl
	@if ./lt-validate examples/invalid.jsonl; then \
		echo "FAIL: expected examples/invalid.jsonl to be rejected" >&2; exit 1; \
	else \
		echo "OK: invalid fixture rejected as expected"; \
	fi

release:  ## guard, then tag from VERSION and push — CI builds and publishes the release
	@test -z "$$(git status --porcelain)" || { echo "working tree not clean" >&2; exit 1; }
	@test "$$(git rev-parse --abbrev-ref HEAD)" = "main" || { echo "not on main" >&2; exit 1; }
	git tag $(VERSION)
	git push origin $(VERSION)
