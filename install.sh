#! /bin/zsh -ex

git pull
go get -v -d -u ./...
go mod tidy
git diff --quiet HEAD go.* || git add go.mod go.sum && git ci -m 'chore: go update' && git push -u

GIT_COMMIT=$(git rev-list -1 HEAD)
go build -ldflags "-X main.CommitSHA=$GIT_COMMIT" .
mv -v etu ~/bin/
