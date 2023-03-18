GOPRIVATE := gitlab.corp.mail.ru,stash.mail.ru
export GOPRIVATE

ROOT_PATH := $(PWD)
BIN_PATH := $(ROOT_PATH)/third_party
PATH := $(BIN_PATH):$(PATH)
export ROOT_PATH

.PHONY: build
build:
	go build ./...

.PHONY: goimports
goimports: third_party/goimports
	find .\
		-type f\
		-name '*.go'\
		\! -name '*_easyjson.go'\
		\! -name '*mock_generated.go'\
		\! -name '*_mock.go'\
		\! -path './.git/*'\
		\! -path './vendor/*' | xargs goimports -w

.PHONY: lint
lint: third_party/golangci-lint
	GOGC=1000 ./third_party/golangci-lint run -c ./build/ci/golangci.yml -v $(ROOT_PATH)/...
	GOGC=1000 ./third_party/golangci-lint run -c ./build/ci/golangci.yml -c ./build/ci/golangci_tests.yml -v $(ROOT_PATH)/...

third_party/golangci-lint:
	curl https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b ./third_party v1.23.1

.PHONY: test
test:
	go test --coverprofile=.coverprofile --coverpkg=./... ./...
	fgrep -v "mock_generated.go" .coverprofile > .coverprofile_excl
	go tool cover -func=.coverprofile_excl

.PHONY: tidyvendor
tidyvendor:
	go mod tidy
	go mod vendor

.PHONY: generate
generate: third_party/mockgen
	 go generate ./...

third_party/mockgen:
	go build -o $(ROOT_PATH)/$(@) github.com/golang/mock/mockgen

third_party/goimports:
	go build -o $(ROOT_PATH)/$(@) golang.org/x/tools/cmd/goimports

.PHONY: groupimports
groupimports:
	find .\
		-type f\
		-name '*.go'\
		\! -name '*_easyjson.go'\
		\! -name '*_mock.go'\
		\! -name '*generated.go'\
		\! -path './.git/*'\
		\! -path '*.resolvers.go'\
		\! -path './vendor/*' \
		\! -name '*mock_generated.go' | xargs -L 1 python3 build/ci/scripts/group_imports.pys

.PHONY: hardecho
hardecho:
	go get gitlab.corp.mail.ru/calendar/hardecho@$(VER)
	make tidyvendor


