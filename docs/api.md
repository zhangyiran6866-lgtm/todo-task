# 接口文档（API Reference）

> **Base URL**：`http://localhost:8080/api/v1`
> **认证方式**：`Authorization: Bearer <access_token>`
> **响应格式**：`{ "code": 0, "message": "成功", "data": {} }`

---

## 一、认证模块 `/auth`

### 1.1 注册

```
POST /auth/register
```

**Request Body**

```json
{
  "email": "user@example.com",
  "password": "Abc123456",
  "nickname": "张三"
}
```

**Response**

```json
{
  "code": 0,
  "message": "成功",
  "data": {
    "access_token": "eyJ...",
    "refresh_token": "eyJ..."
  }
}
```

---

### 1.2 登录

```
POST /auth/login
```

**Request Body**

```json
{
  "email": "user@example.com",
  "password": "Abc123456"
}
```

**Response** — 同注册

---

### 1.3 刷新 Token

```
POST /auth/refresh
```

**Request Body**

```json
{ "refresh_token": "eyJ..." }
```

**Response** — 返回新的 `access_token`

---

### 1.4 退出登录

```
POST /auth/logout
Authorization: Bearer <access_token>
```

**Request Body**

```json
{ "refresh_token": "eyJ..." }
```

将 refresh_token 加入黑名单，**Response** `data: null`

---

## 二、用户模块 `/users` 🔒

> 以下接口均需要 `Authorization` Header

### 2.1 获取当前用户信息

```
GET /users/me
```

**Response**

```json
{
  "code": 0,
  "data": {
    "id": "...",
    "email": "user@example.com",
    "nickname": "张三",
    "language": "zh",
    "theme": "cyan",
    "created_at": "2024-01-01T00:00:00Z"
  }
}
```

---

### 2.2 更新用户信息

```
PATCH /users/me
```

**Request Body**（字段均可选）

```json
{
  "nickname": "李四",
  "language": "en",
  "theme": "purple"
}
```

**Response**

```json
{
  "code": 0,
  "message": "成功",
  "data": {
    "id": "...",
    "email": "user@example.com",
    "nickname": "李四",
    "language": "en",
    "theme": "purple",
    "created_at": "2024-01-01T00:00:00Z",
    "updated_at": "2024-01-02T00:00:00Z"
  }
}
```

> 若请求体为空（未传任何可更新字段），返回 `400`。

---

### 2.3 修改密码

```
PUT /users/me/password
```

**Request Body**

```json
{
  "old_password": "Abc123456",
  "new_password": "NewPass789"
}
```

---

## 三、任务模块 `/tasks` 🔒

### 3.1 创建任务

```
POST /tasks
```

**Request Body**

```json
{
  "title": "完成接口文档",
  "description": "编写所有 API 接口说明",
  "priority": "critical",
  "due_at": "2024-12-31T23:59:59Z"
}
```

`priority` 可选值：`critical` / `important` / `urgent` / `routine` / `low`

---

### 3.2 获取任务列表

```
GET /tasks?status=todo&priority=urgent&limit=20&cursor=<last_id>
```

**Query 参数**

| 参数 | 类型 | 说明 |
|------|------|------|
| `status` | string | 可选，过滤状态 |
| `priority` | string | 可选，过滤优先级（`critical`/`important`/`urgent`/`routine`/`low`） |
| `limit` | int | 默认 20，最大 50 |
| `cursor` | string | 游标分页，上一页最后一条 `id` |

**Response**

```json
{
  "code": 0,
  "data": {
    "items": [...],
    "next_cursor": "..."
  }
}
```

---

### 3.3 获取单个任务

```
GET /tasks/:id
```

---

### 3.4 更新任务

```
PATCH /tasks/:id
```

**Request Body**（字段均可选）

```json
{
  "title": "新标题",
  "status": "in_progress",
  "priority": "important"
}
```

---

### 3.5 删除任务（软删除）

```
DELETE /tasks/:id
```

**Response** `data: null`

---

## 四、日志模块 `/logs` 🔒

> 以下接口均需要 `Authorization` Header

### 4.1 查询日志列表

```
GET /logs?channel=app&level=error&module=task&keyword=timeout&start_at=2026-04-27T00:00:00Z&end_at=2026-04-27T23:59:59Z&page=1&page_size=20
```

**Query 参数**

| 参数 | 类型 | 说明 |
|------|------|------|
| `channel` | string | 可选，日志通道：`app`/`error`/`audit` |
| `level` | string | 可选，日志级别：`debug`/`info`/`warn`/`error` |
| `module` | string | 可选，模块过滤（如 `auth`、`task`） |
| `keyword` | string | 可选，关键词模糊检索 |
| `start_at` | string | 可选，开始时间（RFC3339 或 `yyyy-mm-dd`） |
| `end_at` | string | 可选，结束时间（RFC3339 或 `yyyy-mm-dd`） |
| `page` | int | 页码，从 1 开始，默认 1 |
| `page_size` | int | 每页条数，默认 20，最大 100 |
| `limit` | int | 兼容字段，等价 `page_size` |
| `cursor` | string | 兼容字段，偏移游标分页（不推荐） |

**Response**

```json
{
  "code": 0,
  "message": "成功",
  "data": {
    "items": [
      {
        "id": "2fa2c1...",
        "channel": "app",
        "timestamp": "2026-04-27T10:22:11.120+08:00",
        "level": "info",
        "module": "task",
        "action": "list_tasks",
        "message": "http request completed",
        "request_id": "d6f84d2a..."
      }
    ],
    "pagination": {
      "total": 126,
      "page": 1,
      "page_size": 20,
      "total_pages": 7,
      "has_next": true,
      "has_prev": false
    }
  }
}
```

---

### 4.2 获取日志详情

```
GET /logs/:id?channel=app
```

**说明**

- `id` 来自日志列表 `items[].id`
- `channel` 可选，传入后可缩小检索范围并提升查询速度

---

## 五、错误码速查

| code | HTTP Status | 含义 |
|------|------------|------|
| 0 | 200 | 成功 |
| 400 | 400 | 请求参数错误 |
| 401 | 401 | 未认证或 Token 失效 |
| 403 | 403 | 无权限 |
| 404 | 404 | 资源不存在 |
| 409 | 409 | 资源冲突（如邮箱已注册） |
| 500 | 500 | 服务器内部错误 |
