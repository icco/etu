#! /bin/bash
#
# Inspired by https://prettier.io/docs/en/precommit.html

yarn

jsfiles=$(git ls-tree --name-only -r HEAD | grep -e js -e css)
[ -z "$jsfiles" ] && exit 0

echo "$jsfiles" | xargs $(yarn bin)/prettier --write
