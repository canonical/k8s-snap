#!/bin/bash

# This script automates various operations on 2-node HA A-A Canonical K8s
# clusters that use the default datastore, Dqlite.
#
# Prerequisites:
# * required packages installed using the "install_packages" command.
# * initialized K8s cluster, both nodes joined
# * the current user has ssh access to the peer node.
#   - used to handle K8s services and transfer Dqlite data
# * the current user has passwordless sudo enabled.
sourced=0

DEBUG=${DEBUG:-0}
if [[ "${BASH_SOURCE[0]}" != "${0}" ]]; then
    sourced=1
else
    sourced=0
    set -eEu -o pipefail

    if [[ $DEBUG -eq 1 ]]; then
        export PS4='+(${BASH_SOURCE}:${LINENO}): ${FUNCNAME[0]:+${FUNCNAME[0]}(): }'
        set -x
    fi
fi

SYSTEMD_SERVICE_NAME=${SYSTEMD_SERVICE_NAME:-"2ha_k8s"}
DRBD_MOUNT_DIR=${DRBD_MOUNT_DIR:-"/mnt/drbd0"}
SSH_USERNAME=${SSH_USERNAME:-"ubuntu"}
SSH_OPTS=${SSH_OPTS:-"-o StrictHostKeyChecking=no -o ConnectTimeout=5"}
K8SD_LOG_LEVEL=${K8SD_LOG_LEVEL:-"0"}
K8S_SNAP_CHANNEL=${K8S_SNAP_CHANNEL:-"latest/edge"}
DRBD_RES_NAME=${DRBD_RES_NAME:-"r0"}
DRBD_READY_TIMEOUT=${DRBD_READY_TIMEOUT:-30}
PEER_READY_TIMEOUT=${PEER_READY_TIMEOUT:-60}

K8SD_PATH=${K8SD_PATH:-/snap/k8s/current/bin/k8sd}

K8S_DQLITE_STATE_DIR=/var/snap/k8s/common/var/lib/k8s-dqlite
K8SD_STATE_DIR="/var/snap/k8s/common/var/lib/k8sd/state"

K8S_DQLITE_STATE_BKP_DIR=/var/snap/k8s/common/var/lib/k8s-dqlite.bkp
K8SD_STATE_BKP_DIR="/var/snap/k8s/common/var/lib/k8sd/state.bkp"

K8S_DQLITE_INFO_YAML="$K8S_DQLITE_STATE_DIR/info.yaml"
K8S_DQLITE_CLUSTER_YAML="$K8S_DQLITE_STATE_DIR/cluster.yaml"

K8SD_INFO_YAML="$K8SD_STATE_DIR/database/info.yaml"
K8SD_CLUSTER_YAML="$K8SD_STATE_DIR/database/cluster.yaml"

# Backup yamls are expected to contain the right node ids and
# addresses while the DRBD files may contain settings from the other node
# and have to be updated.
K8S_DQLITE_INFO_BKP_YAML="$K8S_DQLITE_STATE_BKP_DIR/info.yaml"
K8S_DQLITE_CLUSTER_BKP_YAML="$K8S_DQLITE_STATE_BKP_DIR/cluster.yaml"
K8SD_INFO_BKP_YAML="$K8SD_STATE_BKP_DIR/database/info.yaml"
K8SD_CLUSTER_BKP_YAML="$K8SD_STATE_BKP_DIR/database/cluster.yaml"

K8SD_RECOVERY_TARBALL="$K8SD_STATE_DIR/recovery_db.tar.gz"
# K8SD will remove this file upon starting. We need to create a backup that
# can be transferred to other nodes.
K8SD_RECOVERY_TARBALL_BKP="$K8SD_STATE_DIR/recovery_db.bkp.tar.gz"

DQLITE_ROLE_VOTER=0
DQLITE_ROLE_STANDBY=1
DQLITE_ROLE_SPARE=2

function log_message () {
    local msg="[$(date -uIseconds)] $@"
    >&2 echo -e "$msg"
}

function get_dqlite_node_id() {
    local infoYamlPath=$1
    sudo cat $infoYamlPath | yq -r '.ID'
}

function get_dqlite_node_addr() {
    local infoYamlPath=$1
    sudo cat $infoYamlPath | yq -r '.Address'
}

