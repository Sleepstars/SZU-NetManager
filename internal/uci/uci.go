package uci

import (
    "fmt"
    "regexp"
    "strings"

    "github.com/Sleepstars/SZU-NetManager/internal/sshqueue"
)

type Client struct { q *sshqueue.Queue }

func New(q *sshqueue.Queue) *Client { return &Client{q: q} }

// Show returns raw `uci show mwan3` output.
func (c *Client) Show() (string, error) { return c.q.Exec("uci show mwan3") }

// MemberMapping parses `uci show mwan3` and returns interface->member name.
func (c *Client) MemberMapping(raw string) map[string]string {
    // e.g., mwan3.wan_m1_w2.interface='wan'
    re := regexp.MustCompile(`^mwan3\.(\S+)\.interface='(\S+)'$`)
    mapping := map[string]string{}
    lines := strings.Split(raw, "\n")
    for _, ln := range lines {
        ln = strings.TrimSpace(ln)
        if m := re.FindStringSubmatch(ln); len(m) == 3 {
            member := m[1]
            iface := m[2]
            mapping[iface] = member
        }
    }
    return mapping
}

func (c *Client) SetMemberWeight(member string, weight int) error {
    _, err := c.q.Exec(fmt.Sprintf("uci set mwan3.%s.weight='%d'", member, weight))
    return err
}

func (c *Client) Commit() error { _, err := c.q.Exec("uci commit mwan3"); return err }
func (c *Client) Restart() error { _, err := c.q.Exec("/etc/init.d/mwan3 restart"); return err }
func (c *Client) Status() (string, error) { return c.q.Exec("mwan3 status") }

// Backup and rollback helpers
func (c *Client) Backup() (string, error) {
    // returns path to backup file
    out, err := c.q.Exec("cp /etc/config/mwan3 /tmp/mwan3.backup && echo /tmp/mwan3.backup")
    if err != nil { return "", err }
    return strings.TrimSpace(out), nil
}
func (c *Client) Rollback(backupPath string) error {
    if backupPath == "" { return fmt.Errorf("empty backup path") }
    _, err := c.q.Exec(fmt.Sprintf("cp %s /etc/config/mwan3", backupPath))
    return err
}

