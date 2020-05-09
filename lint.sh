#! /bin/zsh
#
# Inspired by https://prettier.io/docs/en/precommit.html

jsfiles=$(git ls-tree --name-only -r HEAD | grep -e js -e css -e md)
[ -z "$jsfiles" ] && exit 0

for f in $(echo $jsfiles | xargs printf "%s\n"); do
  $(yarn bin)/prettier --write ${(q)f}
done

$(yarn bin)/eslint ./components ./lib

$(yarn bin)/depcheck .

yarn run build
