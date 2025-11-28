# ZKTeco Attendance Shifts 报表服务

一个面向 ZKTeco 考勤系统（SQL Server）的轻量级 Web 服务，用于按月汇总员工的出勤与加班时长，并提供浏览器查看和 CSV 导出功能。

## 功能特性
- 按当前月份生成考勤报表（每日“上班/加班”时长）
- 浏览器查看简洁表格页面：`/`
- 一键下载月度 CSV 报表：`/download`
- 读取 SQL Server 中 `userinfo`、`departments`、`attshifts` 三张表
- 可配置数据库连接与 HTTP 端口

## 目录结构
```
cmd/
  attshifts/  程序入口（main.go）
internal/
  config/     配置读取
  db/         数据库连接
  service/    业务查询（用户、考勤）
  web/        HTTP 路由与页面/导出
scripts/
  go-run.bat  运行脚本（Windows）
  go-build.bat 构建脚本（Windows）
example/
  config.example.json 示例配置
```

- 程序入口：`cmd/attshifts/main.go`
- 路由注册：`internal/web/web.go` 的 `RegisterRoutes()`
- 业务查询：`internal/service/service.go` 的 `QueryUsers()`、`QueryAtt()`
- 数据库连接：`internal/db/db.go` 的 `Init()`、`Get()`、`Close()`
- 配置读取：`internal/config/config.go` 的 `Load()`

## 运行环境
- Go ≥ 1.24（go.mod: `go 1.24.5`）
- SQL Server（兼容 ZKTeco 原始库结构）
- 网络可访问数据库实例

## 安装与快速开始
1. 安装 Go（确保 `go version` ≥ 1.24）
2. 克隆项目到本地
3. 配置数据库连接（见下文“配置文件”）
4. 运行服务：
   - Windows：
     - 在项目根目录双击 `scripts/go-run.bat`，或执行：
       ```bat
       go run ./cmd/attshifts
       ```
   - 其他平台：
       ```sh
       go run ./cmd/attshifts
       ```
5. 浏览器访问 `http://127.0.0.1:8080/`（端口可通过配置调整）

## 配置文件
程序在启动时读取 `config.json`，优先使用当前工作目录下的文件，若不存在则回退到可执行文件同目录。

示例（请勿提交真实凭据到仓库）：
```json
{
  "server": "127.0.0.1",
  "port": 1433,
  "user": "readonly",
  "password": "<your-password>",
  "database": "zkeco",
  "http_port": 8080
}
```

字段说明：
- `server` / `port`：SQL Server 地址与端口
- `user` / `password`：数据库账号与密码（建议只读权限）
- `database`：ZKTeco 数据库名（常见为 `zkeco`）
- `http_port`：HTTP 服务端口（空或 0 时默认 8080）

## HTTP 接口
- `GET /`：当前月份的考勤汇总页面（按部门排序，显示工号、姓名、部门与每日“上/加”时长）
- `GET /download`：下载当前月份 CSV 文件，列包含“部门/工号/姓名”及每日的“上班/加班”聚合

## 数据来源与聚合逻辑
- 员工与部门：
  - `userinfo`（过滤 `deltag=0`）
  - `departments`（左连接获取部门名）
- 日考勤聚合：
  - 表：`attshifts`
  - 时间范围：当月 1 日至当月末
  - 字段聚合：`SUM(realworkday) AS work`、`SUM(overtime) AS over`
  - 分组：`userid, attdate`

## 构建与部署
- 本地构建（Windows）：
  ```bat
  scripts\go-build.bat
  ```
- 跨平台构建（示例）：
  ```sh
  go build ./cmd/attshifts -o attshifts
  ```
- 部署时将 `config.json` 放置在可执行文件同目录，或在工作目录提供该文件。

## 常见问题
- 无法连接数据库：
  - 检查 `server/port/user/password` 是否正确
  - 确保 SQL Server 对当前主机与端口开放访问
  - 若开启加密，请调整连接字符串（目前默认 `encrypt=disable`）
- 页面为空或数据缺失：
  - 确认当月是否存在 `attshifts` 数据
  - 确认 `userinfo` 的有效员工（`deltag=0`）

## 安全与合规
- 切勿将包含真实凭据的 `config.json` 提交到公共仓库
- 建议在生产中使用只读数据库账号
- 如需更严格的安全策略（例如加密连接），可扩展 `internal/db/db.go` 的连接字符串配置

## 许可证
本项目建议使用 MIT 许可证；如需修改，请在仓库根目录添加 `LICENSE` 文件并更新本节说明。

## 致谢
- MSSQL 驱动：`github.com/microsoft/go-mssqldb`

---

欢迎提交 Issue 与 PR，贡献特性包括：日期范围选择、部门/员工筛选、导出为 Excel、国际化与主题样式等。
