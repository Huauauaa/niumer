# niumer

使用 [Wails v2](https://wails.io/) 构建的桌面应用：Go 后端 + **React** + **TypeScript** + **Vite** + **Tailwind CSS**。界面布局参考 VS Code（活动栏、侧边栏、编辑器标签、底部面板、状态栏）。

## 环境要求

- **Go** 1.22+（见 `go.mod` 与 `toolchain`）
- **Node.js** 与 **npm**（用于前端依赖与构建）
- （可选）**Wails CLI**，用于 `wails dev` / `wails build`  
  ```bash
  go install github.com/wailsapp/wails/v2/cmd/wails@latest
  ```
- 使用「考勤」刷新功能时，本机需能启动 **Chrome / Chromium**（或由 `WORK_HOUR_CHROME_PATH` 指定可执行文件）

## 开发

在项目根目录**任选其一**（推荐前两种，可避免 `goproxy.io` 拉取 `modernc.org` 时出现 **502**）：

```bash
make dev
```

或：

```bash
chmod +x scripts/dev.sh   # 首次
./scripts/dev.sh
```

若直接使用 `wails dev`，请确保未将 `GOPROXY` 设为易失败的 `https://goproxy.io`；可改为：

```bash
export GOPROXY=https://proxy.golang.org,direct
# 国内可试: export GOPROXY=https://goproxy.cn,direct
wails dev
```

使用 **Cursor / VS Code** 打开本仓库时，内置终端会读取 **`.vscode/settings.json`** 中的 `GOPROXY`（覆盖你 shell 里配置的 `goproxy.io`），一般可直接运行 `wails dev`。若仍失败，请在外部终端（如系统「终端.app」）里先 `export GOPROXY=...` 再执行。

`wails dev` 会安装前端依赖、启动 Vite 开发服务并启动桌面窗口。

**考勤 mock 与 Pull Request 列表**：另开终端在项目根执行 `go run ./cmd/mockserver`（默认 `http://127.0.0.1:17890`）。前端 **Pull Request** 活动栏会从 `GET /pull-request` 拉分页列表；选中项在右侧用 iframe 打开该项的 `url`（一般为同机 `GET /pr-preview/{id}` 页面）。可通过环境变量 **`VITE_PULL_REQUEST_API_BASE`** 指向其他基地址。

### Makefile 快捷命令

| 命令 | 说明 |
|------|------|
| `make dev` | 使用 `Makefile` 中的 `GOPROXY`（默认 `proxy.golang.org`）运行 `wails dev` |
| `make tidy` | 同上代理执行 `go mod tidy`，拉全依赖 |
| `make build` | 生产构建（`wails build`） |

更换镜像：`make dev GO_MOD_PROXY=https://goproxy.cn,direct`

若你更习惯手动分步执行：

```bash
cd frontend
npm install
npm run dev
```

另开一个终端，在生成好 `frontend/dist` 的前提下用 Go 运行（通常仍推荐直接使用 `wails dev`）。

## 应用配置

配置文件路径由 Go 的 `os.UserConfigDir()` 决定，固定为其中的 **`niumer/config.json`**：

- **macOS**：`~/Library/Application Support/niumer/config.json`
- **Linux**（常见）：`~/.config/niumer/config.json`（或 `$XDG_CONFIG_HOME/niumer/config.json`）
- **Windows**：`%AppData%\niumer\config.json`

| 字段 | 说明 |
|------|------|
| `blogWorkDir` | 博客 / Markdown 工作目录，默认 `~/Documents/niumer-blog` |
| `jsonFormatterWorkDir` | JSON 格式化器草稿目录，默认 `~/Documents/niumer-json-formatter`（Windows / Linux 亦为「用户目录/Documents/…」）；目录内保存 `draft.json` |
| `workHourDbPath` | 考勤 SQLite 文件路径；留空则使用下文的默认 `work_hour.db` |

可在应用内通过偏好设置修改；也可直接编辑 JSON（修改后重启或按界面逻辑重新加载）。

## Work hour（考勤同步）

1. **刷新**：Go 使用 **chromedp** 无头打开登录页获取 Cookie（流程与 `scripts/work-hour/get_work_hour.py` 中的 Playwright 示例一致），再请求业务接口并将结果写入 **SQLite**。
2. **Cookie**：仅从浏览器会话读取，**不从环境变量注入**。
3. **默认数据库**：未配置 `workHourDbPath` 时，使用用户配置目录下的 **`…/niumer/work_hour.db`**（与 `config.json` 同父目录 `niumer`）。
4. **YAML 配置**：考勤相关 URL / CSS 选择器默认写在仓库 **`configs/config.yaml`**；按环境叠加 **`configs/config.dev.yaml`** 或 **`configs/config.prod.yaml`**（由环境变量 **`NIUMER_ENV`** 选择，未设置时视为 **`dev`**）。文件在构建时嵌入二进制，修改后需重新 `wails build` / `go build`。`WORK_HOUR_*` 环境变量仍**优先于** YAML 中的值。

**可选环境变量**（覆盖 YAML 中的同含义项）：

- `NIUMER_ENV`：`dev` | `prod`，决定加载哪个 `config.<env>.yaml` 叠加层
- `WORK_HOUR_LOGIN_URL`、`WORK_HOUR_WAIT_CSS`
- `WORK_HOUR_CHROME_PATH`（Chrome/Chromium 可执行文件路径）
- `WORK_HOUR_TENANT_URL`、`WORK_HOUR_HR_ID_URL`、`WORK_HOUR_API_URL`

首次克隆或新增依赖后，在项目根执行：`make tidy` 或 `go mod tidy`。

## 生产构建

```bash
make build
```

或：

```bash
wails build
```

或自行构建前端后再编译 Go：

```bash
cd frontend && npm run build && cd ..
go build -o build/niumer .
```

嵌入资源依赖 `frontend/dist`（见 `main.go` 中的 `//go:embed all:frontend/dist`），请先完成前端 `npm run build`。

## npm 异常时的辅助脚本

若本机 `~/.npmrc` 配置不当，可能出现 **`npm` 无输出、立即失败** 的情况。可在仓库根目录使用：

```bash
./scripts/npm.sh install
./scripts/npm.sh run build
./scripts/npm.sh run dev
```

该脚本在隔离环境下调用 npm，并固定使用项目内 `frontend/.npmrc` 中的官方源。长期建议在 `~/.npmrc` 中使用合法的 `registry=...` 行（不要使用错误的 `//registry=...` 形式）。

## 目录结构

| 路径 | 说明 |
|------|------|
| `main.go` / `app.go` | Wails 入口与可绑定到前端的 Go API |
| `workhour_fetch.go` / `workhour_chromedp.go` / `workhour.go` | 考勤拉取、浏览器 Cookie、SQLite |
| `wails.json` | Wails 构建与前端脚本配置 |
| `Makefile` | `dev` / `tidy` / `build`，统一 `GOPROXY` |
| `frontend/` | Vite + React + TS 源码与 `package.json` |
| `frontend/src/components/` | VS Code 风格布局相关组件 |
| `scripts/dev.sh` | 与 `make dev` 类似：固定 `GOPROXY` 后执行 `wails dev` |
| `scripts/npm.sh` | 可选的 npm 包装脚本 |
| `scripts/work-hour/get_work_hour.py` | 可选：Playwright 对照调试（非运行时依赖） |
| `.vscode/settings.json` | 编辑器集成终端的 `GOPROXY` 等 |

## 参考

- [Wails 文档](https://wails.io/docs/introduction)
- [Vite](https://vitejs.dev/) · [Tailwind CSS](https://tailwindcss.com/)
