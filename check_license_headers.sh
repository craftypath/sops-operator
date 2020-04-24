#!/usr/bin/env bash

# Copyright The SOPS Operator Authors
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     https://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

set -o errexit
set -o nounset
set -o pipefail

regex_find() {
    local dir="$1"
    local pattern="$2"

    if [[ "$OSTYPE" == darwin* ]]; then
        find -E "$dir" -regex "$pattern"
    else
        find "$dir" -regextype posix-extended -regex "$pattern"
    fi
}

check_header() {
    local file="$1"

    grep -q '// Copyright The SOPS Operator Authors' "$file"
    grep -q 'https://www.apache.org/licenses/LICENSE-2.0' "$file"
}

main() {
    local files_without_header=()

    for file in $(regex_find . '.*\.(sh|go)'); do
        if ! check_header "$file"; then
            files_without_header+=("$file")
        fi
    done

    if [[ -n "${files_without_header[*]}" ]]; then
        echo "ERROR: Files without license header found:" >&2
        printf '%s\n' "${files_without_header[@]}" >&2
        exit 1
    fi
}

main
