# Black Box Game

一个基于 Go 的小型 Web 解谜游戏。玩家通过交换柱子的位置，在看不到真实高度的前提下，只依赖“当前蓄水量”的反馈，尝试把局面调整到理论最优解。

项目现在已经拆成了比较清晰的分层结构，适合继续往里加新玩法、规则、存储方式和前端交互。

## 核心玩法

- 每局开始后，后端会随机生成一组柱子高度。
- 前端默认不展示柱子的真实高度，只显示当前积水量。
- 玩家每次选择两根柱子进行交换，会消耗 1 步。
- 如果当前排列达到该关的理论最大积水量，则本关通关。
- 通关后进入下一关，并获得额外步数奖励。
- 如果步数耗尽，游戏结束，成绩会写入排行榜。

## 当前特性

- Go 原生 `net/http` 提供 Web 服务
- 前后端分离但部署简单，静态资源直接由后端托管
- 基于 Cookie 的会话管理
- 本地 JSON 排行榜持久化
- 复古街机风格前端界面

## 项目结构

```text
blackbox-game/
├─ cmd/
│  └─ blackbox-game/
│     └─ main.go              # 程序入口
├─ internal/
│  ├─ cli/
│  │  └─ renderer.go          # 旧 CLI 渲染代码，当前 Web 主流程未使用
│  ├─ game/
│  │  ├─ api.go               # game 包对外暴露的薄封装
│  │  ├─ engine.go            # 核心算法：积水计算、目标分求解
│  │  └─ session.go           # 游戏状态、会话、关卡初始化
│  ├─ store/
│  │  └─ leaderboard.go       # 排行榜读取与写入
│  └─ web/
│     └─ server.go            # HTTP 路由与接口处理
├─ static/
│  ├─ index.html              # 页面骨架
│  ├─ game.js                 # 前端交互逻辑
│  └─ style.css               # 前端样式
├─ leaderboard.json           # 本地排行榜数据
├─ go.mod
└─ README.md
```

## 运行方式

### 1. 直接运行

```powershell
go run ./cmd/blackbox-game
```

启动后默认监听：

- 本机地址：`http://localhost:8080`
- 局域网地址：程序会在终端里自动打印可访问 IP

### 2. 构建可执行文件

```powershell
go build -o blackbox-game.exe ./cmd/blackbox-game
```

然后直接运行：

```powershell
.\blackbox-game.exe
```

## 主要接口

当前后端提供的接口比较简单，适合后续继续扩展：

- `GET /`
  - 返回游戏首页
- `POST /api/start`
  - 开始新游戏
- `POST /api/swap`
  - 交换两个柱子的位置
- `POST /api/next`
  - 进入下一关
- `POST /api/quit`
  - 主动退出并结算成绩
- `GET /api/leaderboard`
  - 获取排行榜

## 数据与状态

### 会话

- 会话保存在内存中
- 使用 `session_id` Cookie 区分玩家
- 服务重启后，会话不会保留

### 排行榜

- 排行榜保存在根目录的 `leaderboard.json`
- 当前只保留 Top 5
- 同名玩家会刷新自己的最高记录

## 开发建议

如果你后面准备继续加功能，推荐按下面的边界继续扩展：

- 加新规则或新关卡逻辑：优先放到 `internal/game`
- 加新接口或页面流程：优先放到 `internal/web`
- 换排行榜存储方式，比如 SQLite、Redis、远端服务：优先改 `internal/store`
- 改 UI、动画、交互细节：放到 `static`

## 适合的下一步演进

这个项目很适合继续往下面几个方向走：

- 给 `internal/game` 增加单元测试，先覆盖 `trap`、关卡初始化和通关判定
- 把目标分求解算法从暴力搜索优化掉，避免高关卡时性能抖动
- 在 `internal/web` 里继续拆出 service 层，减少 handler 直接操作状态
- 给排行榜增加更多字段，比如总局数、总交换次数、最快通关时间
- 增加更多玩法元素，比如障碍柱、技能、每日挑战、种子关卡

## 开发环境

- Go `1.25.1`
- 前端为原生 HTML / CSS / JavaScript
- 不依赖额外数据库或前端构建工具

## 备注

- 当前仓库中保留了一个旧的 `internal/cli/renderer.go`，它更多是历史代码，不在现在的 Web 主流程里。
- 当前排行榜是本地文件方案，适合单机试玩或局域网演示；如果以后要长期运行，建议把存储层独立出来。
