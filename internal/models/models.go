package models

type Bandwidth int

const (
    BW20  Bandwidth = 20
    BW50  Bandwidth = 50
    BW100 Bandwidth = 100
    BW200 Bandwidth = 200
)

type AccountState string

const (
    StateIdle       AccountState = "IDLE"
    StateConnecting AccountState = "CONNECTING"
    StateOnline     AccountState = "ONLINE"
    StateRetrying   AccountState = "RETRYING"
    StateFailed     AccountState = "FAILED"
    StateDisabled   AccountState = "DISABLED"
)

type Account struct {
    ID         int64
    Username   string
    Password   string
    Bandwidth  Bandwidth
    Status     AccountState
    LastUsedAt int64 // unix seconds
    Disabled   bool
}

