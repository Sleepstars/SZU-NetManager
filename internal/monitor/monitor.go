package monitor

import (
    "context"
    "net/http"
    "time"

    "github.com/Sleepstars/SZU-NetManager/internal/ws"
)

type Trigger func(ctx context.Context, wanIface string)

type Config struct {
    Interval time.Duration
    TestURLs []string
}

type Provider interface { // minimal interface to fetch iface map
    All(ctx context.Context) (map[string]string, error)
}

type Monitor struct {
    hub      *ws.Hub
    cfg      Config
    prov     Provider
    trigger  Trigger
    client   *http.Client
}

func New(h *ws.Hub, cfg Config, prov Provider, trigger Trigger) *Monitor {
    return &Monitor{hub: h, cfg: cfg, prov: prov, trigger: trigger, client: &http.Client{Timeout: 5 * time.Second}}
}

func (m *Monitor) Run(ctx context.Context) {
    ticker := time.NewTicker(m.cfg.Interval)
    defer ticker.Stop()
    m.checkAndMaybeTrigger(ctx)
    for {
        select {
        case <-ctx.Done():
            return
        case <-ticker.C:
            m.checkAndMaybeTrigger(ctx)
        }
    }
}

func (m *Monitor) checkAndMaybeTrigger(ctx context.Context) {
    if m.cfg.Interval <= 0 || len(m.cfg.TestURLs) == 0 { return }
    ok := false
    for _, u := range m.cfg.TestURLs {
        req, _ := http.NewRequestWithContext(ctx, http.MethodGet, u, nil)
        resp, err := m.client.Do(req)
        if err == nil && resp.StatusCode >= 200 && resp.StatusCode < 500 {
            ok = true
        }
        if resp != nil { resp.Body.Close() }
        if ok { break }
    }
    if ok { return }
    m.hub.Broadcast("检测到网络不可用，触发故障转移")
    ifaces, err := m.prov.All(ctx)
    if err != nil { m.hub.Broadcast("读取接口映射失败"); return }
    for wanIface := range ifaces {
        m.trigger(ctx, wanIface)
    }
}

