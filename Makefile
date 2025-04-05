VERSION := $(shell cat VERSION)

proto-gen:
	protoc --go_out=. --go-grpc_out=. proto/messages.proto

release:
	git tag -a $(VERSION) -m "release $(VERSION)"
	git push origin $(VERSION)