function get_dqlite_node_role() {
    local infoYamlPath=$1
    sudo cat $infoYamlPath | yq -r '.Role'
}

function get_dqlite_role_from_cluster_yaml() {
    # Note that the cluster.yaml role may not match the info.yaml role.
    # In case of a freshly joined node, info.yaml will have "voter" role
    # while cluster.yaml has "spare" role.
    local clusterYamlPath=$1
    local nodeId=$2

    # Update the specified node.
    sudo cat $clusterYamlPath | \
        yq -r "(.[] | select(.ID == \"$nodeId\") | .Role )"
}

function set_dqlite_node_role() {
    # The yq snap installs in confined mode, so it's unable to access the
    # dqlite config files.
    # In order to modify files in-place, we're using sponge. It reads all
    # the stdin data before opening the output file.
    local infoYamlPath=$1
    local role=$2
    sudo cat $infoYamlPath | \
        yq ".Role = $role" |
        sudo sponge $infoYamlPath
}

# Update cluster.yaml, setting the specified node as voter (role = 0).
# The other nodes will become spares, having the role set to 2.
function set_dqlite_node_as_sole_voter() {
    local clusterYamlPath=$1
    local nodeId=$2

    # Update the specified node.
    sudo cat $clusterYamlPath | \
        yq "(.[] | select(.ID == \"$nodeId\") | .Role ) = 0" | \
        sudo sponge $clusterYamlPath

    # Update the other nodes.
    sudo cat $clusterYamlPath | \
        yq "(.[] | select(.ID != \"$nodeId\") | .Role ) = 2" | \
        sudo sponge $clusterYamlPath
}

