package main

import (
    "context"
    "fmt"
    "log"
    "net/http"
    "os"
    "os/signal"
    "syscall"
    "time"

    "github.com/Sleepstars/SZU-NetManager/internal/config"
    "github.com/Sleepstars/SZU-NetManager/internal/db"
    "github.com/Sleepstars/SZU-NetManager/internal/api"
    "github.com/Sleepstars/SZU-NetManager/internal/login"
    "github.com/Sleepstars/SZU-NetManager/internal/monitor"
    "github.com/Sleepstars/SZU-NetManager/internal/sshqueue"
    "github.com/Sleepstars/SZU-NetManager/internal/uci"
    "github.com/Sleepstars/SZU-NetManager/internal/httpmw"
    "github.com/Sleepstars/SZU-NetManager/internal/ws"
)

func main() {
    cfg := config.Load()

    // DB
    database, err := db.Open(cfg.DBPath)
    if err != nil {
        log.Fatalf("open db: %v", err)
    }
    defer database.Close()
    if err := db.Migrate(database); err != nil {
        log.Fatalf("migrate db: %v", err)
    }

    // WebSocket hub
    hub := ws.NewHub()
    go hub.Run()

    // Dependencies for API
    q, err := sshqueue.New(
        cfg.SSHHost+":"+fmt.Sprintf("%d", cfg.SSHPort),
        cfg.SSHUser,
        cfg.SSHKeyPath,
    )
    if err != nil { log.Fatalf("ssh queue: %v", err) }
    uciClient := uci.New(q)
    runner := &login.Runner{ BinaryPath: cfg.SZULoginPath }
    server := api.New(database, hub, cfg.DBPath, uciClient, runner)

    mux := server.Routes()
    mux.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) { ws.ServeWS(hub, w, r) })

    // Monitor (failover)
    mon := monitor.New(hub, monitor.Config{Interval: time.Duration(cfg.MonitorEvery) * time.Second, TestURLs: cfg.MonitorURLs}, server.IfaceMap, func(ctx context.Context, wanIface string) {
        // delegate to server
        server.LoginForIface(ctx, wanIface)
    })
    go mon.Run(context.Background())

    // Serve embedded UI if present
    if cfg.WebDir != "" {
        fs := http.FileServer(http.Dir(cfg.WebDir))
        mux.Handle("/", spaFallback(fs, cfg.WebDir))
    }

    handler := httpmw.CORS(mux)

    srv := &http.Server{
        Addr:              cfg.ListenAddr,
        Handler:           handler,
        ReadHeaderTimeout: 5 * time.Second,
    }

    go func() {
        log.Printf("SZU-NetManager backend listening on %s", cfg.ListenAddr)
        if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
            log.Fatalf("http listen: %v", err)
        }
    }()

    // Graceful shutdown
    stop := make(chan os.Signal, 1)
    signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
    <-stop
    ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
    defer cancel()
    _ = srv.Shutdown(ctx)
}

// spaFallback serves index.html for unknown paths (for SPA routing), while letting /api and /ws pass through.
func spaFallback(next http.Handler, dir string) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        // Let API and WS pass
        if r.URL.Path == "/" || r.URL.Path == "/index.html" ||
            (len(r.URL.Path) >= 4 && r.URL.Path[:4] == "/api") ||
            (len(r.URL.Path) >= 3 && r.URL.Path[:3] == "/ws") {
            next.ServeHTTP(w, r)
            return
        }
        // Try to serve existing file; if not found, fall back to index.html
        fpath := dir + r.URL.Path
        if _, err := os.Stat(fpath); err == nil {
            next.ServeHTTP(w, r)
            return
        }
        http.ServeFile(w, r, dir+"/index.html")
    })
}
