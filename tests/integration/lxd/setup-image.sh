#!/bin/bash -xe

#### Description
#
# Build an LXD image for use in the k8s-snap integration tests. A number of base
# distros is supported (e.g. Ubuntu, Debian, AlmaLinux).
#
# Optionally, the script fetches the OCI images needed by k8s-snap, so that the
# images do not have to be pulled repeatedly by the end to end tests.
#
#### Configuration
#
# See next section for all configuration options that this script accepts as environment variables.
#
#### Examples
#
# Build LXD image 'k8s/ubuntu' based on Ubuntu 24.04
#
#    $ BASE_IMAGE=ubuntu:22.04 OUT_IMAGE_ALIAS=k8s/ubuntu ./setup-image.sh
#
# Build LXD image 'k8s/debian' based on Debian 12
#
#   $ BASE_IMAGE=images:debian/12 BASE_DISTRO=debian OUT_IMAGE_ALIAS=k8s/debian ./setup-image.sh
#
# Build LXD image 'k8s/almalinux' based on AlmaLinux 9
#
#   $ BASE_IMAGE=images:almalinux/9 BASE_DISTRO=almalinux OUT_IMAGE_ALIAS=k8s/almalinux ./setup.image.sh
#

DIR=`realpath $(dirname "${0}")`

################################################################################
# configuration

BASE_IMAGE="${BASE_IMAGE:=ubuntu:22.04}"                      # base image
BASE_DISTRO="${BASE_DISTRO:=ubuntu}"                          # base distro of the image

TEST_SNAP="${TEST_SNAP:=}"                                    # path to './k8s.snap' to test
BASE_SNAP="${BASE_SNAP:=core20}"                              # base snap to install on the image, e.g. 'core20'
IMAGES=""                                                     # list of images to fetch for side-loading

OUT_IMAGES_DIR="${OUT_IMAGES_DIR:=${DIR}/k8s-e2e-images}"     # directory where OCI images will be fetched
OUT_IMAGE_ALIAS="${OUT_IMAGE_ALIAS:=k8s-e2e}"                 # image alias to create

REGCTL="${REGCTL:=${DIR}/../../../src/k8s/tools/regctl.sh}"   # path to regctl binary

EXTRA_IMAGES="${EXTRA_IMAGES:=}"                              # space separated list of extra images to fetch for side-loading

################################################################################
# figure out base snap and list of images
if [ "${TEST_SNAP}" != "" ]; then
  dir="$(mktemp -d)"
  unsquashfs -d "${dir}/snap" "${TEST_SNAP}"

  BASE_SNAP="$(cat "${dir}/snap/meta/snap.yaml" | grep base: | head -n1 | sed "s/base: //")"
  IMAGES="$(cat "${dir}/snap/images.txt")"

  rm -rf "${dir}"
fi

################################################################################
# launch an instance from base image
lxc launch "${BASE_IMAGE}" tmp-builder

################################################################################
# distro specific steps
case "${BASE_DISTRO}" in
  ubuntu)
    # snapd is preinstalled on Ubuntu OSes
    lxc shell tmp-builder -- bash -c 'snap wait core seed.loaded'
    lxc shell tmp-builder -- bash -c 'snap install '"${BASE_SNAP}"
    ;;
  almalinux)
    # install snapd and ensure /snap/bin is in the environment
    lxc shell tmp-builder -- bash -c 'while ! ping -c1 snapcraft.io; do sleep 1; done'
    lxc shell tmp-builder -- bash -c 'dnf install epel-release -y'
    lxc shell tmp-builder -- bash -c 'dnf install tar sudo -y'
    lxc shell tmp-builder -- bash -c 'dnf install fuse squashfuse -y'
    lxc shell tmp-builder -- bash -c 'dnf install snapd -y'

    lxc shell tmp-builder -- bash -c 'systemctl enable --now snapd.socket'
    lxc shell tmp-builder -- bash -c 'ln -s /var/lib/snapd/snap /snap'
    lxc shell tmp-builder -- bash -c 'snap wait core seed.loaded'
    lxc shell tmp-builder -- bash -c 'snap install snapd '"${BASE_SNAP}"
    lxc shell tmp-builder -- bash -c 'echo PATH=$PATH:/snap/bin >> /etc/environment'
    ;;
  debian)
    # install snapd and ensure /snap/bin is in the environment
    lxc shell tmp-builder -- bash -c 'apt update'
    lxc shell tmp-builder -- bash -c 'apt install -y squashfuse snapd fuse'
    lxc shell tmp-builder -- bash -c 'snap wait core seed.loaded'
    lxc shell tmp-builder -- bash -c 'snap install snapd '"${BASE_SNAP}"
    lxc shell tmp-builder -- bash -c 'echo PATH=$PATH:/snap/bin >> /etc/environment'
    lxc shell tmp-builder -- bash -c 'apt autoremove; apt clean; apt autoclean; rm -rf /var/lib/apt/lists'

    # NOTE(neoaggelos): disable apparmor in containerd, as it causes trouble in the default setup
    lxc shell tmp-builder -- bash -c '
      mkdir -p /var/snap/k8s/common/etc/containerd/conf.d
      echo "
      [plugins.\"io.containerd.grpc.v1.cri\"]
        disable_apparmor=true
      " | tee /var/snap/k8s/common/etc/containerd/conf.d/10-debian-disable-apparmor.toml
    '
    ;;
  *)
    echo "Unsupported BASE_DISTRO value: ${BASE_DISTRO}"
    exit 1
    ;;
esac

################################################################################
# create snapshot and export as image
lxc snapshot tmp-builder snapshot
lxc publish local:tmp-builder/snapshot --alias "${OUT_IMAGE_ALIAS}"

################################################################################
# cleanup
lxc rm tmp-builder --force

################################################################################
# fetch images
mkdir -p "${OUT_IMAGES_DIR}"
for image in ${IMAGES} ${EXTRA_IMAGES}; do
  file="${OUT_IMAGES_DIR}/$(echo $image | tr ':/' '-').tar"
  [ ! -f "${file}" ] && "${REGCTL}" image export --platform=local --user-agent=containerd/v1.6.33 "${image}" "${file}"
done