function get_dql_peer_ip() {
    local clusterYamlPath=$1
    local nodeId=$2

    local addresses=( $(sudo cat $clusterYamlPath | \
         yq "(.[] | select(.ID != \"$nodeId\") | .Address )") )

    if [[ ${#addresses[@]} -gt 1 ]]; then
        log_message "More than one dql peers found: ${addresses[@]}"
        exit 1
    fi

    if [[ ${#addresses[@]} -lt 1 ]]; then
        log_message "No dql peers found."
        exit 1
    fi

    echo ${addresses[0]} | cut -d ":" -f 1
}

# This function moves the dqlite state directories to the DRBD mount,
# replacing them with symlinks. This ensures that the primary will always use
# the latest DRBD data.
#
# The existing contents are moved to a backup folder, which can be used as
# part of the recovery process.
function move_statedirs() {
    sudo mkdir -p $DRBD_MOUNT_DIR/k8s-dqlite
    sudo mkdir -p $DRBD_MOUNT_DIR/k8sd

    log_message "Validating dqlite state directories."
    check_statedir $K8S_DQLITE_STATE_DIR $DRBD_MOUNT_DIR/k8s-dqlite
    check_statedir $K8SD_STATE_DIR $DRBD_MOUNT_DIR/k8sd

    if [[ ! -L $K8S_DQLITE_STATE_DIR ]] || [[ ! -L $K8SD_STATE_DIR ]]; then
        local k8sDqliteNodeId=`get_dqlite_node_id $K8S_DQLITE_INFO_YAML`
        if [[ -z $k8sDqliteNodeId ]]; then
            log_message "Couldn't retrieve k8s-dqlite node id."
            exit 1
        fi


        local expRole=`get_expected_dqlite_role`
        # For fresh k8s clusters, the info.yaml role may not match the cluster.yaml role.
        local k8sDqliteRole=`get_dqlite_role_from_cluster_yaml \
            $K8S_DQLITE_CLUSTER_YAML $k8sDqliteNodeId`

        if [[ $expRole -ne $k8sDqliteRole ]]; then
            # TODO: consider automating this. We may move the pacemaker resource
            # ourselves and maybe even copy the remote files through scp or ssh.
            # However, there's a risk of race conditions.
            log_message "DRBD volume mounted on replica, refusing to transfer dqlite files."
            log_message "Move the DRBD volume to the primary node (through the fs_res Pacemaker resource) and try again."
            log_message "Example: sudo crm resource move fs_res <primary_node> && sudo crm resource clear fs_res"
            exit 1
        fi
    fi

    # Ensure that the k8s services are stopped.
    log_message "Stopping k8s services."
    sudo snap stop k8s

    if [[ ! -L $K8S_DQLITE_STATE_DIR ]]; then
        log_message "Not a symlink: $K8S_DQLITE_STATE_DIR, " \
                    "transferring to $DRBD_MOUNT_DIR/k8s-dqlite"
        sudo cp -r $K8S_DQLITE_STATE_DIR/. $DRBD_MOUNT_DIR/k8s-dqlite

        log_message "Creating k8s-dqlite state dir backup: $K8S_DQLITE_STATE_BKP_DIR"
        sudo rm -rf $K8S_DQLITE_STATE_BKP_DIR
        sudo mv $K8S_DQLITE_STATE_DIR/ $K8S_DQLITE_STATE_BKP_DIR

        log_message "Creating symlink $K8S_DQLITE_STATE_DIR -> $DRBD_MOUNT_DIR/k8s-dqlite"
        sudo ln -sf $DRBD_MOUNT_DIR/k8s-dqlite $K8S_DQLITE_STATE_DIR
    else
        log_message "Symlink $K8S_DQLITE_STATE_DIR points to $DRBD_MOUNT_DIR/k8s-dqlite"
    fi

    if [[ ! -L $K8SD_STATE_DIR ]]; then
        log_message "Not a symlink: $K8SD_STATE_DIR, " \
                    "transferring to $DRBD_MOUNT_DIR/k8sd"
        sudo cp -r $K8SD_STATE_DIR/. $DRBD_MOUNT_DIR/k8sd

        log_message "Creating k8sd state dir backup: $K8SD_STATE_BKP_DIR"
        sudo rm -rf $K8SD_STATE_BKP_DIR
        sudo mv $K8SD_STATE_DIR/ $K8SD_STATE_BKP_DIR

        log_message "Creating symlink $K8SD_STATE_DIR -> $DRBD_MOUNT_DIR/k8sd"
        sudo ln -sf $DRBD_MOUNT_DIR/k8sd $K8SD_STATE_DIR
    else
        log_message "Symlink $K8SD_STATE_DIR points to $DRBD_MOUNT_DIR/k8sd"
    fi
}

function ensure_mount_rw() {
    if ! mount | grep "on $DRBD_MOUNT_DIR type" &> /dev/null; then
        log_message "Missing DRBD mount: $DRBD_MOUNT_DIR"
        return 1
    fi

    if ! mount | grep "on $DRBD_MOUNT_DIR type" | grep "rw" &> /dev/null; then
        log_message "DRBD mount read-only: $DRBD_MOUNT_DIR"
        return 1
    fi
}

function wait_drbd_promoted() {
    log_message "Waiting for one of the DRBD nodes to be promoted."

    local pollInterval=2
    # Special parameter, no need to increase it ourselves.
    SECONDS=0

    while [[ $SECONDS -lt $DRBD_READY_TIMEOUT ]]; do
        if sudo crm resource status drbd_master_slave | grep Promoted ; then
            log_message "DRBD node promoted."
            return 0
        else
            log_message "No DRBD node promoted yet, retrying in ${pollInterval}s"
            sleep $pollInterval
        fi
    done

    log_message "Timed out waiting for primary DRBD node." \
                "Waited: ${SECONDS}. Timeout: ${DRBD_READY_TIMEOUT}s."
    return 1
}

function ensure_drbd_unmounted() {
    if mount | grep "on $DRBD_MOUNT_DIR type" &> /dev/null ; then
        log_message "DRBD device mounted: $DRBD_MOUNT_DIR"
        return 1
    fi
}

function ensure_drbd_ready() {
    ensure_mount_rw 

    diskStatus=`sudo drbdadm status r0 | grep disk | head -1 | cut -d ":" -f 2`
    if [[ $diskStatus != "UpToDate" ]]; then
        log_message "DRBD disk status not ready. Current status: $diskStatus"
        return 1
    else
        log_message "DRBD disk up to date."
    fi
}

function wait_drbd_primary () {
    log_message "Waiting for primary DRBD node to be ready."

    local pollInterval=2
    # Special parameter, no need to increase it ourselves.
    SECONDS=0

    while [[ $SECONDS -lt $DRBD_READY_TIMEOUT ]]; do
        if ensure_drbd_ready; then
            log_message "Primary DRBD node ready."
            return 0
        else
            log_message "Primary DRBD node not ready yet, retrying in ${pollInterval}s"
            sleep $pollInterval
        fi
    done

    log_message "Timed out waiting for primary DRBD node." \
                "Waited: ${SECONDS}. Timeout: ${DRBD_READY_TIMEOUT}s."
    return 1
}

function wait_for_peer_k8s() {
    local k8sDqliteNodeId=`get_dqlite_node_id $K8S_DQLITE_INFO_BKP_YAML`
    if [[ -z $k8sDqliteNodeId ]]; then
        log_message "Couldn't retrieve k8s-dqlite node id."
        exit 1
    fi

    local peerIp=`get_dql_peer_ip $K8S_DQLITE_CLUSTER_BKP_YAML $k8sDqliteNodeId`
    if [[ -z $peerIp ]]; then
        log_message "Couldn't retrieve dqlite peer ip."
        exit 1
    fi

    log_message "Waiting for k8s to start on peer: $peerIp. Timeout: ${PEER_READY_TIMEOUT}s."

    local pollInterval=2
    # Special parameter, no need to increase it ourselves.
    SECONDS=0

    while [[ $SECONDS -lt $PEER_READY_TIMEOUT ]]; do
        if ssh $SSH_OPTS $SSH_USERNAME@$peerIp sudo k8s status &> /dev/null; then
            log_message "Peer ready."
            return 0
        else
            log_message "Peer not ready yet, retrying in ${pollInterval}s."
            sleep $pollInterval
        fi
    done

    log_message "Timed out waiting for k8s services to start on peer." \
                "Waited: ${SECONDS}. Timeout: ${PEER_READY_TIMEOUT}s."
    return 1

}

# "drbdadm status" throws the following if our service starts before
# Pacemaker initialized DRBD (even on the secondary).
#
#  r0: No such resource
#  Command 'drbdsetup-84 status r0' terminated with exit code 10
function wait_drbd_resource () {
    log_message "Waiting for DRBD resource."

    local pollInterval=2
    # Special parameter, no need to increase it ourselves.
    SECONDS=0

    while [[ $SECONDS -lt $DRBD_READY_TIMEOUT ]]; do
        if sudo drbdadm status &> /dev/null; then
            log_message "DRBD ready."
            return 0
        else
            log_message "DRBD not ready yet, retrying in ${pollInterval}s" 
            sleep $pollInterval
        fi
    done

    log_message "Timed out waiting for DRBD resource." \
                "Waited: ${SECONDS}. Timeout: ${DRBD_READY_TIMEOUT}s."
    return 1
}

# Based on the drbd volume state, we decide if this node should be a
# dqlite voter or a spare.
function get_expected_dqlite_role() {
    drbdResRole=`sudo drbdadm status $DRBD_RES_NAME | head -1 | grep role | cut -d ":" -f 2`

    case $drbdResRole in
        "Primary")
            echo $DQLITE_ROLE_VOTER
            ;;
        "Secondary")
            echo $DQLITE_ROLE_SPARE
            ;;
        *)
            log_message "Unexpected DRBD role: $drbdResRole"
            exit 1
            ;;
    esac
}

function validate_drbd_state() {
    wait_drbd_promoted

    drbdResRole=`sudo drbdadm status $DRBD_RES_NAME | head -1 | grep role | cut -d ":" -f 2`

    case $drbdResRole in
        "Primary")
            wait_drbd_primary
            ;;
        "Secondary")
            ensure_drbd_unmounted
            ;;
        *)
            log_message "Unexpected DRBD role: $drbdResRole"
            exit 1
            ;;
    esac
}

