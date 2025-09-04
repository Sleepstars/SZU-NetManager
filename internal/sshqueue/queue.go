package sshqueue

import (
    "bytes"
    "fmt"
    "golang.org/x/crypto/ssh"
    "io"
    "os"
    "sync"
)

type Queue struct {
    mu    sync.Mutex
    addr  string
    conf  *ssh.ClientConfig
}

func New(addr, user, keyPath string) (*Queue, error) {
    key, err := os.ReadFile(keyPath)
    if err != nil { return nil, fmt.Errorf("read key: %w", err) }
    signer, err := ssh.ParsePrivateKey(key)
    if err != nil { return nil, fmt.Errorf("parse key: %w", err) }
    cfg := &ssh.ClientConfig{ User: user, Auth: []ssh.AuthMethod{ssh.PublicKeys(signer)}, HostKeyCallback: ssh.InsecureIgnoreHostKey() }
    return &Queue{ addr: addr, conf: cfg }, nil
}

// NewWithPassword creates a queue using password authentication.
func NewWithPassword(addr, user, password string) (*Queue, error) {
    cfg := &ssh.ClientConfig{ User: user, Auth: []ssh.AuthMethod{ssh.Password(password)}, HostKeyCallback: ssh.InsecureIgnoreHostKey() }
    return &Queue{ addr: addr, conf: cfg }, nil
}

// Exec runs a command synchronously with internal mutex to ensure serialization.
func (q *Queue) Exec(cmd string) (string, error) {
    q.mu.Lock(); defer q.mu.Unlock()
    client, err := ssh.Dial("tcp", q.addr, q.conf)
    if err != nil { return "", fmt.Errorf("ssh dial: %w", err) }
    defer client.Close()
    sess, err := client.NewSession()
    if err != nil { return "", fmt.Errorf("ssh session: %w", err) }
    defer sess.Close()
    var stdout, stderr bytes.Buffer
    sess.Stdout = &stdout
    sess.Stderr = &stderr
    if err := sess.Run(cmd); err != nil { return "", fmt.Errorf("run: %w, stderr: %s", err, stderr.String()) }
    io.Copy(io.Discard, &stderr)
    return stdout.String(), nil
}
