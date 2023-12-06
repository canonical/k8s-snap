package database

import (
	"context"
	"database/sql"
	_ "embed"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/canonical/microcluster/cluster"
)

type KubernetesIdentity struct {
	Username string   `json:"username"`
	Groups   []string `json:"groups"`
}

var (
	k8sdTokensStmts = map[string]int{
		"insert-token":         mustPrepareStatement("k8sd-tokens", "insert-token.sql"),
		"select-auth-by-token": mustPrepareStatement("k8sd-tokens", "select-auth-by-token.sql"),
		"select-token-by-auth": mustPrepareStatement("k8sd-tokens", "select-token-by-auth.sql"),
	}
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
	txStmt, err := cluster.Stmt(tx, k8sdTokensStmts["select-auth-by-token"])
	if err != nil {
		return "", nil, fmt.Errorf("failed to prepare statement: %w", err)
	}
	var username, groupsString string
	if err := txStmt.QueryRowContext(ctx, token).Scan(&username, &groupsString); err != nil {
		if err == sql.ErrNoRows {
			return "", nil, fmt.Errorf("invalid token")
		}
		return "", nil, fmt.Errorf("failed to check token: %w", err)
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
	selectTxStmt, err := cluster.Stmt(tx, k8sdTokensStmts["select-token-by-auth"])
	if err != nil {
		return "", fmt.Errorf("failed to prepare select statement: %w", err)
	}
	var token string
	if selectTxStmt.QueryRowContext(ctx, username, groupsString).Scan(&token) == nil {
		return token, nil
	}

	// TODO: make this crypto safe
	token = fmt.Sprintf("token-123-%d", time.Now().Nanosecond())
	insertTxStmt, err := cluster.Stmt(tx, k8sdTokensStmts["insert-token"])
	if err != nil {
		return "", fmt.Errorf("failed to prepare insert statement: %w", err)
	}
	if _, err := insertTxStmt.ExecContext(ctx, username, groupsString, token); err != nil {
		return "", fmt.Errorf("insert token query failed: %w", err)
	}

	return token, nil
}
