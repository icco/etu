#! /bin/sh

go get -v -d ./...
go build .
mv -v etu ~/bin/
