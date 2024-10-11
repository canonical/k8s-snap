package k8sd

import (
	"bufio"
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"os"
	"path"
	"regexp"
	"slices"
	"strings"
	"time"

	"github.com/canonical/go-dqlite"
	"github.com/canonical/go-dqlite/app"
	"github.com/canonical/go-dqlite/client"
	"github.com/canonical/lxd/shared"
	"github.com/canonical/lxd/shared/termios"
	"github.com/canonical/microcluster/v3/cluster"
	"github.com/canonical/microcluster/v3/microcluster"
	"github.com/spf13/cobra"
	"golang.org/x/sys/unix"
	"gopkg.in/yaml.v2"

	cmdutil "github.com/canonical/k8s/cmd/util"
	"github.com/canonical/k8s/pkg/log"
	"github.com/canonical/k8s/pkg/utils"
)

const preRecoveryMessage = `You should only run this command if:
 - A quorum of cluster members is permanently lost
 - You are *absolutely* sure all k8s daemons are stopped (sudo snap stop k8s)
 - This instance has the most up to date database

Note that before applying any changes, a database backup is created at:
* k8sd (microcluster): /var/snap/k8s/common/var/lib/k8sd/state/db_backup.<timestamp>.tar.gz
* k8s-dqlite: /var/snap/k8s/common/recovery-k8s-dqlite-<timestamp>-pre-recovery.tar.gz
`

const recoveryConfirmation = "Do you want to proceed? (yes/no): "

const nonInteractiveMessage = `Non-interactive mode requested.

The command will assume that the dqlite configuration files have already been
modified with the updated cluster member roles and addresses.

Initiating the dqlite database recovery.
`

const clusterK8sdYamlRecoveryComment = `# Member roles can be modified. Unrecoverable nodes should be given the role "spare".
#
# "voter" (0) - Voting member of the database. A majority of voters is a quorum.
# "stand-by" (1) - Non-voting member of the database; can be promoted to voter.
# "spare" (2) - Not a member of the database.
#
# The edit is aborted if:
# - the number of members changes
# - the name of any member changes
# - the ID of any member changes
# - the address of any member changes
# - no changes are made
`

const clusterK8sDqliteRecoveryComment = `# Member roles can be modified. Unrecoverable nodes should be given the role 2 (spare).
#
# 0 (voter) - Voting member of the database. A majority of voters is a quorum.
# 1 (stand-by) - Non-voting member of the database; can be promoted to voter.
# 2 (spare) - Not a member of the database.
`

const infoYamlRecoveryComment = `# Verify the ID, address and role of the local node.
#
# Cluster members:
`

const daemonYamlRecoveryComment = `# Verify the name and address of the local node.
#
# Cluster members:
`

// Used as part of a regex search, avoid adding special characters.
const yamlHelperCommentFooter = "# ------- everything below will be written -------\n"

var clusterRecoverOpts struct {
	K8sDqliteStateDir string
	NonInteractive    bool
	SkipK8sd          bool
	SkipK8sDqlite     bool
}

func logDebugf(format string, args ...interface{}) {
	// TODO: there may be a problem with the logger, only log.V(0) messages
	// get printed regardless of the specified log level. For now, we'll use our
	// own helper.
	if rootCmdOpts.logDebug {
		msg := fmt.Sprintf(format, args...)
		log.L().Info(msg)
	}

}