# After a failover, the state dir points to the shared DRBD volume.
# We need to restore the node certificate and config files.
function restore_dqlite_confs_and_certs() {
    log_message "Restoring dqlite configs and certificates."

    sudo cp $K8S_DQLITE_STATE_BKP_DIR/info.yaml $K8S_DQLITE_STATE_DIR

    sudo cp $K8SD_STATE_BKP_DIR/database/info.yaml $K8SD_STATE_DIR/database/
    sudo cp $K8SD_STATE_BKP_DIR/daemon.yaml $K8SD_STATE_DIR/

    # restore k8s-dqlite certificates
    sudo cp $K8S_DQLITE_STATE_BKP_DIR/cluster.crt $K8S_DQLITE_STATE_DIR
    sudo cp $K8S_DQLITE_STATE_BKP_DIR/cluster.key $K8S_DQLITE_STATE_DIR

    # restore k8sd certificates
    sudo cp $K8SD_STATE_BKP_DIR/cluster.crt $K8SD_STATE_DIR
    sudo cp $K8SD_STATE_BKP_DIR/cluster.key $K8SD_STATE_DIR
    sudo cp $K8SD_STATE_BKP_DIR/server.crt $K8SD_STATE_DIR
    sudo cp $K8SD_STATE_BKP_DIR/server.key $K8SD_STATE_DIR
}

