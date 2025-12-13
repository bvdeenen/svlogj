.PHONY: install_bash_completion test clean

files := $(shell find . -name "*.go")
version:= $(shell git describe --tags HEAD)

svlogj: ${files} go.mod Makefile
	go build -o $@

test:
	go test -v ./pkg/utils

install_bash_completion: svlogj
	mkdir -p ~/.local/share/bash-completion/completion && \
	./svlogj completion bash > ~/.local/share/bash-completion/completions/svlogj
clean:
	rm svlogj
# vim:ft=Make:noexpandtab