func newClusterRecoverCmd(env cmdutil.ExecutionEnvironment) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "cluster-recover",
		Short: "Recover the cluster from this member if quorum is lost",
		Run: func(cmd *cobra.Command, args []string) {
			log.Configure(log.Options{
				LogLevel:     rootCmdOpts.logLevel,
				AddDirHeader: true,
			})

			if err := recoveryCmdPrechecks(cmd); err != nil {
				cmd.PrintErrf("Recovery precheck failed: %v\n", err)
				env.Exit(1)
			}

			if clusterRecoverOpts.SkipK8sd {
				cmd.Printf("Skipping k8sd recovery.\n")
			} else {
				k8sdTarballPath, err := recoverK8sd()
				if err != nil {
					cmd.PrintErrf("Failed to recover k8sd, error: %v\n", err)
					env.Exit(1)
				}
				cmd.Printf("K8sd cluster changes applied.\n")
				cmd.Printf("New database state saved to %s\n", k8sdTarballPath)
				cmd.Printf("*Before* starting any cluster member, copy %s to %s "+
					"on all remaining cluster members.\n",
					k8sdTarballPath, k8sdTarballPath)
				cmd.Printf("K8sd will load this file during startup.\n\n")
			}

			if clusterRecoverOpts.SkipK8sDqlite {
				cmd.Printf("Skipping k8s-dqlite recovery.\n")
			} else {
				k8sDqlitePreRecoveryTarball, k8sDqlitePostRecoveryTarball, err := recoverK8sDqlite()
				if err != nil {
					cmd.PrintErrf(
						"Failed to recover k8s-dqlite, error: %v, "+
							"pre-recovery backup: %s\n",
						err, k8sDqlitePreRecoveryTarball)
					env.Exit(1)
				}
				cmd.Printf("K8s-dqlite cluster changes applied.\n")
				cmd.Printf("New database state saved to %s\n",
					k8sDqlitePostRecoveryTarball)
				cmd.Printf("*Before* starting any cluster member, copy %s "+
					"on all remaining cluster members and extract the archive to %s.\n",
					k8sDqlitePostRecoveryTarball, clusterRecoverOpts.K8sDqliteStateDir)
				cmd.Printf("Pre-recovery database backup: %s\n\n", k8sDqlitePreRecoveryTarball)
			}
		},
	}

	cmd.Flags().StringVar(&clusterRecoverOpts.K8sDqliteStateDir, "k8s-dqlite-state-dir",
		"", "k8s-dqlite datastore location")
	cmd.Flags().BoolVar(&clusterRecoverOpts.NonInteractive, "non-interactive",
		false, "disable interactive prompts, assume that the configs have been updated")
	cmd.Flags().BoolVar(&clusterRecoverOpts.SkipK8sd, "skip-k8sd",
		false, "skip k8sd recovery")
	cmd.Flags().BoolVar(&clusterRecoverOpts.SkipK8sDqlite, "skip-k8s-dqlite",
		false, "skip k8s-dqlite recovery")

	return cmd
}

func removeYamlHelperComments(content []byte) []byte {
	pattern := fmt.Sprintf("(?s).*?%s *", yamlHelperCommentFooter)
	re := regexp.MustCompile(pattern)
	out := re.ReplaceAll(content, nil)
	return out
}

func removeEmptyLines(content []byte) []byte {
	re := regexp.MustCompile(`(?m)^\s*$`)
	out := re.ReplaceAll(content, nil)
	return out
}

func recoveryCmdPrechecks(cmd *cobra.Command) error {
	log := log.FromContext(cmd.Context())

	log.V(1).Info("Running prechecks.")

	if !termios.IsTerminal(unix.Stdin) && !clusterRecoverOpts.NonInteractive {
		return fmt.Errorf("interactive mode requested in a non-interactive terminal")
	}

	if clusterRecoverOpts.K8sDqliteStateDir == "" {
		return fmt.Errorf("k8s-dqlite state dir not specified")
	}
	if rootCmdOpts.stateDir == "" {
		return fmt.Errorf("k8sd state dir not specified")
	}

	cmd.Print(preRecoveryMessage)
	cmd.Print("\n")

	if clusterRecoverOpts.NonInteractive {
		cmd.Print(nonInteractiveMessage)
		cmd.Print("\n")
	} else {
		reader := bufio.NewReader(os.Stdin)
		cmd.Print(recoveryConfirmation)

		input, err := reader.ReadString('\n')
		if err != nil {
			return fmt.Errorf("couldn't read user input, error: %w", err)
		}
		input = strings.TrimSuffix(input, "\n")

		if strings.ToLower(input) != "yes" {
			return fmt.Errorf("cluster edit aborted; no changes made")
		}

		cmd.Print("\n")
	}

	if !clusterRecoverOpts.SkipK8sDqlite {
		if err := ensureK8sDqliteMembersStopped(cmd.Context()); err != nil {
			return err
		}
	}

	return nil
}