# Promote the current node as primary and prepare the recovery archives.
function promote_as_primary() {
    local k8sDqliteNodeId=`get_dqlite_node_id $K8S_DQLITE_INFO_BKP_YAML`
    if [[ -z $k8sDqliteNodeId ]]; then
        log_message "Couldn't retrieve k8s-dqlite node id."
        exit 1
    fi

    local k8sdNodeId=`get_dqlite_node_id $K8SD_INFO_BKP_YAML`
    if [[ -z $k8sDqliteNodeId ]]; then
        log_message "Couldn't retrieve k8s-dqlite node id."
        exit 1
    fi

    local peerIp=`get_dql_peer_ip $K8S_DQLITE_CLUSTER_YAML $k8sDqliteNodeId`
    if [[ -z $peerIp ]]; then
        log_message "Couldn't retrieve dqlite peer ip."
        exit 1
    fi

    log_message "Stopping local k8s services."
    sudo snap stop k8s

    # After a node crash, there may be a leaked control socket file and
    # k8sd will refuse to perform the recovery. We've just stopped the k8s snap,
    # it should be safe to remove such stale unix sockets.
    log_message "Removing stale control sockets."
    sudo rm -f $K8SD_STATE_DIR/control.socket

    local stoppedPeer=0
    log_message "Checking peer k8s services: $peerIp"
    if ssh $SSH_OPTS $SSH_USERNAME@$peerIp sudo snap services k8s | grep -v inactive | grep "active"; then
        log_message "Attempting to stop peer k8s services."
        # Stop the k8s snap directly instead of the wrapper service so that
        # we won't cause failures if both nodes start at the same time.
        # The secondary will wait for the k8s services to start on the primary.
        if ssh $SSH_OPTS $SSH_USERNAME@$peerIp sudo snap stop k8s; then
            stoppedPeer=1
            log_message "Successfully stopped peer k8s services."
            log_message "The stopped services are going to be restarted after the recovery finishes."
        else
            log_message "Couldn't stop k8s services on the peer node." \
                        "Assuming that it's stopped and proceeding with the recovery."
        fi
    fi

    log_message "Ensuring rw access to DRBD mount."
    # Having RW access to the drbd mount implies that this is the primary node.
    ensure_mount_rw

    restore_dqlite_confs_and_certs

    log_message "Updating dqlite roles."
    # Update info.yaml
    set_dqlite_node_role $K8S_DQLITE_INFO_YAML $DQLITE_ROLE_VOTER
    set_dqlite_node_role $K8SD_INFO_YAML $DQLITE_ROLE_VOTER

    # Update cluster.yaml
    set_dqlite_node_as_sole_voter $K8S_DQLITE_CLUSTER_YAML $k8sDqliteNodeId
    set_dqlite_node_as_sole_voter $K8SD_CLUSTER_YAML $k8sdNodeId

    log_message "Restoring dqlite."
    sudo $K8SD_PATH cluster-recover \
        --state-dir=$K8SD_STATE_DIR \
        --k8s-dqlite-state-dir=$K8S_DQLITE_STATE_DIR \
        --log-level $K8SD_LOG_LEVEL \
        --non-interactive

    # TODO: consider removing offending segments if the last snapshot is behind
    # and then try again.

    log_message "Copying k8sd recovery tarball to $K8SD_RECOVERY_TARBALL_BKP"
    sudo cp $K8SD_RECOVERY_TARBALL $K8SD_RECOVERY_TARBALL_BKP

    log_message "Restarting k8s services."
    sudo snap start k8s

    # TODO: validate k8s status

    if [[ $stoppedPeer -ne 0 ]]; then
        log_message "Restarting peer k8s services: $peerIp"
        # It's importand to issue a restart here since we stopped the k8s snap
        # directly and the wrapper service doesn't currently monitor it.
        ssh $SSH_OPTS $SSH_USERNAME@$peerIp sudo systemctl restart $SYSTEMD_SERVICE_NAME ||
            log_message "Couldn't start peer k8s services."
    fi
}

