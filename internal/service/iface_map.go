package service

import (
    "context"
    "database/sql"
)

type IfaceMap struct { db *sql.DB }

func NewIfaceMap(db *sql.DB) *IfaceMap { return &IfaceMap{db: db} }

// Set maps a mwan interface name (e.g., "wanb") to a host NIC (e.g., "eth1").
func (m *IfaceMap) Set(ctx context.Context, wanIface, nic string) error {
    _, err := m.db.ExecContext(ctx, `INSERT INTO iface_map (wan_iface, nic_name) VALUES (?, ?)
            ON CONFLICT(wan_iface) DO UPDATE SET nic_name=excluded.nic_name`, wanIface, nic)
    if err != nil {
        // Some sqlite versions need an explicit unique constraint. Fallback: try update if exists.
        _, err2 := m.db.ExecContext(ctx, `UPDATE iface_map SET nic_name=? WHERE wan_iface=?`, nic, wanIface)
        if err2 == nil { return nil }
        return err
    }
    return nil
}

func (m *IfaceMap) Get(ctx context.Context, wanIface string) (string, error) {
    row := m.db.QueryRowContext(ctx, `SELECT nic_name FROM iface_map WHERE wan_iface=?`, wanIface)
    var nic string
    if err := row.Scan(&nic); err != nil { return "", err }
    return nic, nil
}

func (m *IfaceMap) All(ctx context.Context) (map[string]string, error) {
    rows, err := m.db.QueryContext(ctx, `SELECT wan_iface, nic_name FROM iface_map`)
    if err != nil { return nil, err }
    defer rows.Close()
    out := map[string]string{}
    for rows.Next() {
        var w, n string
        if err := rows.Scan(&w, &n); err != nil { return nil, err }
        out[w] = n
    }
    return out, rows.Err()
}

