#!/bin/bash

function repeat_checks {
  local cmd="$1"
  local expected="$2"
  local max_retries="${3:-15}"
  local interval="${4:-2}"

  for i in $(seq 1 "$max_retries"); do
    echo "Attempt ${i}/${max_retries}: ${cmd}"
    bash -lc "$cmd" > output.txt 2>&1
    cat output.txt

    if grep -q "$expected" output.txt; then
      return 0
    fi

    if [[ "$i" == "$max_retries" ]]; then
      echo "repeat_checks failed after ${max_retries} attempts"
      return 1
    fi

    sleep "$interval"
  done
}