function process_recovery_files_on_secondary() {
    local peerIp="$1"

    log_message "Ensuring that the drbd volume is unmounted."
    ensure_drbd_unmounted

    log_message "Restoring local dqlite backup files."
    sudo cp -r $K8S_DQLITE_STATE_BKP_DIR/. $DRBD_MOUNT_DIR/k8s-dqlite/
    sudo cp -r $K8SD_STATE_BKP_DIR/. $DRBD_MOUNT_DIR/k8sd/

    sudo rm -f $DRBD_MOUNT_DIR/k8s-dqlite/00*-*
    sudo rm -f $DRBD_MOUNT_DIR/k8s-dqlite/snapshot-*
    sudo rm -f $DRBD_MOUNT_DIR/k8s-dqlite/metadata*

    sudo rm -f $DRBD_MOUNT_DIR/k8sd/database/00*-*
    sudo rm -f $DRBD_MOUNT_DIR/k8sd/database/snapshot-*
    sudo rm -f $DRBD_MOUNT_DIR/k8sd/database/metadata*

    log_message "Retrieving k8sd recovery tarball."
    scp $SSH_OPTS $SSH_USERNAME@$peerIp:$K8SD_RECOVERY_TARBALL_BKP /tmp/
    sudo mv /tmp/`basename $K8SD_RECOVERY_TARBALL_BKP` \
        $K8SD_RECOVERY_TARBALL

    # TODO: do we really need to transfer recovery tarballs in this situation?
    # the spare is simply forwarding the requests to the primary, it doesn't really
    # hold any data.
    lastK8sDqliteRecoveryTarball=`ssh $SSH_USERNAME@$peerIp \
        sudo ls /var/snap/k8s/common/ | \
            grep -P "recovery-k8s-dqlite-.*post-recovery" | \
            tail -1`
    if [ -z "$lastK8sDqliteRecoveryTarball" ]; then
        log_message "couldn't retrieve latest k8s-dqlite recovery tarball from $peerIp"
        exit 1
    fi

    log_message "Retrieving k8s-dqlite recovery tarball."
    scp $SSH_USERNAME@$peerIp:/var/snap/k8s/common/$lastK8sDqliteRecoveryTarball /tmp/
    sudo tar -xf /tmp/$lastK8sDqliteRecoveryTarball -C $K8S_DQLITE_STATE_DIR

    log_message "Updating dqlite roles."
    # Update info.yaml
    set_dqlite_node_role $K8S_DQLITE_INFO_YAML $DQLITE_ROLE_SPARE
    set_dqlite_node_role $K8SD_INFO_YAML $DQLITE_ROLE_SPARE
    # We're skipping cluster.yaml, we expect the recovery archives to contain
    # updated cluster.yaml files.
}

# Recover a former primary, now secondary dqlite node.
# Run "promote_as_primary" on the ther node first.
function rejoin_secondary() {
    log_message "Recovering secondary node."

    local k8sDqliteNodeId=`get_dqlite_node_id $K8S_DQLITE_INFO_BKP_YAML`
    if [[ -z $k8sDqliteNodeId ]]; then
        log_message "Couldn't retrieve k8s-dqlite node id."
        exit 1
    fi

    local peerIp=`get_dql_peer_ip $K8S_DQLITE_CLUSTER_BKP_YAML $k8sDqliteNodeId`
    if [[ -z $peerIp ]]; then
        log_message "Couldn't retrieve dqlite peer ip."
        exit 1
    fi

    log_message "Stopping k8s services."
    sudo snap stop k8s

    log_message "Adding temporary Pacemaker constraint."
    # We need to prevent failovers from happening while restoring secondary
    # dqlite data, otherwise we may end up overriding or deleting the primary
    # node data.
    #
    # TODO: consider reducing the constraint scope (e.g. resource level constraint
    # instead of putting the entire node in standby).
    sudo crm node standby
    if ! process_recovery_files_on_secondary $peerIp; then
        log_message "Dqlite recovery filed, removing temporary Pacemaker constraints."
        sudo crm node online
        exit 1
    fi

    log_message "Restoring Pacemaker state."
    sudo crm node online

    log_message "Restarting k8s services"
    sudo snap start k8s
}

