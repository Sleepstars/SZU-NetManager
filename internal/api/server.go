package api

import (
    "context"
    "database/sql"
    "encoding/json"
    "fmt"
    "io"
    "net/http"
    "os"
    "path/filepath"
    "time"

    "github.com/Sleepstars/SZU-NetManager/internal/login"
    "github.com/Sleepstars/SZU-NetManager/internal/mwan"
    "github.com/Sleepstars/SZU-NetManager/internal/service"
    "github.com/Sleepstars/SZU-NetManager/internal/uci"
    "github.com/Sleepstars/SZU-NetManager/internal/ws"
    "github.com/Sleepstars/SZU-NetManager/internal/weights"
)

type Server struct {
    DB        *sql.DB
    Hub       *ws.Hub
    Accounts  *service.Accounts
    IfaceMap  *service.IfaceMap
    UCI       *uci.Client
    MWAN      *mwan.Service
    Runner    *login.Runner
    DBPath    string
}

func New(dbConn *sql.DB, hub *ws.Hub, dbPath string, uciClient *uci.Client, runner *login.Runner) *Server {
    return &Server{
        DB:        dbConn,
        Hub:       hub,
        Accounts:  service.NewAccounts(dbConn),
        IfaceMap:  service.NewIfaceMap(dbConn),
        UCI:       uciClient,
        MWAN:      mwan.New(uciClient),
        Runner:    runner,
        DBPath:    dbPath,
    }
}

func (s *Server) Routes() *http.ServeMux {
    mux := http.NewServeMux()
    mux.HandleFunc("/api/health", s.handleHealth)
    mux.HandleFunc("/api/mwan/interfaces", s.handleMWANInterfaces)
    mux.HandleFunc("/api/mwan/status", s.handleMWANStatus)
    mux.HandleFunc("/api/iface-map", s.handleIfaceMap)
    mux.HandleFunc("/api/accounts", s.handleAccounts)
    mux.HandleFunc("/api/login/start", s.handleLoginStart)
    mux.HandleFunc("/api/backup", s.handleBackup)
    mux.HandleFunc("/api/restore", s.handleRestore)
    return mux
}

func writeJSON(w http.ResponseWriter, v any) {
    w.Header().Set("Content-Type", "application/json")
    _ = json.NewEncoder(w).Encode(v)
}

func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) { writeJSON(w, map[string]any{"ok": true}) }

func (s *Server) handleMWANInterfaces(w http.ResponseWriter, r *http.Request) {
    raw, err := s.UCI.Show()
    if err != nil { http.Error(w, err.Error(), 500); return }
    mapping := s.UCI.MemberMapping(raw)
    // unique iface list
    out := map[string]any{"member_map": mapping}
    writeJSON(w, out)
}

func (s *Server) handleMWANStatus(w http.ResponseWriter, r *http.Request) {
    raw, err := s.UCI.Status()
    if err != nil { http.Error(w, err.Error(), 500); return }
    writeJSON(w, map[string]any{"status": raw})
}

func (s *Server) handleIfaceMap(w http.ResponseWriter, r *http.Request) {
    switch r.Method {
    case http.MethodGet:
        m, err := s.IfaceMap.All(r.Context())
        if err != nil { http.Error(w, err.Error(), 500); return }
        writeJSON(w, m)
    case http.MethodPost:
        var req struct{ WanIface, Nic string }
        if err := json.NewDecoder(r.Body).Decode(&req); err != nil { http.Error(w, err.Error(), 400); return }
        if req.WanIface == "" || req.Nic == "" { http.Error(w, "wan_iface and nic required", 400); return }
        if err := s.IfaceMap.Set(r.Context(), req.WanIface, req.Nic); err != nil { http.Error(w, err.Error(), 500); return }
        writeJSON(w, map[string]any{"ok": true})
    default:
        http.Error(w, "method not allowed", 405)
    }
}