func ensureK8sDqliteMembersStopped(ctx context.Context) error {
	log := log.FromContext(ctx)

	log.V(1).Info("Ensuring that all k8s-dqlite members are stopped.")

	clusterYamlPath := path.Join(clusterRecoverOpts.K8sDqliteStateDir, "cluster.yaml")
	membersYaml, err := os.ReadFile(clusterYamlPath)
	if err != nil {
		return fmt.Errorf("could not read k8s-dqlite cluster.yaml, error: %w", err)
	}

	members := []dqlite.NodeInfo{}
	if err = yaml.Unmarshal(membersYaml, &members); err != nil {
		return fmt.Errorf("couldn't parse k8s-dqlite cluster.yaml, error: %w", err)
	}

	crt := path.Join(clusterRecoverOpts.K8sDqliteStateDir, "cluster.crt")
	key := path.Join(clusterRecoverOpts.K8sDqliteStateDir, "cluster.key")

	keypair, err := tls.LoadX509KeyPair(crt, key)
	if err != nil {
		return fmt.Errorf("could not load k8s-dqlite certificates, error: %w", err)
	}

	data, err := os.ReadFile(crt)
	if err != nil {
		return fmt.Errorf("could not read k8s-dqlite certificate, error: %w", err)
	}

	pool := x509.NewCertPool()
	if !pool.AppendCertsFromPEM(data) {
		return fmt.Errorf("invalid k8s-dqlite certificate")
	}

	dial := client.DialFuncWithTLS(client.DefaultDialFunc, app.SimpleDialTLSConfig(keypair, pool))

	// We'll dial all members concurrently, passing the reachable addresses
	// through a channel.
	availableMembers := []string{}
	c := make(chan string)
	for _, member := range members {
		log.V(1).Info("Checking k8s-dqlite member", "member", member)

		if member.Address == "" {
			return fmt.Errorf("k8s-dqlite member check failed, missing member address.")
		}

		go func(ctx context.Context, dialFunc client.DialFunc, addr string) {
			// We expect the member nodes to be unavailable, let's not wait
			// more than 3 seconds during this sanity test.
			ctx_timeout, cancel := context.WithTimeout(ctx, 3*time.Second)
			defer cancel()

			log.V(1).Info("testing k8s-dqlite member connectivity", "addr", addr)
			conn, err := dialFunc(ctx_timeout, addr)
			if err == nil {
				log.V(1).Info("k8s-dqlite member reachable", "addr", addr)
				conn.Close()
				// The cluster member is reachable, pass back the address.
				c <- addr
			} else {
				log.V(1).Info("k8s-dqlite member unreachable",
					"addr", addr, "error", err)
				c <- ""
			}
		}(ctx, dial, member.Address)
	}

	for _, _ = range members {
		addr, ok := <-c
		if !ok {
			return fmt.Errorf("channel closed unexpectedly")
		}
		if addr != "" {
			availableMembers = append(availableMembers, addr)
		}
	}

	if len(availableMembers) != 0 {
		return fmt.Errorf("Some k8s-dqlite services are still running, "+
			"please stop them and try again: %v. "+
			"Run `sudo snap stop k8s` to stop all k8s services on a given node.",
			availableMembers)
	}

	return nil
}

// yamlEditorGuide is a convenience wrapper around shared.TextEditor
// that passes the current file contents prepended by the guide contents,
// which are meant to assist the user. Returns the user-edited file contents.
// If applyChanges is set, the changes made by the user are applied to the file.
func yamlEditorGuide(
	path string,
	readFile bool,
	guideContent []byte,
	applyChanges bool,
) ([]byte, error) {
	currContent := []byte{}
	var err error
	if readFile {
		currContent, err = os.ReadFile(path)
		if err != nil {
			return nil, fmt.Errorf("could not read file: %s, error: %w", path, err)
		}
		currContent = removeEmptyLines(currContent)
	}

	textEditorContent := slices.Concat(
		[]byte(guideContent),
		[]byte("\n"),
		currContent)
	newContent, err := shared.TextEditor("", textEditorContent)
	if err != nil {
		return nil, fmt.Errorf("text editor failed, error: %w", err)
	}

	newContent = removeYamlHelperComments(newContent)
	newContent = removeEmptyLines(newContent)

	if applyChanges {
		err = os.WriteFile(path, newContent, os.FileMode(0o644))
		if err != nil {
			return nil, fmt.Errorf("could not write file: %s, error: %w", path, err)
		}
	}

	return newContent, err
}

