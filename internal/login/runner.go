package login

import (
    "context"
    "fmt"
    "os/exec"
    "time"
)

type Runner struct {
    BinaryPath string
}

func (r *Runner) Login(ctx context.Context, iface, username, password string, host string, teaching bool, ip string) error {
    if r.BinaryPath == "" { return fmt.Errorf("empty SZU-login binary path") }
    // Build args
    args := []string{"-i", iface}
    if host != "" { args = append(args, "--host", host) }
    if teaching && ip != "" { args = append(args, "--teaching-ip", ip) }
    if !teaching && ip != "" { args = append(args, "--dormitory-ip", ip) }
    args = append(args, "--username", username, "--password", password)
    cmd := exec.CommandContext(ctx, r.BinaryPath, args...)
    cmd.Stdout = nil
    cmd.Stderr = nil
    return cmd.Run()
}

func (r *Runner) LoginWithTimeout(iface, username, password, host string, teaching bool, ip string, timeout time.Duration) error {
    ctx, cancel := context.WithTimeout(context.Background(), timeout)
    defer cancel()
    return r.Login(ctx, iface, username, password, host, teaching, ip)
}

