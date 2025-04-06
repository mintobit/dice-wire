VERSION := $(shell cat VERSION)

unittest:
	go test ./...

proto-gen:
	protoc --go_out=. --go-grpc_out=. proto/messages.proto

release:
	git tag -a $(VERSION) -m "release $(VERSION)"
	git push origin $(VERSION)
