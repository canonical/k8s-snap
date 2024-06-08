#!/bin/bash -eu

sudo DEBIAN_FRONTEND=noninteractive TZ=Etc/UTC apt-get install -y --no-install-recommends build-essential automake libtool gettext autopoint tclsh tcl libsqlite3-dev pkg-config git > /dev/null
