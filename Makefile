.PHONY: install_completion test

files := $(shell find . -name "*.go")

svlogj: ${files} go.mod Makefile
	go build -o $@

test:
	go test -v ./pkg/utils

install_completion: svlogj
	./svlogj --generate-completion=bash > ~/.bash_completion.d/svlogj

# vim:ft=Make:noexpandtab
