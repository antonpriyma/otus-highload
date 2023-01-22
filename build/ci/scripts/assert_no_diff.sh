#!/bin/bash

set -exu

diff_files=$(git diff --name-only)
if [[ -n "$diff_files" ]]; then
    echo "found diff $diff_files"
    git diff &> diff.log
    git status
    exit 1
fi

exit 0
