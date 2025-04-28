#!/usr/bin/env bash

set -ex

PROJECT_BASE_DIR=$1
SCRIPT_DIR=$(realpath $(dirname "$BASH_SOURCE"))

if [[ -z $PROJECT_BASE_DIR ]]; then
    PROJECT_BASE_DIR="$SCRIPT_DIR/.."
fi

cd "${PROJECT_BASE_DIR}"

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
#
# NOTE: Running add-apt-repository -y ppa:dqlite/dev may flake with a 504 Gateway Time-out from Launchpad.
# Avoid this by adding the apt source lists manually.
# GPG signing key from: https://launchpad.net/~dqlite/+archive/ubuntu/dev
release="$(lsb_release --codename --short)"
wget -qO- "https://keyserver.ubuntu.com/pks/lookup?op=get&search=0x392A47B5A84EACA9B2C43CDA06CD096F50FB3D04" | sudo tee /etc/apt/trusted.gpg.d/dqlite-dev.asc
echo "deb-src https://ppa.launchpadcontent.net/dqlite/dev/ubuntu $release main" | sudo tee /etc/apt/sources.list.d/dqlite-dev.list
echo "deb https://ppa.launchpadcontent.net/dqlite/dev/ubuntu $release main" | sudo tee /etc/apt/sources.list.d/dqlite-dev.list
sudo apt-get update

sudo apt-get install -y dqlite-tools-v2 libdqlite1.17-dev
sudo make clean
go build -a ./...

TICSQServer -project k8s-snap -tmpdir /tmp/tics -branchdir $PROJECT_BASE_DIR
