# Bootstrap issues on a host with custom routing policy rules

## Problem

{{product}} bootstrap process might fail or face networking issues when
custom routing policy rules are defined, such as rules in a netplan file.

## Explanation

Cilium, which is the current implementation for the `network` feature,
introduces and adjusts certain ip rules with
hard-coded priorities of `0` and `100`.

Adding ip rules with a priority lower than or equal to `100` might introduce
conflicts and cause networking issues.

## Solution

Adjust the custom defined `ip rule` to have a
priority value that is greater than `100`.
