#!/usr/bin/env bash

DIR=`realpath $(dirname "${0}")`

# Initialize node for integration tests
"${DIR}/connect-interfaces.sh"
"${DIR}/network-requirements.sh"
