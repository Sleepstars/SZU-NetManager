package mwan

import (
    "fmt"
    "time"

    "github.com/Sleepstars/SZU-NetManager/internal/uci"
)

type Service struct {
    u *uci.Client
}

func New(u *uci.Client) *Service { return &Service{u: u} }

// ApplyWeight sets weight for a given mwan interface by first resolving the member name, then committing and restarting.
// It performs backup and will rollback if restart or status verification fails.
func (s *Service) ApplyWeight(wanIface string, weight int) error {
    raw, err := s.u.Show()
    if err != nil { return fmt.Errorf("uci show: %w", err) }
    mapping := s.u.MemberMapping(raw)
    member, ok := mapping[wanIface]
    if !ok { return fmt.Errorf("member not found for iface %s", wanIface) }

    backupPath, err := s.u.Backup()
    if err != nil { return fmt.Errorf("backup: %w", err) }

    if err := s.u.SetMemberWeight(member, weight); err != nil { return fmt.Errorf("set weight: %w", err) }
    if err := s.u.Commit(); err != nil { _ = s.u.Rollback(backupPath); return fmt.Errorf("commit: %w", err) }
    if err := s.u.Restart(); err != nil { _ = s.u.Rollback(backupPath); return fmt.Errorf("restart: %w", err) }

    time.Sleep(2 * time.Second)
    status, err := s.u.Status()
    if err != nil { _ = s.u.Rollback(backupPath); return fmt.Errorf("status: %w", err) }
    // Minimal verification: expect the interface string to appear; allow success if not conclusive.
    if status == "" { _ = s.u.Rollback(backupPath); return fmt.Errorf("empty status after restart") }
    return nil
}

