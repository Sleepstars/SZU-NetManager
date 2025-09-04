# SZU-NetManager 开发与测试指南

本项目是在 OpenWRT 路由器环境中运行的“校园网多拨（mwan3）+ 负载均衡 + 故障转移”管理服务：
- 后端：Go（提供 REST API + WebSocket 实时日志，负责 SSH 执行 UCI、调用 SZU-login 登录、权重更新与回滚、故障转移监控）
- 前端：React + Ant Design（Vite）
- 数据库：SQLite（使用 CGO-free 库）

说明：当前仅保留“教学区”登录路径；不检测宿舍区。

---

## 快速启动（开发环境）

前置要求：
- Go 1.21+
- Node.js 18+/20+（建议 20）
- 可 SSH 访问到 OpenWRT 路由器（后端通过 SSH 执行 `uci`/`mwan3`）
- 本地可执行的 SZU-login 二进制（仅在“直接在路由器运行后端”时必须；Docker 方式会自动下载）

### 1) 启动后端（本地开发）

在 `SZU-NetManager` 目录：

```bash
#（可选）设置 Go 代理避免网络超时
export GOPROXY=https://goproxy.cn,direct

# 必要环境变量（按需修改）
export NM_LISTEN=":8080"                   # 后端监听端口
export NM_DB="szu-netmanager.db"          # SQLite 文件路径（自动初始化）
export NM_SSH_HOST="192.168.1.1"          # 路由器地址
export NM_SSH_PORT=22                      # 路由器 SSH 端口
export NM_SSH_USER="root"                 # 路由器 SSH 用户
export NM_SSH_KEY="$HOME/.ssh/id_rsa"     # 路由器 SSH 私钥（默认方式）
export NM_SSH_PASS=""                      # 可选：设置后改用“密码登录”
export NM_MONITOR_INTERVAL=30              # 故障检测间隔（秒）
export NM_MONITOR_URLS="https://www.baidu.com,https://www.qq.com"

# 若本机直接执行 SZU-login（仅在“后端运行于路由器或同一网络环境”时可用）
# Docker 部署无需设置，该二进制会在镜像构建时下载
export NM_SZU_LOGIN="/usr/local/bin/srun-login"

#（可选）前端静态目录（生产构建后）
export NM_WEB_DIR="web/dist"

# 启动后端
go run ./cmd/netmanager
```

说明：
- 后端会通过 SSH 串行执行 UCI 命令，原子化更新 `mwan3` 配置，失败自动回滚；重启 `mwan3` 时会有短暂网络中断。
- 登录调用 `SZU-login` 时会使用 `-i <网卡>` 绑定到指定 NIC（仅 Linux/路由器有效）。

### 2) 启动前端（Vite 开发服务器）

在 `SZU-NetManager/web` 目录：

```bash
npm i
npm run dev
```

- 浏览器打开 `http://localhost:5173`
- Vite 已代理到后端：`/api` 与 `/ws` -> `http://localhost:8080`

## 使用 Makefile 一键操作

常用命令（在 `SZU-NetManager` 根目录执行）：

```bash
# 安装依赖（Go + Web）
make init

# 构建后端到 bin/netmanager
make build-backend

# 运行后端（可在命令行覆盖环境变量）
make run-backend NM_SSH_HOST=192.168.1.1 NM_SSH_USER=root NM_SSH_KEY=~/.ssh/id_rsa

# 构建前端到 web/dist
make build-frontend

# 前端开发（Vite）
make dev-frontend

# 先构建前端，再构建后端
make build-all

# 构建 Docker 镜像（可覆盖 SZU_LOGIN_URL）
make docker-build SZU_LOGIN_URL=https://github.com/Sleepstars/SZU-login/releases/latest/download/srun-login-linux-amd64

# 以 host 网络模式运行镜像并传入 NM_* 配置
make docker-run NM_SSH_HOST=192.168.1.1 NM_SSH_USER=root NM_SSH_KEY=~/.ssh/id_rsa

# 打印当前生效的 NM_* 变量
make print-env

# 清理构建产物
make clean
```

---

## 开发/联调流程建议

1. 设置向导
   - 前端“设置向导”面板点击“加载接口”，后端执行 `uci show mwan3` 解析出可用接口（如 `wan`、`wanb`）。
   - 为每个 `wan*` 填写对应的宿主 NIC（如 `eth0`、`eth1`），点击“保存映射”。这决定 `-i <NIC>` 绑定到哪块物理口。
