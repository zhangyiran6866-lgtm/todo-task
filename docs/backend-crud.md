# 后端 CRUD 开发规范

> **技术栈**：Go 1.22 + Gin + MongoDB | **参见**：`agent/backend/skill.md`

---

## 一、分层职责速查

| 层 | 职责 | 禁止 |
|----|------|------|
| Handler | 解析参数 → 调用 Service → 映射 HTTP 状态码 | 写业务逻辑、查数据库 |
| Service | 业务编排、定义业务错误 | 直接操作 MongoDB |
| Repository | 只做 MongoDB CRUD | 写业务判断逻辑 |

---

## 二、标准 CRUD 实现模板

### 2.1 新建资源（Create）

```
Handler: 绑定 JSON → 调用 svc.Create → response.OK
Service: 填充时间戳/UserID → 调用 repo.Insert → 返回新建实体
Repository: InsertOne → 返回插入结果
```

**关键点**：
- `CreatedAt`/`UpdatedAt` 由 Service 层在写入前统一赋值 `time.Now()`
- 冲突错误（如邮箱唯一约束）在 Repository 捕获并转为 `ErrConflict`

---

### 2.2 查询列表（List）

```
Handler: 解析 Query 参数（status/priority/limit/cursor）→ 调用 svc.List
Service: 校验参数合法性 → 调用 repo.FindMany
Repository: 构建 filter + 游标分页（禁止 $skip）→ Find + SetLimit
```

**游标分页约定**：
- 默认按 `created_at DESC, _id DESC` 排序
- 前端传 `cursor = 上一页最后一条 id`

---

### 2.3 查询单条（Get）

```
Handler: 解析并验证 :id 为合法 ObjectID → 调用 svc.GetByID
Service: 调用 repo.FindByID → 404 错误映射
Repository: FindOne → ErrNoDocuments → ErrTaskNotFound
```

---

### 2.4 更新资源（Update）

```
Handler: 绑定 JSON（仅更新字段）→ 调用 svc.Update
Service: 校验归属权（user_id 匹配）→ 调用 repo.UpdateByID
Repository: $set 指定字段 + updated_at → UpdateOne
```

**注意**：
- 使用 `$set` 局部更新，**禁止** `ReplaceOne`
- 必须在 Service 层校验任务归属，防止越权

---

### 2.5 删除资源（Delete）

```
Handler: 解析 :id → 调用 svc.Delete
Service: 校验归属权 → 调用 repo.SoftDelete
Repository: $set { is_deleted: true, deleted_at: now() }
```

**禁止物理删除**，统一软删除。

---

## 三、错误处理约定

```go
// Repository 层：DB 错误 → 业务错误
var (
    ErrTaskNotFound  = errors.New("task not found")
    ErrUserNotFound  = errors.New("user not found")
    ErrEmailConflict = errors.New("email already exists")
)

// Service 层：业务校验错误
var (
    ErrForbidden = errors.New("forbidden")
)

// Handler 层：业务错误 → HTTP 状态码
switch {
case errors.Is(err, service.ErrTaskNotFound): response.NotFound(...)
case errors.Is(err, service.ErrForbidden):    response.Forbidden(...)
default:                                       response.InternalError(...)
}
```

---

## 四、参数校验规范

- Handler 层使用 `c.ShouldBindJSON(&req)` 绑定参数
- 业务必填字段在 Service 层用 `if req.X == ""` 显式校验
- ObjectID 解析失败统一返回 400

---

## 五、禁止清单

| # | 禁止行为 |
|---|----------|
| 1 | Handler 层写 MongoDB 查询 |
| 2 | 忽略 error（`_, _ :=`） |
| 3 | 直接调用 `c.JSON` 响应 |
| 4 | 物理删除数据（`DeleteOne`） |
| 5 | 不携带 Context 超时的 DB 操作 |
| 6 | Service 层不校验资源归属权 |
