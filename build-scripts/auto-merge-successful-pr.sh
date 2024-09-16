#!/bin/bash

set -e

# Fetch the open pull requests
prs=$(gh pr list --state open --json number,headRefName,labels | jq '[.[] | select(.labels | any(.name == "automerge"))]')

for pr in $(echo "$prs" | jq -r '.[] | @base64'); do
    _jq() {
        echo ${pr} | base64 --decode | jq -r ${1}
    }

    pr_number=$(_jq '.number')
    head_branch=$(_jq '.headRefName')

    # Check status checks for each PR
    checks_passed=$(gh pr checks $pr_number --json bucket | jq -r '.[].bucket == "pass"' | sort | uniq)

if [[ "$checks_passed" == "true" ]]; then
    echo "All status checks passed for PR #$pr_number. Proceeding with merge..."
    gh pr review $pr_number --approve -b "All status checks passed for PR #$pr_number."
    gh pr merge $pr_number --auto --squash
else
    echo "Status checks have not passed for PR #$pr_number. Skipping merge."
fi
done