2. 账号池
   - 在“账号池”页面添加校园网账号（带宽可选 20/50/100/200）。
   - 系统会优先选择带宽高、且长时间未使用的账号，降低被挤占概率。
3. 触发登录
   - 在“设置向导”或“接口状态”页面可对指定 `wan` 点击“立即登录/尝试登录”。
   - 后端将：选择账号 → 调用 SZU-login 绑定到 `NIC` 登录（教学区路径）→ 根据带宽计算 `weight` → 备份配置 → `uci set` 更新对应 member 权重 → `commit` → 重启 `mwan3` → 状态校验，失败自动回滚。
4. 实时日志
   - “实时日志”面板通过 WebSocket `/ws` 展示关键阶段（如“开始为 wanb 接口登录新账号”、“配置已更新，正在重启 mwan3 服务...”、“登录成功！”）。
5. 健康检查与故障转移
   - 后端按 `NM_MONITOR_URLS` 定期探测；连续失败则对映射的接口触发重登。

---

## API 速查（用于自测）

```bash
# 探活
curl http://localhost:8080/api/health

# 读取 mwan3 成员映射（接口 -> member）
curl http://localhost:8080/api/mwan/interfaces

# 读取当前状态（原始输出）
curl http://localhost:8080/api/mwan/status

# 设置接口与 NIC 映射
curl -X POST http://localhost:8080/api/iface-map \
  -H 'Content-Type: application/json' \
  -d '{"WanIface":"wanb","Nic":"eth1"}'

# 添加账号
curl -X POST http://localhost:8080/api/accounts \
  -H 'Content-Type: application/json' \
  -d '{"Username":"u","Password":"p","Bandwidth":100}'

# 触发登录（教学区路径）
curl -X POST 'http://localhost:8080/api/login/start?wan=wanb'

# 备份/恢复配置（数据库）
curl -OJ http://localhost:8080/api/backup
curl -X POST --data-binary @szu-netmanager.db http://localhost:8080/api/restore
```

---

## Docker 部署（推荐用于真实联调）

镜像包含：已编译的后端 + 打包后的前端 + 已下载的 `srun-login` 二进制。

```bash
# 构建镜像（在 SZU-NetManager 根目录）
docker build -t szu-netmanager .

# 以 host 网络模式运行（可直接访问路由器网络）
docker run --rm --network host \
  -e NM_SSH_HOST=192.168.1.1 \
  -e NM_SSH_USER=root \
  -e NM_SSH_KEY=/root/.ssh/id_rsa \
  # 或使用密码登录：添加 -e NM_SSH_PASS=your-password
  -e NM_MONITOR_INTERVAL=30 \
  -e NM_MONITOR_URLS='https://www.baidu.com,https://www.qq.com' \
  szu-netmanager
```

- 访问 `http://<宿主机IP>:8080` 打开前端（容器内置静态资源 `NM_WEB_DIR=/app/web`）。
- 容器内 `srun-login` 已放置到 `/usr/local/bin/srun-login`，后端自动使用。
- 数据库 SQLite 驱动已内置（CGO-free），程序启动时自动创建并迁移表结构。

---

## 常见问题

- Go 依赖下载缓慢/超时：设置 `GOPROXY=https://goproxy.cn,direct`。
- `mwan3 restart` 瞬断：这是正常现象，后端会尽量缩短窗口并提供回滚。
- 登录只支持“教学区”：NetManager 仅调用 SZU-login 的教学区路径（`-i <NIC>` 绑定）。
- 本地开发无法真实登录：若后端不在路由器环境运行，登录绑定到本机 NIC，多数情况下仅用于流程验证；建议使用 Docker 在路由器侧联调。

---

## 目录结构（选摘）

- `cmd/netmanager`：后端入口
- `internal/api`：REST API + 协调逻辑
- `internal/sshqueue`：串行化 SSH 执行
- `internal/uci`：UCI 解析与操作、备份/回滚
- `internal/mwan`：权重应用与验证
- `internal/service`：账号池与映射存取
- `internal/monitor`：健康检测与故障转移
- `web`：前端（Vite + React + AntD）

如需进一步完善文档或添加示例配置，告诉我你的具体环境（路由器型号、接口名/NIC、OpenWRT 版本）。
