MAKEFLAGS += --silent
SHELL = bash

linux:
	GOOS=linux GOARCH=amd64 go build
mac:
	GOOS=darwin GOARCH=amd64 go build

fmt: prettier

prettier:
	docker run -it --rm \
		-v $$(pwd):/work \
		--user $$(id -u):$$(id -g) \
		jauderho/prettier:2.8.1-alpine \
		--write \
			.github/ \
			README.md