function install_packages() {
    sudo apt-get update

    sudo DEBIAN_FRONTEND=noninteractive apt-get install \
      python3 python3-netaddr \
      pacemaker resource-agents-extra \
      drbd-utils ntp linux-image-generic snap moreutils -y
    sudo modprobe drbd || sudo apt-get install -y linux-modules-extra-$(uname -r)

    sudo snap install jq
    sudo snap install yq
    sudo snap install install k8s --classic $K8S_SNAP_CHANNEL
}

function check_statedir() {
    local stateDir="$1"
    local expLink="$2"

    if [[ ! -e $stateDir ]]; then
        log_message "State directory missing: $stateDir"
        exit 1
    fi

    target=`readlink -f $stateDir`
    if [[ -L "$stateDir" ]] && [[ "$target" != "$expLink" ]]; then
        log_message "Unexpected symlink target. " \
                    "State directory: $stateDir. " \
                    "Expected symlink target: $expLink. " \
                    "Actual symlink target: $target."
        exit 1
    fi

    if [[ ! -L $stateDir ]] &&  [[ ! -z "$( ls -A $expLink )" ]]; then
        log_message "State directory is not a symlink, however the " \
                    "expected link target exists and is not empty. " \
                    "We can't know which files to keep, erroring out. " \
                    "State directory: $stateDir. " \
                    "Expected symlink target: $expLink."
        exit 1
    fi
}

function check_peer_recovery_tarballs() {
    log_message "Retrieving k8s-dqlite node id."
    local k8sDqliteNodeId=`get_dqlite_node_id $K8S_DQLITE_INFO_BKP_YAML`
    if [[ -z $k8sDqliteNodeId ]]; then
        log_message "Couldn't retrieve k8s-dqlite node id."
        exit 1
    fi

    log_message "Retrieving dqlite peer ip."
    local peerIp=`get_dql_peer_ip $K8S_DQLITE_CLUSTER_BKP_YAML $k8sDqliteNodeId`
    if [[ -z $peerIp ]]; then
        log_message "Couldn't retrieve dqlite peer ip."
        exit 1
    fi

    log_message "Checking for recovery taballs on $peerIp."

    k8sdRecoveryTarball=`ssh $SSH_OPTS $SSH_USERNAME@$peerIp \
        sudo ls -A "$K8SD_RECOVERY_TARBALL_BKP"`
    if [[ -z $k8sdRecoveryTarball ]]; then
        log_message "Peer $peerIp doesn't have k8sd recovery tarball."
        return 1
    fi

    lastK8sDqliteRecoveryTarball=`ssh $SSH_OPTS $SSH_USERNAME@$peerIp \
        sudo ls /var/snap/k8s/common/ | \
            grep -P "recovery-k8s-dqlite-.*post-recovery"`
    if [[ -z $k8sdRecoveryTarball ]]; then
        log_message "Peer $peerIp doesn't have k8s-dqlite recovery tarball."
        return 1
    fi
}