func (s *Server) handleAccounts(w http.ResponseWriter, r *http.Request) {
    switch r.Method {
    case http.MethodGet:
        list, err := s.Accounts.List(r.Context())
        if err != nil { http.Error(w, err.Error(), 500); return }
        writeJSON(w, list)
    case http.MethodPost:
        var req struct{ Username, Password string; Bandwidth int }
        if err := json.NewDecoder(r.Body).Decode(&req); err != nil { http.Error(w, err.Error(), 400); return }
        if req.Username == "" || req.Password == "" || req.Bandwidth <= 0 { http.Error(w, "invalid payload", 400); return }
        id, err := s.Accounts.Add(r.Context(), req.Username, req.Password, req.Bandwidth)
        if err != nil { http.Error(w, err.Error(), 500); return }
        writeJSON(w, map[string]any{"id": id})
    default:
        http.Error(w, "method not allowed", 405)
    }
}

// handleLoginStart triggers login for a given mwan iface (e.g., wan or wanb), selects next account, applies weight, and broadcasts logs.
func (s *Server) handleLoginStart(w http.ResponseWriter, r *http.Request) {
    wanIface := r.URL.Query().Get("wan")
    if wanIface == "" { http.Error(w, "wan query required", 400); return }
    go s.LoginForIface(r.Context(), wanIface)
    writeJSON(w, map[string]any{"ok": true})
}

func (s *Server) LoginForIface(ctx context.Context, wanIface string) {
    s.Hub.Broadcast(fmt.Sprintf("开始为 %s 接口登录新账号", wanIface))

    nic, err := s.IfaceMap.Get(ctx, wanIface)
    if err != nil { s.Hub.Broadcast(fmt.Sprintf("获取网卡映射失败: %v", err)); return }
    if nic == "" { s.Hub.Broadcast("未配置网卡映射，请先在设置中选择 NIC"); return }

    acct, err := s.Accounts.NextCandidate(ctx)
    if err != nil { s.Hub.Broadcast(fmt.Sprintf("选择账号失败: %v", err)); return }
    if acct == nil { s.Hub.Broadcast("没有可用账号"); return }

    _ = s.Accounts.UpdateState(ctx, acct.ID, "CONNECTING")

    // Invoke SZU-login
    if err := s.Runner.LoginWithTimeout(nic, acct.Username, acct.Password, "", true, "", 40*time.Second); err != nil {
        s.Hub.Broadcast(fmt.Sprintf("%s 接口登录失败: %v", wanIface, err))
        _ = s.Accounts.UpdateState(ctx, acct.ID, "RETRYING")
        return
    }

    _ = s.Accounts.UpdateState(ctx, acct.ID, "ONLINE")
    _ = s.Accounts.MarkUsedNow(ctx, acct.ID)
    s.Hub.Broadcast(fmt.Sprintf("%s 接口登录成功！", wanIface))

    // Apply weight based on bandwidth
    w := weights.FromBandwidth(acct.Bandwidth)
    s.Hub.Broadcast(fmt.Sprintf("配置已更新为权重 %d，正在重启 mwan3 服务...", w))
    if err := s.MWAN.ApplyWeight(wanIface, w); err != nil {
        s.Hub.Broadcast(fmt.Sprintf("mwan3 应用权重失败并已回滚: %v", err))
        return
    }
    s.Hub.Broadcast("mwan3 已重启并生效")
}

func (s *Server) handleBackup(w http.ResponseWriter, r *http.Request) {
    w.Header().Set("Content-Disposition", "attachment; filename= szu-netmanager.db")
    w.Header().Set("Content-Type", "application/octet-stream")
    f, err := os.Open(s.DBPath)
    if err != nil { http.Error(w, err.Error(), 500); return }
    defer f.Close()
    _, _ = io.Copy(w, f)
}

func (s *Server) handleRestore(w http.ResponseWriter, r *http.Request) {
    if r.Method != http.MethodPost { http.Error(w, "method not allowed", 405); return }
    // Accept raw file body to replace DB
    tmp := filepath.Join(filepath.Dir(s.DBPath), fmt.Sprintf("restore-%d.db", time.Now().Unix()))
    f, err := os.Create(tmp)
    if err != nil { http.Error(w, err.Error(), 500); return }
    if _, err := io.Copy(f, r.Body); err != nil { f.Close(); http.Error(w, err.Error(), 500); return }
    _ = f.Close()
    // replace
    if err := os.Rename(tmp, s.DBPath); err != nil { http.Error(w, err.Error(), 500); return }
    writeJSON(w, map[string]any{"ok": true})
}
