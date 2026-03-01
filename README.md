# Webpage Cache（高并发网页快照系统）

## 项目简介

Webpage Cache 是一个基于 Go 构建的生产级高并发网页截图系统。

本项目采用异步任务架构，结合 Redis 任务队列、Worker Pool
并发控制机制以及浏览器实例池优化策略，实现高吞吐、可扩展、可恢复的截图服务。

------------------------------------------------------------------------

## API 接口说明

### 1. 创建截图任务

- 方法：`POST /screenshot`
- Content-Type：`application/json`

请求体：

```json
{
  "url": "https://developer.mozilla.org/zh-CN/"
}
```

成功响应（202）：

```json
{
  "request_id": "f7d7c2f5-3a16-4f13-906e-0d70b8a7d8b4",
  "biz_code": "ACCEPTED",
  "message": "task accepted",
  "data": {
    "task_id": "6bb9b5e3-3d90-4d59-9f99-cd3af3a9b8c8",
    "status": "pending"
  }
}
```

常见失败：

- `400` + `biz_code=INVALID_REQUEST`：请求体不合法
- `400` + `biz_code=INVALID_URL`：URL 非法（仅支持 `http/https`）
- `500` + `biz_code=INTERNAL_ERROR`：服务内部错误

---

### 2. 查询任务状态

- 方法：`GET /screenshot/:id`

成功响应（200）：

```json
{
  "request_id": "6be7f45f-3142-43b3-a3dc-2ef1f730cf90",
  "biz_code": "OK",
  "message": "ok",
  "data": {
    "id": "6bb9b5e3-3d90-4d59-9f99-cd3af3a9b8c8",
    "url": "https://developer.mozilla.org/zh-CN/",
    "status": "done",
    "result_url": "/static/screenshots/6bb9b5e3-3d90-4d59-9f99-cd3af3a9b8c8.png",
    "error_msg": "",
    "retry_count": 0,
    "created_at": "2026-02-28T08:00:00Z",
    "updated_at": "2026-02-28T08:00:05Z"
  }
}
```

常见失败：

- `400` + `biz_code=INVALID_TASK_ID`：`id` 不是合法 UUID
- `404` + `biz_code=TASK_NOT_FOUND`：任务不存在

---

### 3. 任务状态说明

- `pending`：已入队，等待 worker 处理
- `processing`：worker 正在处理
- `done`：截图成功
- `failed`：截图失败（可查看 `error_msg`）

---

### 4. 监控接口

- 方法：`GET /metrics`
- 用途：Prometheus 拉取服务指标（请求量、耗时、任务处理指标等）


------------------------------------------------------------------------

## 系统架构

### 总体架构图

``` mermaid
flowchart TD
    Client --> API_Server
    API_Server --> Redis_Queue
    Redis_Queue --> Worker_Cluster
    Worker_Cluster --> Browser_Pool
    Worker_Cluster --> MySQL
    Worker_Cluster --> Object_Storage
```
[详细说明](./架构设计.md)

------------------------------------------------------------------------

## 核心架构设计

+ 异步任务模型
+ Redis 队列解耦

+ Worker Pool 并发控制

+ Chrome 进程池优化

+ 任务状态机设计

+ 支持水平扩展
------------------------------------------------------------------------

## 项目结构

    cmd/
        server/
    internal/
        api/
        service/
        repository/
        worker/
        queue/
        browser/
        model/
    config/
    deployments/

------------------------------------------------------------------------

## 代理池接入

为应对部分网站反爬限制，系统已支持“代理池 + 重试换代理”。

### 工作机制

- 通过 `PROXY_URLS` 配置多个代理地址（逗号分隔）
- Browser Pool 初始化实例时按轮询分配代理
- 任务失败重试时，会优先避开上一次使用的代理

### 关键配置

- `PROXY_URLS`：代理地址列表
- `BROWSER_POOL_SIZE`：Chrome 实例数
- `MAX_TABS_PER_BROWSER`：每实例并发 tab 数

总截图并发上限约为：

`BROWSER_POOL_SIZE * MAX_TABS_PER_BROWSER`

### 配置示例

```bash
PROXY_URLS=http://user:pass@proxy1:8080,http://user:pass@proxy2:8080
BROWSER_POOL_SIZE=2
MAX_TABS_PER_BROWSER=3
```

如果不需要代理，将 `PROXY_URLS` 置空即可（默认直连）。

------------------------------------------------------------------------


## 技术栈

-   Go
-   Gin
-   Redis
-   MySQL
-   chromedp

------------------------------------------------------------------------

## 部署

### 本地运行

``` bash
go mod tidy
go run cmd/server/main.go
```

### Docker部署

[README_Docker](./deployments/README_Docker.md)


------------------------------------------------------------------------

## 未来优化方向
待补充

------------------------------------------------------------------------

## License

MIT