function start_service() {
    log_message "Initializing node."

    # DRBD is the primary source of truth for the dqlite role.
    # We need to wait for it to become available.
    wait_drbd_resource

    # dump the drbd and pacemaker status for debugging purposes.
    sudo drbdadm status
    sudo crm status

    validate_drbd_state

    move_statedirs

    local expRole=`get_expected_dqlite_role`
    case $expRole in
        $DQLITE_ROLE_VOTER)
            log_message "Assuming the dqlite voter role (primary)."

            # We'll assume that if the primary stopped, it needs to go through
            # the recovery process.
            promote_as_primary
            ;;
        $DQLITE_ROLE_SPARE)
            log_message "Assuming the dqlite spare role (secondary)."

            wait_for_peer_k8s

            if check_peer_recovery_tarballs; then
                log_message "Recovery tarballs found, initiating recovery."
                rejoin_secondary
            else
                # Maybe the primary didn't change and we don't need to go
                # through the recovery process.
                # TODO: consider comparing the cluster.yaml files from the
                # two nodes.
                log_message "Recovery tarballs missing, skipping recovery."
                log_message "Starting k8s services."
                sudo snap k8s start
            fi
            ;;
        *)
            log_message "Unexpected dqlite role: $expRole"
            exit 1
            ;;
    esac
}

function clean_recovery_data() {
    log_message "Cleaning up dqlite recovery data."
    rm -f $K8SD_RECOVERY_TARBALL
    rm -f $K8SD_RECOVERY_TARBALL_BKP
    rm -f $K8S_DQLITE_STATE_DIR/recovery-k8s-dqlite*
}

function purge() {
    log_message "Removing the k8s snap and all the associated files."

    sudo snap remove --purge k8s

    if [[ -d $DRBD_MOUNT_DIR ]]; then
        log_message "Cleaning up $DRBD_MOUNT_DIR."
        sudo rm -rf $DRBD_MOUNT_DIR/k8sd
        sudo rm -rf $DRBD_MOUNT_DIR/k8s-dqlite

        if ! ensure_drbd_unmounted; then
            log_message "Cleaning up $DRBD_MOUNT_DIR mount point."

            # The replicas use the mount dir directly, without a block device
            # attachment. We need to clean up the mount point as well.
            #
            # We're using another mount with "--bind" to bypass the drbd mount.
            tempdir=`mktemp -d`
            # We need to mount the parent dir.
            sudo mount --bind `dirname $DRBD_MOUNT_DIR` $tempdir
            sudo rm -rf $tempdir/`basename $DRBD_MOUNT_DIR`/k8sd
            sudo rm -rf $tempdir/`basename $DRBD_MOUNT_DIR`/k8s-dqlite
            sudo umount $tempdir
            sudo rm -rf $tempdir
        fi
    fi
}

function clear_taints() {
    log_message "Clearing tainted Pacemaker resources."
    sudo crm resource clear ha_k8s_failover_service
    sudo crm resource clear fs_res
    sudo crm resource clear drbd_master_slave

    sudo crm resource cleanup ha_k8s_failover_service
    sudo crm resource cleanup fs_res
    sudo crm resource cleanup drbd_master_slave
}

function main() {
    local command=$1

    case $command in
        "move_statedirs")
            move_statedirs
            ;;
        "install_packages")
            install_packages
            ;;
        "start_service")
            start_service
            ;;
        "clean_recovery_data")
            clean_recovery_data
            ;;
        "purge")
            purge
            ;;
        "clear_taints")
            clear_taints
            ;;
        *)
            cat << EOF
Unknown command: $1

usage: $0 <command>

Commands:
    move_statedirs          Move the dqlite state directories to the DRBD mount,
                            replacing them with symlinks.
                            The existing contents are moved to a backup folder,
                            which can be used as part of the recovery process.
    install_packages        Install the packages required by the 2-node HA
                            cluster.
    start_service           Initialize the k8s services, taking the following
                            steps:
                            1. Based on the drbd state, decide if this node
                               should assume the primary (dqlite voter) or
                               secondary (spare) role.
                            2. If this is the first start, transfer the dqlite
                               state directories and create backups.
                            3. If this node is a primary, promote it and initiate
                               the dqlite recovery, creating recovery tarballs.
                               Otherwise, copy over the recovery files and
                               join the existing cluster as a spare.
                            4. Start the k8s services.
                            IMPORTANT: ensure that the DRBD volume is attached
                            to the primary node when running the command for
                            the first time.
    clean_recovery_data     Remove database recovery files. Should be called
                            after the cluster has been fully recovered.
    purge                   Remove the k8s snap and all its associated files.
    clear_taints            Clear tainted Pacemaker resources.

EOF
            ;;
    esac
}

if [[ $sourced -ne 1 ]]; then
    main $@
fi