func createK8sDqliteRecoveryTarball(pathSuffix string) (string, error) {
	timestamp := time.Now().Format("2006-01-02T150405Z0700")
	fname := fmt.Sprintf("recovery-k8s-dqlite-%s-%s.tar.gz", timestamp, pathSuffix)
	tarballPath := path.Join("/var/snap/k8s/common", fname)

	// info.yaml is used by go-dqlite to keep track of the current cluster member's
	// ID and address. We shouldn't replicate the recovery member's info.yaml
	// to all other members, so exclude it from the tarball:
	err := utils.CreateTarball(
		tarballPath, clusterRecoverOpts.K8sDqliteStateDir, ".",
		[]string{"info.yaml", "k8s-dqlite.sock", "cluster.key", "cluster.crt"})

	return tarballPath, err
}

// On success, returns the recovery tarball path.
func recoverK8sd() (string, error) {
	m, err := microcluster.App(
		microcluster.Args{
			StateDir: rootCmdOpts.stateDir,
		},
	)
	if err != nil {
		return "", fmt.Errorf("could not initialize microcluster app, error: %w", err)
	}

	// The following method parses cluster.yaml and filters out the entries
	// that are not included in the trust store. Note that in case of k8s-dqlite,
	// there is no trust store.
	members, err := m.GetDqliteClusterMembers()
	if err != nil {
		return "", fmt.Errorf("could not retrieve K8sd cluster members, error: %w", err)
	}

	oldMembersYaml, err := yaml.Marshal(members)
	if err != nil {
		return "", fmt.Errorf("could not serialize cluster members, error: %w", err)
	}

	clusterYamlPath := path.Join(m.FileSystem.DatabaseDir, "cluster.yaml")
	clusterYamlCommentHeader := fmt.Sprintf("# K8sd cluster configuration\n# (based on the trust store and %s)\n", clusterYamlPath)

	clusterYamlContent := oldMembersYaml
	if !clusterRecoverOpts.NonInteractive {
		// Interactive mode requested (default).
		// Assist the user in configuring dqlite.
		clusterYamlContent, err = yamlEditorGuide(
			"",
			false,
			slices.Concat(
				[]byte(clusterYamlCommentHeader),
				[]byte("#\n"),
				[]byte(clusterK8sdYamlRecoveryComment),
				[]byte(yamlHelperCommentFooter),
				[]byte("\n"),
				oldMembersYaml,
			),
			false,
		)
		if err != nil {
			return "", fmt.Errorf("interactive text editor failed, error: %w", err)
		}

		infoYamlPath := path.Join(m.FileSystem.DatabaseDir, "info.yaml")
		infoYamlCommentHeader := fmt.Sprintf("# K8sd info.yaml\n# (%s)\n", infoYamlPath)
		_, err = yamlEditorGuide(
			infoYamlPath,
			true,
			slices.Concat(
				[]byte(infoYamlCommentHeader),
				[]byte("#\n"),
				[]byte(infoYamlRecoveryComment),
				utils.YamlCommentLines(clusterYamlContent),
				[]byte("\n"),
				[]byte(yamlHelperCommentFooter),
			),
			true,
		)
		if err != nil {
			return "", fmt.Errorf("interactive text editor failed, error: %w", err)
		}

		daemonYamlPath := path.Join(m.FileSystem.StateDir, "daemon.yaml")
		daemonYamlCommentHeader := fmt.Sprintf("# K8sd daemon.yaml\n# (%s)\n", daemonYamlPath)
		_, err = yamlEditorGuide(
			daemonYamlPath,
			true,
			slices.Concat(
				[]byte(daemonYamlCommentHeader),
				[]byte("#\n"),
				[]byte(daemonYamlRecoveryComment),
				utils.YamlCommentLines(clusterYamlContent),
				[]byte("\n"),
				[]byte(yamlHelperCommentFooter),
			),
			true,
		)
		if err != nil {
			return "", fmt.Errorf("interactive text editor failed, error: %w", err)
		}
	}

	newMembers := []cluster.DqliteMember{}
	if err = yaml.Unmarshal(clusterYamlContent, &newMembers); err != nil {
		return "", fmt.Errorf("couldn't parse cluster.yaml, error: %w", err)
	}

	// As of 2.0.2, the following microcluster method will:
	// * validate the member changes
	//     * ensure that no members were added or removed
	//     * the member IDs hasn't changed
	//     * there is at least one voter
	//     * the addresses can be parsed
	//     * there are no duplicate addresses
	// * ensure that all cluster members are stopped
	// * create a database backup
	// * reconfigure Raft
	// * address changes based on the new cluster.yaml:
	//     * refresh the local info.yaml and daemon.yaml
	//     * update the trust store addresses
	//     * prepare an sql script that updates the member addresses from the
	//       "core_cluster_members" table, executed when k8sd starts.
	// * rewrite cluster.yaml
	// * create a recovery tarball of the k8sd database dir and store it
	//   in the state dir.
	tarballPath, err := m.RecoverFromQuorumLoss(newMembers)
	if err != nil {
		return "", fmt.Errorf("k8sd recovery failed, error: %w", err)
	}

	return tarballPath, nil
}

