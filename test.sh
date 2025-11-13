  cd tests/integration
  # Split the input tags into an array and pass them as individual args
  IFS=' ' read -ra TAGS <<< "performance"
  TAGS_ARGS=""
  for tag in "${TAGS[@]}"; do
    TAGS_ARGS="$TAGS_ARGS --tags $tag"
  done

  # Collect test names and convert to JSON array for GitHub Actions
  # There is no easy way to get the test names from pytest, so we use the --collect-only flag
  # and parse the output to extract the test names.
  TEST_FILES=$(tox -e integration -- --collect-only $TAGS_ARGS --quiet --no-header tests/ | grep :: || true)

  # Convert to JSON array for GitHub Actions
  if [ -z "$TEST_FILES" ]; then
    JSON_ARRAY='[]'
  else
    JSON_ARRAY=$(printf '%s\n' "$TEST_FILES" | jq -R -s -c 'split("\n") | map(select(length > 0))')
  fi
  echo "Found test files: $TEST_FILES ($JSON_ARRAY)"
