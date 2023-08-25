#!/usr/bin/env bash

go install -v golang.org/x/tools/cmd/goimports@latest
go install -v github.com/fzipp/gocyclo/cmd/gocyclo@latest
go install -v github.com/golangci/golangci-lint/cmd/golangci-lint@latest
go install -v github.com/go-critic/go-critic/cmd/gocritic@latest
go install -v github.com/BurntSushi/toml/cmd/tomlv@latest
go install -v mvdan.cc/gofumpt@latest
go install -v github.com/kisielk/errcheck@latest
# go install -v github.com/quasilyte/go-ruleguard/cmd/ruleguard@latest
go install -v honnef.co/go/tools/cmd/staticcheck@latest
