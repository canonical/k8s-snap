#!/usr/bin/env bash

SCRIPT_DIR=$(realpath $(dirname "$BASH_SOURCE"))

set -ex
cd "${SCRIPT_DIR}/.."

sudo apt-get update
sudo apt-get install -y python3-venv
python3 -m venv .venv/tics
source .venv/tics/bin/activate

# Install python dependencies
pip install -r tests/integration/requirements-test.txt
pip install -r tests/integration/requirements-dev.txt

cd src/k8s

# TICS requires us to have the test results in cobertura xml format under the
# directory use below
sudo make go.unit
go install github.com/boumenot/gocover-cobertura@latest
gocover-cobertura < coverage.txt > coverage.xml
mkdir -p .coverage
mv ./coverage.xml ./.coverage/

# Install the TICS and staticcheck
go install honnef.co/go/tools/cmd/staticcheck@v0.5.1
. <(curl --silent --show-error 'https://canonical.tiobe.com/tiobeweb/TICS/api/public/v1/fapi/installtics/Script?cfg=default&platform=linux&url=https://canonical.tiobe.com/tiobeweb/TICS/')

# We need to have our project built
# We load the dqlite libs here instead of doing through make because TICS
# will try to build parts of the project itself
sudo add-apt-repository -y ppa:dqlite/dev
sudo apt-get install -y dqlite-tools-v2 libdqlite1.17-dev
sudo make clean
go build -a ./...

TICSQServer -project k8s-snap -tmpdir /tmp/tics -branchdir $SCRIPT_DIR/..
