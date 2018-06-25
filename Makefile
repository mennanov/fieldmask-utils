# vi: ft=make

GOPATH:=$(shell go env GOPATH)

.PHONY: proto test test-with-coverage

proto:
	go get github.com/golang/protobuf/protoc-gen-go
	protoc -I . -I ${GOPATH}/src test.proto --go_out=${GOPATH}/src

test:
	@go get github.com/rakyll/gotest
	gotest -p 1 -v ./...


test-with-coverage:
	@go get github.com/rakyll/gotest
	gotest -p 1 -v -covermode=count -coverprofile=coverage.out ./...
	goveralls -coverprofile=coverage.out -service travis-ci -repotoken ${COVERALLS_TOKEN}