func recoverK8sDqlite() (string, string, error) {
	k8sDqliteStateDir := clusterRecoverOpts.K8sDqliteStateDir

	var err error
	clusterYamlContent := []byte{}
	clusterYamlPath := path.Join(k8sDqliteStateDir, "cluster.yaml")
	clusterYamlCommentHeader := fmt.Sprintf("# k8s-dqlite cluster configuration\n# (%s)\n", clusterYamlPath)

	if clusterRecoverOpts.NonInteractive {
		clusterYamlContent, err = os.ReadFile(clusterYamlPath)
		if err != nil {
			return "", "", fmt.Errorf(
				"could not read k8s-dqlite cluster.yaml, error: %w", err)
		}
	} else {
		// Interactive mode requested (default).
		// Assist the user in configuring dqlite.
		clusterYamlContent, err = yamlEditorGuide(
			clusterYamlPath,
			true,
			slices.Concat(
				[]byte(clusterYamlCommentHeader),
				[]byte("#\n"),
				[]byte(clusterK8sDqliteRecoveryComment),
				[]byte(yamlHelperCommentFooter),
			),
			true,
		)
		if err != nil {
			return "", "", fmt.Errorf("interactive text editor failed, error: %w", err)
		}

		infoYamlPath := path.Join(k8sDqliteStateDir, "info.yaml")
		infoYamlCommentHeader := fmt.Sprintf("# k8s-dqlite info.yaml\n# (%s)\n", infoYamlPath)
		_, err = yamlEditorGuide(
			infoYamlPath,
			true,
			slices.Concat(
				[]byte(infoYamlCommentHeader),
				[]byte("#\n"),
				[]byte(infoYamlRecoveryComment),
				utils.YamlCommentLines(clusterYamlContent),
				[]byte("\n"),
				[]byte(yamlHelperCommentFooter),
			),
			true,
		)
		if err != nil {
			return "", "", fmt.Errorf("interactive text editor failed, error: %w", err)
		}
	}

	newMembers := []dqlite.NodeInfo{}
	if err = yaml.Unmarshal(clusterYamlContent, &newMembers); err != nil {
		return "", "", fmt.Errorf("couldn't parse cluster.yaml, error: %w", err)
	}

	// microcluster creates two backup tarballs, one before the recovery was attempted
	// and another one after. We'll do the same.
	preRecoveryTarball, err := createK8sDqliteRecoveryTarball("pre-recovery")
	if err != nil {
		return "", "", fmt.Errorf("failed to create pre-recovery backup tarball, error: %w", err)
	}

	if err = dqlite.ReconfigureMembershipExt(k8sDqliteStateDir, newMembers); err != nil {
		return preRecoveryTarball, "", fmt.Errorf("k8s-dqlite recovery failed, error: %w", err)
	}

	postRecoveryTarball, err := createK8sDqliteRecoveryTarball("post-recovery")
	if err != nil {
		return preRecoveryTarball, "", fmt.Errorf("failed to create post-recovery backup tarball, error: %w", err)
	}

	// TODO: steps performed by the microcluster lib for k8sd that could also
	// apply to k8s-dqlite:
	//   * validate cluster.yaml changes
	return preRecoveryTarball, postRecoveryTarball, nil
}
