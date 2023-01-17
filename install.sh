#! /bin/sh

git pull
go get -v -d -u ./...
go mod tidy -compat=1.19
go build .
git add go.mod go.sum && git ci -m 'chore: go update'
git push -u
mv -v etu ~/bin/
