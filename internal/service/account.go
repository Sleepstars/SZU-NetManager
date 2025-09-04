package service

import (
    "context"
    "database/sql"
    "errors"
    "time"
)

type Account struct {
    ID         int64
    Username   string
    Password   string
    Bandwidth  int
    Status     string
    LastUsedAt int64
    Disabled   bool
}

type Accounts struct { db *sql.DB }

func NewAccounts(db *sql.DB) *Accounts { return &Accounts{db: db} }

func (a *Accounts) List(ctx context.Context) ([]Account, error) {
    rows, err := a.db.QueryContext(ctx, `SELECT id, username, password, bandwidth, status, last_used_at, disabled FROM accounts ORDER BY id ASC`)
    if err != nil { return nil, err }
    defer rows.Close()
    var out []Account
    for rows.Next() {
        var x Account
        var disabledInt int
        if err := rows.Scan(&x.ID, &x.Username, &x.Password, &x.Bandwidth, &x.Status, &x.LastUsedAt, &disabledInt); err != nil { return nil, err }
        x.Disabled = disabledInt != 0
        out = append(out, x)
    }
    return out, rows.Err()
}

func (a *Accounts) Add(ctx context.Context, username, password string, bandwidth int) (int64, error) {
    res, err := a.db.ExecContext(ctx, `INSERT INTO accounts (username, password, bandwidth, status, last_used_at, disabled) VALUES (?, ?, ?, 'IDLE', 0, 0)`, username, password, bandwidth)
    if err != nil { return 0, err }
    return res.LastInsertId()
}

func (a *Accounts) UpdateState(ctx context.Context, id int64, state string) error {
    _, err := a.db.ExecContext(ctx, `UPDATE accounts SET status=? WHERE id=?`, state, id)
    return err
}

func (a *Accounts) MarkUsedNow(ctx context.Context, id int64) error {
    _, err := a.db.ExecContext(ctx, `UPDATE accounts SET last_used_at=? WHERE id=?`, time.Now().Unix(), id)
    return err
}

// NextCandidate selects the next account according to rules:
// - not disabled
// - prefer higher bandwidth
// - prefer longer time since last_used_at
// - ignore FAILED accounts by default
func (a *Accounts) NextCandidate(ctx context.Context) (*Account, error) {
    row := a.db.QueryRowContext(ctx, `
        SELECT id, username, password, bandwidth, status, last_used_at, disabled
        FROM accounts
        WHERE disabled=0 AND status <> 'FAILED'
        ORDER BY bandwidth DESC, CASE last_used_at WHEN 0 THEN -9223372036854775808 ELSE last_used_at END ASC
        LIMIT 1`)
    var x Account
    var disabledInt int
    if err := row.Scan(&x.ID, &x.Username, &x.Password, &x.Bandwidth, &x.Status, &x.LastUsedAt, &disabledInt); err != nil {
        if errors.Is(err, sql.ErrNoRows) { return nil, nil }
        return nil, err
    }
    x.Disabled = disabledInt != 0
    return &x, nil
}

