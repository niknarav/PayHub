.PHONY: proto-lint proto-gen

proto-lint:
	cd contracts && buf lint

proto-gen: proto-lint
	cd contracts && buf generate
