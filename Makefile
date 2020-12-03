SHELL:=/bin/bash

changelog_args=-o CHANGELOG.md -p '^v'
test_command=richgo test ./... $(TEST_ARGS) -v --cover

run-server: check-modd-exists
	@modd -f ./.modd/server.modd.conf

test: lint test-only

test-only:
	$(test_command)

lint: check-cognitive-complexity
	golangci-lint run --print-issued-lines=false --exclude-use-default=false --enable=golint --enable=goimports  --enable=unconvert --enable=unparam --concurrency=2 --timeout=10m

check-cognitive-complexity:
	gocognit -over 15 .

check-modd-exists:
	@modd --version > /dev/null

.PHONY: test test-only check-modd-exists

changelog:
ifdef version
	$(eval changelog_args=--next-tag $(version) $(changelog_args))
endif
	git-chglog $(changelog_args)

build:
	@pkger
	@go build -o bin/machinerydash .