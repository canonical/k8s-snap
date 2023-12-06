package database

import (
	"context"
	"database/sql"
	_ "embed"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/canonical/lxd/lxd/db/cluster"
)

type KubernetesIdentity struct {
	Username string   `json:"username"`
	Groups   []string `json:"groups"`
}

var (
	//go:embed sql/queries/k8sd-tokens-select-auth-by-token.sql
	k8sdTokensSelectAuthByTokenSQL  string
	k8sdTokensSelectAuthByTokenStmt = cluster.RegisterStmt(k8sdTokensSelectAuthByTokenSQL)

	//go:embed sql/queries/k8sd-tokens-select-token-by-auth.sql
	k8sdTokensSelectTokenByAuthSQL  string
	k8sdTokensSelectTokenByAuthStmt = cluster.RegisterStmt(k8sdTokensSelectTokenByAuthSQL)

	//go:embed sql/queries/k8sd-tokens-insert-token.sql
	k8sdTokensInsertTokenSQL  string
	k8sdTokensInsertTokenStmt = cluster.RegisterStmt(k8sdTokensInsertTokenSQL)
)

func groupsToString(inGroups []string) (string, error) {
	groupMap := make(map[string]struct{}, len(inGroups))
	groups := make([]string, 0, len(inGroups))
	for _, group := range inGroups {
		if group == "" {
			return "", fmt.Errorf("group cannot be empty")
		}
		if _, duplicate := groupMap[group]; duplicate {
			return "", fmt.Errorf("duplicate group %s", group)
		}
		groupMap[group] = struct{}{}
		groups = append(groups, group)
	}
	sort.Strings(groups)
	return strings.Join(groups, ","), nil
}

func groupsToList(inGroups string) []string {
	if inGroups == "" {
		return []string{}
	}
	return strings.Split(inGroups, ",")
}

// CheckToken returns the username and groups of a token (if valid).
// CheckToken returns an error in case the token is not valid.
func CheckToken(ctx context.Context, tx *sql.Tx, token string) (string, []string, error) {
	// TODO(neoaggelos): use prepared statements once we figure out what's wrong
	var username, groupsString string
	if err := tx.QueryRowContext(ctx, k8sdTokensSelectAuthByTokenSQL, token).Scan(&username, &groupsString); err != nil {
		return "", nil, fmt.Errorf("unknown token %s: %w", token, err)
	}

	return username, groupsToList(groupsString), nil
}

// GetOrCreateToken returns a token that matches the specified identify (username and groups).
// GetOrCreateToken will create an existing token (if availble).
// GetOrCreateToken will create a new token otherwise.
// GetOrCreateToken returns an error in case the auth is empty or a token could not be generated.
func GetOrCreateToken(ctx context.Context, tx *sql.Tx, username string, groups []string) (string, error) {
	if username == "" {
		return "", fmt.Errorf("username cannot be empty")
	}
	groupsString, err := groupsToString(groups)
	if err != nil {
		return "", fmt.Errorf("invalid groups: %w", err)
	}
	var token string
	if tx.QueryRowContext(ctx, k8sdTokensSelectTokenByAuthSQL, username, groupsString).Scan(&token) == nil {
		return token, nil
	}

	// TODO: make this crypto safe
	token = fmt.Sprintf("token-123-%d", time.Now().Nanosecond())

	if _, err := tx.ExecContext(ctx, k8sdTokensInsertTokenSQL, username, groupsString, token); err != nil {
		return "", fmt.Errorf("insert token query failed: %w", err)
	}

	return token, nil
}

// GetOrCreateToken_prepared is GetOrCreateToken, but uses prepared statements
// TODO(neoaggelos): try to figure out why prepared statements fail with ""
func GetOrCreateToken_prepared(ctx context.Context, tx *sql.Tx, username string, groups []string) (string, error) {
	if username == "" {
		return "", fmt.Errorf("username cannot be empty")
	}
	groupsString, err := groupsToString(groups)
	if err != nil {
		return "", fmt.Errorf("invalid groups: %w", err)
	}

	findStmt, err := cluster.Stmt(tx, k8sdTokensSelectTokenByAuthStmt)
	if err != nil {
		return "", fmt.Errorf("failed to prepare select statement: %w", err)
	}

	var token string
	if findStmt.QueryRowContext(ctx, username, groupsString).Scan(&token) == nil {
		return token, nil
	}

	insertStmt, err := cluster.Stmt(tx, k8sdTokensInsertTokenStmt)
	if err != nil {
		return "", fmt.Errorf("failed to prepare insert statement: %w", err)
	}

	token = fmt.Sprintf("token-123-%d", time.Now().Nanosecond())
	if _, err := insertStmt.ExecContext(ctx, username, groupsString, token); err != nil {
		return "", fmt.Errorf("insert query failed: %w", err)
	}

	return token, nil
}
