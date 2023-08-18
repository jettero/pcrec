GO_VERSION := $(shell go version | cut -d' ' -f3 | sed -e 's/^go//' | cut -d. -f1,2)

test:
	go test

clean:
	git clean -dfx
	go clean -cache
	go mod tidy -v -go=$(GO_VERSION) -compat=$(GO_VERSION)
