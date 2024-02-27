package database

import (
	"context"
	"crypto/rand"
	"database/sql"
	_ "embed"
	"encoding/hex"
	"fmt"
	"sort"
	"strings"

	"github.com/canonical/microcluster/cluster"
)

var (
	k8sdTokensStmts = map[string]int{
		"insert-token":       MustPrepareStatement("kubernetes-auth-tokens", "insert-token.sql"),
		"select-by-token":    MustPrepareStatement("kubernetes-auth-tokens", "select-by-token.sql"),
		"select-by-username": MustPrepareStatement("kubernetes-auth-tokens", "select-by-username.sql"),
		"delete-by-token":    MustPrepareStatement("kubernetes-auth-tokens", "delete-by-token.sql"),
		"delete-by-username": MustPrepareStatement("kubernetes-auth-tokens", "delete-by-username.sql"),
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
	txStmt, err := cluster.Stmt(tx, k8sdTokensStmts["select-by-token"])
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
// GetOrCreateToken will create an existing token (if available).
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
	selectTxStmt, err := cluster.Stmt(tx, k8sdTokensStmts["select-by-username"])
	if err != nil {
		return "", fmt.Errorf("failed to prepare select statement: %w", err)
	}
	var token string
	if selectTxStmt.QueryRowContext(ctx, username, groupsString).Scan(&token) == nil {
		return token, nil
	}

	// generate random bytes for the token
	b := make([]byte, 20)
	if _, err := rand.Read(b); err != nil {
		return "", fmt.Errorf("is the system entropy low? failed to get random bytes: %w", err)
	}
	token = fmt.Sprintf("token::%s", hex.EncodeToString(b))

	insertTxStmt, err := cluster.Stmt(tx, k8sdTokensStmts["insert-token"])
	if err != nil {
		return "", fmt.Errorf("failed to prepare insert statement: %w", err)
	}
	if _, err := insertTxStmt.ExecContext(ctx, username, groupsString, token); err != nil {
		return "", fmt.Errorf("insert token query failed: %w", err)
	}

	return token, nil
}

// DeleteTokenOf deletes the token of the specified user and groups (if any).
// DeleteTokenOf returns nil if there is no token for the specified user and groups.
func DeleteTokenOf(ctx context.Context, tx *sql.Tx, username string, groups []string) error {
	if username == "" {
		return fmt.Errorf("username cannot be empty")
	}
	groupsString, err := groupsToString(groups)
	if err != nil {
		return fmt.Errorf("invalid groups: %w", err)
	}

	deleteTxStmt, err := cluster.Stmt(tx, k8sdTokensStmts["delete-by-username"])
	if err != nil {
		return fmt.Errorf("failed to prepare delete statement: %w", err)
	}
	if _, err := deleteTxStmt.ExecContext(ctx, username, groupsString); err != nil {
		return fmt.Errorf("delete token query failed: %w", err)
	}
	return nil
}

// DeleteToken deletes the specified token (if any).
// DeleteToken returns nil if the token is not valid.
func DeleteToken(ctx context.Context, tx *sql.Tx, token string) (string, error) {
	if token == "" {
		return "", fmt.Errorf("token cannot be empty")
	}

	deleteTxStmt, err := cluster.Stmt(tx, k8sdTokensStmts["delete-by-token"])
	if err != nil {
		return "", fmt.Errorf("failed to prepare delete statement: %w", err)
	}
	if _, err := deleteTxStmt.ExecContext(ctx, token); err != nil {
		return "", fmt.Errorf("delete token query failed: %w", err)
	}
	return token, nil
}
