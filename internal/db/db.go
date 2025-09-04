package db

import (
    "database/sql"
    _ "github.com/ncruces/go-sqlite3/driver"
)

func Open(path string) (*sql.DB, error) {
    return sql.Open("sqlite3", path)
}

func Migrate(db *sql.DB) error {
    stmts := []string{
        `CREATE TABLE IF NOT EXISTS accounts (
            id INTEGER PRIMARY KEY AUTOINCREMENT,
            username TEXT NOT NULL,
            password TEXT NOT NULL,
            bandwidth INTEGER NOT NULL DEFAULT 50,
            status TEXT NOT NULL DEFAULT 'IDLE',
            last_used_at INTEGER NOT NULL DEFAULT 0,
            disabled INTEGER NOT NULL DEFAULT 0
        );`,
        `CREATE TABLE IF NOT EXISTS iface_map (
            id INTEGER PRIMARY KEY AUTOINCREMENT,
            wan_iface TEXT NOT NULL,
            nic_name TEXT NOT NULL
        );`,
        `CREATE UNIQUE INDEX IF NOT EXISTS idx_iface_map_wan ON iface_map(wan_iface);`,
        `CREATE TABLE IF NOT EXISTS kv (
            k TEXT PRIMARY KEY,
            v TEXT NOT NULL
        );`,
    }
    for _, s := range stmts {
        if _, err := db.Exec(s); err != nil { return err }
    }
    return nil
}
