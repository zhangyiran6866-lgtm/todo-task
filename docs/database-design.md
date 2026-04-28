# 数据库设计文档

> **数据库**：MongoDB | **规范**：参见 `agent/database/skill.md`

---

## 一、集合规划总览

| 集合名 | 用途 | 状态 |
|--------|------|------|
| `users` | 用户账号信息 | ✅ 已规划 |
| `tasks` | 任务数据 | ✅ 已规划 |
| `token_blacklist` | JWT refresh token 黑名单 | ✅ 已规划 |

---

## 二、集合详细设计

### 2.1 `users` — 用户表

```go
type User struct {
    ID        primitive.ObjectID `bson:"_id,omitempty" json:"id"`
    Email     string             `bson:"email"         json:"email"`
    Password  string             `bson:"password"      json:"-"`        // bcrypt 加密
    Nickname  string             `bson:"nickname"      json:"nickname"`
    Language  string             `bson:"language"      json:"language"` // "zh" | "en"
    Theme     string             `bson:"theme"         json:"theme"`    // "cyan"|"purple"|"green"|"pink"
    IsDeleted bool               `bson:"is_deleted"    json:"-"`
    CreatedAt time.Time          `bson:"created_at"    json:"created_at"`
    UpdatedAt time.Time          `bson:"updated_at"    json:"updated_at"`
}
```

**索引**：
- `email`：唯一索引
- `created_at`：倒序索引

---

### 2.2 `tasks` — 任务表

```go
type Task struct {
    ID          primitive.ObjectID `bson:"_id,omitempty"  json:"id"`
    UserID      primitive.ObjectID `bson:"user_id"        json:"user_id"`
    Title       string             `bson:"title"          json:"title"`
    Description string             `bson:"description"    json:"description"`
    Status      string             `bson:"status"         json:"status"`      // "todo"|"in_progress"|"done"
    Priority    string             `bson:"priority"       json:"priority"`    // "critical"|"important"|"urgent"|"routine"|"low"
    DueAt       *time.Time         `bson:"due_at"         json:"due_at"`
    IsDeleted   bool               `bson:"is_deleted"     json:"-"`
    DeletedAt   *time.Time         `bson:"deleted_at"     json:"-"`
    CreatedAt   time.Time          `bson:"created_at"     json:"created_at"`
    UpdatedAt   time.Time          `bson:"updated_at"     json:"updated_at"`
}
```

**索引**：
- `{user_id: 1, is_deleted: 1, created_at: -1}`：复合索引（ESR 规则）
- `{user_id: 1, status: 1}`：状态筛选索引

---

### 2.3 `token_blacklist` — Token 黑名单表

```go
type TokenBlacklist struct {
    ID        primitive.ObjectID `bson:"_id,omitempty" json:"id"`
    Token     string             `bson:"token"         json:"-"`
    ExpiredAt time.Time          `bson:"expired_at"    json:"-"` // TTL 索引自动清理
    CreatedAt time.Time          `bson:"created_at"    json:"-"`
}
```

**索引**：
- `token`：唯一索引
- `expired_at`：TTL 索引（自动过期删除）

---

## 三、软删除规范

所有业务数据禁止物理删除，统一执行软删除：

```js
// 删除操作
{ $set: { is_deleted: true, deleted_at: new Date() } }

// 所有查询默认过滤
{ is_deleted: { $ne: true } }
```

---

## 四、状态枚举

| 字段 | 可选值 |
|------|--------|
| `tasks.status` | `todo` / `in_progress` / `done` |
| `tasks.priority` | `critical` / `important` / `urgent` / `routine` / `low` |
| `users.language` | `zh` / `en` |
| `users.theme` | `cyan` / `purple` / `green` / `pink` |
