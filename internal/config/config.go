package config

import (
    "fmt"
    "os"
)

type Config struct {
    ListenAddr   string
    DBPath       string
    SSHHost      string
    SSHPort      int
    SSHUser      string
    SSHKeyPath   string
    SZULoginPath string
    MonitorURLs  []string
    MonitorEvery int // seconds
    WebDir       string
}

func Load() *Config {
    cfg := &Config{
        ListenAddr:   getEnv("NM_LISTEN", ":8080"),
        DBPath:       getEnv("NM_DB", "szu-netmanager.db"),
        SSHHost:      getEnv("NM_SSH_HOST", "127.0.0.1"),
        SSHUser:      getEnv("NM_SSH_USER", "root"),
        SZULoginPath: getEnv("NM_SZU_LOGIN", "/usr/local/bin/srun-login"),
    }
    // default port 22
    cfg.SSHPort = 22
    if v := os.Getenv("NM_SSH_PORT"); v != "" {
        // ignore error silently; keep default
        var p int
        _, _ = fmt.Sscanf(v, "%d", &p)
        if p > 0 { cfg.SSHPort = p }
    }
    // default key path
    cfg.SSHKeyPath = getEnv("NM_SSH_KEY", "/root/.ssh/id_rsa")
    // monitor
    cfg.MonitorEvery = 30
    if v := os.Getenv("NM_MONITOR_INTERVAL"); v != "" {
        var s int
        _, _ = fmt.Sscanf(v, "%d", &s)
        if s > 0 { cfg.MonitorEvery = s }
    }
    urls := os.Getenv("NM_MONITOR_URLS")
    if urls == "" {
        cfg.MonitorURLs = []string{"https://www.baidu.com", "https://www.qq.com"}
    } else {
        // split by comma
        var out []string
        cur := ""
        for _, ch := range urls {
            if ch == ',' {
                if cur != "" { out = append(out, cur) }
                cur = ""
            } else { cur += string(ch) }
        }
        if cur != "" { out = append(out, cur) }
        cfg.MonitorURLs = out
    }
    // web dir (for embedded SPA)
    cfg.WebDir = getEnv("NM_WEB_DIR", "web/dist")
    return cfg
}

func getEnv(key, def string) string {
    if v := os.Getenv(key); v != "" {
        return v
    }
    return def
}
