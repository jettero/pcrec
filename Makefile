GO_VERSION := $(shell go version | cut -d' ' -f3 | sed -e 's/^go//' | cut -d. -f1,2)

quick-test qtest qt:
	go test -failfast

test full-test ft:
	go test -v

fmt:
	@ find . -type f -name \*.go -print0 | xargs -tr0n1 go fmt

tidy:
	go mod tidy -v -go=$(GO_VERSION) -compat=$(GO_VERSION)

clean:
	git clean -dfx
	go clean -cache

fixup fu:
	@make --no-print-directory clean
	@make --no-print-directory fmt
	@make --no-print-directory tidy

clean-test ct:
	@make --no-print-directory fixup
	@make --no-print-directory test

update:
	go get -u ./...
	go mod tidy
	pre-commit clean
	pre-commit gc
	pre-commit autoupdate

