GO_VERSION := $(shell go version | cut -d' ' -f3 | sed -e 's/^go//' | cut -d. -f1,2)

test:
	go test

fmt:
	@ find . -type f -name \*.go -print0 | xargs -tr0n1 go fmt

tidy:
	go mod tidy -v -go=$(GO_VERSION) -compat=$(GO_VERSION)

fixup ft: fmt tidy

clean:
	git clean -dfx
	go clean -cache
	@make --no-print-directory tidy
