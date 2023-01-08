#!/usr/bin/env bash
set -ev
go vet
test -z "$(go fmt ./...)" # fail if not formatted properly
go test -v -race -coverprofile=coverage.out -covermode=atomic ./...
ls -d logadapter/* | xargs -I {} bash -c "cd '{}' \
&& go mod tidy && go mod vendor \
&& go test -race -coverprofile=coverage.out -covermode=atomic -coverpkg=./... ./... \
&& cat coverage.out | grep -v \"mode:\" >> ../../coverage.out \
&& rm coverage.out"
# for go repo with nested modules, remove repo prefix, otherwise goveralls will failed.
sed -i -e 's/github.com\/simukti\/sqldb-logger/./g' coverage.out
