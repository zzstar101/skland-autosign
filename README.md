## 森空岛签到（Go 版本）

使用 Go 实现的森空岛自动签到服务，支持多账号管理和多种推送通知方式，适合 Docker、云函数、青龙面板等多种环境部署。

> 说明：原有基于 Nitro/TypeScript 的实现仍然保留在仓库中，推荐在新环境优先使用 Go 版本。

### 功能特点

- 🌟 支持多账号管理
- 🤖 一次执行/定时任务均可使用（由外部调度，如 cron、云函数触发器、青龙计划任务）
- 📱 支持多种推送通知方式（通过通用 Webhook URL）
- 🔄 支持错误自动重试

### 配置说明（环境变量）

- **`TOKENS`**：森空岛凭据，多个账号用逗号分隔（必填）  
  示例：`TOKENS=token1,token2`
- **`NOTIFICATION_URLS`**：通知 URL 列表，多个用逗号分隔（可选）  
  示例：`NOTIFICATION_URLS=https://your-webhook-url`
- **`MAX_RETRIES`**：单角色签到失败时的最大重试次数，默认 `3`（可选）  
  示例：`MAX_RETRIES=5`

凭据获取方式与原项目一致：登录森空岛或鹰角通行证，访问对应接口获取 `content` 字段值，然后填入 `TOKENS`。

### 本地运行（Go）

1. 安装 Go（建议 1.21+）。  
2. 在项目根目录执行：

```bash
cd go
go run ./cmd/skland-attendance -mode=once
```

运行前请在当前环境中设置好 `TOKENS`、`NOTIFICATION_URLS` 等变量。

### Docker 部署（Go 版本）

项目根目录的 `Dockerfile` 已改为构建并运行 Go 版本，适合一次性执行的签到任务。

#### 使用 Docker 构建并运行

```bash
docker build -t skland-attendance-go .

docker run --rm \
  -e TOKENS="your-token-1,your-token-2" \
  -e NOTIFICATION_URLS="https://your-webhook-url" \
  skland-attendance-go
```

容器退出码为 `0` 表示全部账号成功或已签到，非 `0` 表示有失败账号，可用于外部告警或重试逻辑。

#### 使用 docker-compose 定时执行

仓库中提供示例文件 `docker-compose.go-example.yml`：

- 可通过宿主机 crontab 定期执行：

```bash
0 8 * * * cd /path/to/project && docker compose -f docker-compose.go-example.yml run --rm skland-attendance
```

这样每天早上 8 点拉起一个容器执行一次签到，执行完自动退出并释放资源。

### 云函数部署（通用说明）

Go 版本提供了云函数入口示例 `go/cmd/lambda/main.go`，可用于 AWS Lambda 或其他支持 Go 的云函数平台。

通用步骤：

1. 在 `go/` 目录构建二进制并打包：

```bash
cd go
GOOS=linux GOARCH=amd64 go build -o bootstrap ./cmd/lambda
zip function.zip bootstrap
```

2. 在云函数控制台创建函数：
   - 运行时选择 Go（或使用自定义运行时，入口为打包好的 `bootstrap`）。
   - 上传 `function.zip` 作为函数代码。
   - 在环境变量中配置 `TOKENS`、`NOTIFICATION_URLS`、`MAX_RETRIES`。

3. 创建定时触发器（如每日固定时间触发一次函数）。  

函数返回的 JSON 中包含 `result` 字段和完整统计信息，方便在日志或监控中查看。

### 青龙面板部署

青龙面板可以通过两种方式使用本项目的 Go 版本：

#### 方式 A：预编译二进制（推荐）

1. 在本地或青龙容器中编译：

```bash
cd /path/to/repo/go
go build -o skland-attendance ./cmd/skland-attendance
```

2. 将生成的 `skland-attendance` 放到青龙脚本目录，例如 `/ql/data/scripts/skland-attendance`。

3. 在青龙面板中添加环境变量：
   - `TOKENS`：森空岛凭据列表。
   - `NOTIFICATION_URLS`：通知地址（可选）。
   - `MAX_RETRIES`：最大重试次数（可选）。

4. 在青龙中新增定时任务，命令示例：

```bash
cd /ql/data/scripts/skland-attendance && ./skland-attendance -mode=once
```

设置合适的 cron 表达式（例如每天 8 点执行一次）。

#### 方式 B：源码直接运行

1. 在青龙容器中克隆仓库：

```bash
cd /ql/data/scripts
git clone https://github.com/yourname/skland-daily-attendance.git
```

2. 在青龙任务中使用：

```bash
cd /ql/data/scripts/skland-daily-attendance/go && go run ./cmd/skland-attendance -mode=once
```

此方式每次执行会编译一次，耗时略长，适合调试或临时使用，长期推荐方式 A。

### HTTP 服务模式（兼容旧行为）

如需通过 HTTP 调用的方式触发签到（类似原 Nitro `/_nitro/tasks/attendance`），可以使用 `http` 模式：

```bash
cd go
go run ./cmd/skland-attendance -mode=http -addr=":8080"
```

然后通过 `GET http://your-host:8080/attendance` 触发签到，返回结果 JSON 中包含：

```json
{
  "result": "success | failed",
  "stats": { ... }
}
```

你也可以在自建服务器或其他平台上将此服务部署为常驻进程，然后用外部定时任务访问该 HTTP 接口。

### 注意事项

- 本项目仅用于学习和研究目的，请合理使用，避免频繁调用 API 影响账号安全。
- Go 版本默认使用内存存储“今日已签到”状态，适合一次执行的任务。例如 Docker 一次性容器、云函数、青龙任务等。

### License

MIT
