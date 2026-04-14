# MongoDB 数据库操作规范 · Skill

> **适用范围**：所有涉及 MongoDB 的操作，包括 `packages/backend/internal/repository/` 和 `packages/backend/internal/model/`
> **权威来源**：MongoDB 官方文档 · MongoDB Go Driver 最佳实践 · 行业通行规范
> **强制级别**：AI 编写任何数据库相关代码前，必须先加载并严格遵守本文件。

---

## 一、核心原则

**设计 Schema 的第一问题是：应用怎么读写数据，而不是数据长什么样。**

MongoDB 不是关系型数据库，不要用关系型思维设计文档模型。

---

## 二、Schema 设计规范

### 2.1 嵌入（Embed）vs 引用（Reference）决策

| 维度 | 选嵌入 | 选引用 |
|------|--------|--------|
| 访问模式 | 父子数据总是一起读取 | 子数据经常独立读取 |
| 数量关系 | 一对少（1:few，如用户地址 3-5 条） | 一对多/多对多（数量不确定或无上限） |
| 更新方式 | 父子一起原子更新 | 子数据频繁独立变更 |
| 数据共享 | 不共享 | 被多处引用（如文章分类） |
| 读写比 | 读多写少（嵌入加速读） | 写多读少（引用减少冗余） |

**默认策略：从嵌入开始，遇到以下情况再改为引用：**
- 数组可能无限增长（日志、评论流、订单条目）
- 同一数据被多个父文档引用
- 单文档超过 **2MB**（硬限制 16MB，建议实际控制在 2MB 以内）

### 2.2 禁止的 Schema 反模式

```
❌ 无限增长数组（Unbounded Array）
   user.log_ids = [id1, id2, id3, ...10万条]
   → 文档膨胀，读写都慢

❌ 超深嵌套（Deep Nesting）
   order.detail.item.variant.sku.price
   → 查询困难，索引无法下推到深层字段

❌ 多态字段（Polymorphic Field）
   field: string | array | object  // 类型不固定
   → $type 过滤才能区分，应用层处理复杂

❌ 用字段名编码数据（Field-as-Value）
   { "2024-01": 100, "2024-02": 200 }  // 月份作为 key
   → 无法索引，无法查询范围
```

### 2.3 典型设计模式（本项目适用）

**计算模式（Computed Pattern）**：预存统计值，避免实时计算

```go
// ✅ 博客文章预存阅读数，而非每次 count()
type Post struct {
    ID        primitive.ObjectID `bson:"_id"`
    Title     string             `bson:"title"`
    ViewCount int64              `bson:"view_count"` // 写入时 $inc 递增
}
```

**扩展引用模式（Extended Reference）**：引用时冗余高频字段

```go
// ✅ 评论中冗余作者 name，避免每次 $lookup users
type Comment struct {
    ID         primitive.ObjectID `bson:"_id"`
    PostID     primitive.ObjectID `bson:"post_id"`
    AuthorID   primitive.ObjectID `bson:"author_id"`
    AuthorName string             `bson:"author_name"` // 冗余，加速读
    Content    string             `bson:"content"`
}
```

---

## 三、Model 定义规范

### 3.1 Struct Tag 规范

```go
// ✅ 完整示例
type User struct {
    ID        primitive.ObjectID `bson:"_id,omitempty"  json:"id"`
    Email     string             `bson:"email"          json:"email"           validate:"required,email"`
    Password  string             `bson:"password"       json:"-"`              // 禁止暴露给前端
    Name      string             `bson:"name"           json:"name"`
    Role      string             `bson:"role"           json:"role"`
    IsActive  bool               `bson:"is_active"      json:"is_active"`
    CreatedAt time.Time          `bson:"created_at"     json:"created_at"`
    UpdatedAt time.Time          `bson:"updated_at"     json:"updated_at"`
}
```

**Tag 规则**：
- `bson` 和 `json` tag **必须同时声明**，字段名用 `snake_case`
- `_id` 字段必须加 `omitempty`，防止插入零值 ObjectID
- 敏感字段（密码、Token）的 `json` tag 必须为 `"-"`
- 时间字段统一用 `time.Time`，**禁止**用 `int64` Unix 时间戳
- 引用其他集合的字段类型用 `primitive.ObjectID`，tag 为 `bson:"xxx_id"`

### 3.2 索引声明（在 Repository 初始化时创建）

```go
// ✅ 在 Repository 初始化时，通过代码创建索引
func NewUserRepository(db *mongo.Database) *mongoUserRepository {
    col := db.Collection("users")

    // email 唯一索引
    col.Indexes().CreateOne(context.Background(), mongo.IndexModel{
        Keys:    bson.D{{Key: "email", Value: 1}},
        Options: options.Index().SetUnique(true),
    })

    // created_at 倒序索引（分页查询用）
    col.Indexes().CreateOne(context.Background(), mongo.IndexModel{
        Keys: bson.D{{Key: "created_at", Value: -1}},
    })

    return &mongoUserRepository{col: col}
}
```

---

## 四、索引设计规范

### 4.1 ESR 规则（复合索引字段排序）

**Equality → Sort → Range**，依次排列：

```go
// 查询：status = "published" AND created_at > lastTime ORDER BY created_at DESC
// ✅ 正确复合索引顺序：先等值(status)，再排序(created_at)，范围(created_at)与排序重合

bson.D{
    {Key: "status",     Value: 1},   // Equality：等值匹配
    {Key: "created_at", Value: -1},  // Sort + Range：排序兼范围
}
```

### 4.2 验证索引是否生效

```go
// ✅ 使用 explain 确认走索引（开发/调试阶段使用）
// 在 MongoDB Shell 中执行：
// db.posts.find({status: "published"}).explain("executionStats")
// 关注：
//   "stage": "IXSCAN"  ← 走索引 ✅
//   "stage": "COLLSCAN" ← 全表扫描 ❌
//   totalDocsExamined vs nReturned：差距越大，索引越低效
```

### 4.3 索引使用原则

```
✅ 为所有频繁查询的过滤字段建立索引
✅ 使用 explain("executionStats") 验证实际执行计划
✅ 定期用 $indexStats 识别并删除未使用的索引
✅ 大集合上的写频繁字段，评估索引对写入的开销
❌ 不要过度索引（每个索引都会拖慢写入速度并占用内存）
❌ 不要对低基数字段（如 bool / 少数枚举值）单独建索引
```

---

## 五、查询规范

### 5.1 Client 管理规范

```go
// ✅ 全局单例，禁止每次请求创建新连接
var globalClient *mongo.Client

func InitMongoDB(uri string) {
    ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
    defer cancel()

    client, err := mongo.Connect(ctx, options.Client().ApplyURI(uri).
        SetMaxPoolSize(20).    // 根据并发 goroutine 数调整
        SetMinPoolSize(5))     // 保持最小连接数，避免冷启动
    if err != nil {
        panic(fmt.Errorf("mongodb connect failed: %w", err))
    }
    globalClient = client
}
```

### 5.2 Context 超时规范

```go
// ✅ 所有 DB 操作必须携带超时 Context
func (r *mongoUserRepo) FindByID(ctx context.Context, id primitive.ObjectID) (*model.User, error) {
    // 如果调用方没有设置超时，在 repository 层加一层保险
    ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
    defer cancel()

    var user model.User
    err := r.col.FindOne(ctx, bson.M{"_id": id}).Decode(&user)
    if errors.Is(err, mongo.ErrNoDocuments) {
        return nil, ErrUserNotFound
    }
    if err != nil {
        return nil, fmt.Errorf("repository.FindByID: %w", err)
    }
    return &user, nil
}
```

### 5.3 投影（Projection）规范

```go
// ✅ 只查需要的字段，减少网络传输和内存消耗
opts := options.FindOne().SetProjection(bson.M{
    "email":      1,
    "name":       1,
    "created_at": 1,
    "password":   0, // 明确排除敏感字段（双重保险）
})
r.col.FindOne(ctx, filter, opts)

// ❌ 禁止：查询全部字段后在应用层过滤
```

### 5.4 分页规范（游标分页 vs offset 分页）

```go
// ❌ 禁止：$skip 深分页，数据量大时指数级变慢
opts := options.Find().SetSkip(int64(page * pageSize)).SetLimit(int64(pageSize))

// ✅ 正确：游标分页（基于最后一条记录的锚点字段翻页）
// 前端传入上一页最后一条记录的 created_at 和 _id
filter := bson.M{
    "status": "published",
    "$or": bson.A{
        bson.M{"created_at": bson.M{"$lt": lastCreatedAt}},
        bson.M{
            "created_at": lastCreatedAt,
            "_id":        bson.M{"$lt": lastID},
        },
    },
}
opts := options.Find().SetSort(bson.D{
    {Key: "created_at", Value: -1},
    {Key: "_id", Value: -1},
}).SetLimit(int64(pageSize))
```

> **规则**：前 10 页可以接受 `$skip`；超过 10 页或数据量超过 10 万条，必须改用游标分页。

### 5.5 批量操作规范

```go
// ✅ 批量插入使用 InsertMany，禁止循环单条 Insert
docs := make([]interface{}, len(items))
for i, item := range items {
    docs[i] = item
}
r.col.InsertMany(ctx, docs)

// ✅ 批量更新使用 BulkWrite
models := []mongo.WriteModel{
    mongo.NewUpdateOneModel().
        SetFilter(bson.M{"_id": id1}).
        SetUpdate(bson.M{"$set": bson.M{"status": "done"}}),
    mongo.NewUpdateOneModel().
        SetFilter(bson.M{"_id": id2}).
        SetUpdate(bson.M{"$set": bson.M{"status": "done"}}),
}
r.col.BulkWrite(ctx, models, options.BulkWrite().SetOrdered(false))
```

---

## 六、聚合管道规范（⚠️ 高风险区域）

### 🔴 核心原则

> **能用 `find()` 解决的，就不要写 `aggregate()`；能分步做的，就不要一步聚合。**

聚合管道功能强大，但极易造成严重性能问题。**在没有充分理解其执行机制和数据规模影响前，必须优先使用：简单查询 + 应用层处理** 替代聚合管道。

聚合管道只在以下情况允许使用：
1. 明确知道集合的数据量级（< 10 万条无索引，或 > 10 万条但有合适索引覆盖）
2. 已通过 `explain("executionStats")` 验证执行计划
3. 简单查询 + 应用层处理确实无法满足需求

---

### 6.1 常见性能陷阱（必须规避）

| # | 问题 | 说明 |
|---|------|------|
| 1 | **全表扫描 COLLSCAN** | `$match` 条件字段无索引支持，会扫描整个集合，CPU 和内存飙升 |
| 2 | **内存限制 100MB** | 单个聚合操作默认最多使用 100MB 内存，超限报错：`$group or $sort exceeds memory limit` |
| 3 | **缺少索引导致排序慢** | `$sort` 阶段若无法使用索引，会在内存或磁盘中进行大排序 |
| 4 | **`$lookup` 变相实现 JOIN** | 在大数据量下等价于嵌套循环，性能极差；外联的目标集合必须在 join 字段上有索引 |
| 5 | **深度嵌套与多阶段处理** | 多个 `$project`、`$addFields` 层层传递，增加 CPU 开销 |
| 6 | **分页使用 `$skip` + `$limit`** | 深分页时跳过大量文档，响应时间指数级上升 |

---

### 6.2 聚合管道正确使用规范

| 规范 | 说明 |
|------|------|
| ✅ **始终以 `$match` 开头** | 确保匹配字段有合适索引（如 `task_id`、`status`）；过滤越早文档越少 |
| ✅ **避免在大集合上做 `$group`** | 改为定时预计算 + 结果缓存（存入单独集合） |
| ✅ **控制返回字段** | 紧随 `$match` 之后使用 `$project`，及早剔除无用字段，减少后续阶段数据量 |
| ✅ **启用 `allowDiskUse: true`** | 仅在必要时允许落盘（仍会影响性能，不是银弹） |
| ✅ **用游标分页替代 `$skip`** | 示例：`{ created_at: { $gt: last }, _id: { $gt: last_id } }` |
| ✅ **先查 ID 再取详情** | 聚合只取 ID 列表，再用 `$in` 批量查详情，分两步减少中间数据量 |

---

### 6.3 聚合管道代码规范

```go
// ✅ 正确示例：统计各状态文章数量
pipeline := mongo.Pipeline{
    // 第一步：$match 开头，走索引，过滤掉无关数据
    {{Key: "$match", Value: bson.M{
        "author_id": authorID,
        "is_deleted": false,
    }}},
    // 第二步：$project 及早剔除无用字段
    {{Key: "$project", Value: bson.M{
        "status": 1,
    }}},
    // 第三步：$group 统计
    {{Key: "$group", Value: bson.M{
        "_id":   "$status",
        "count": bson.M{"$sum": 1},
    }}},
}

opts := options.Aggregate().SetAllowDiskUse(true)
cursor, err := r.col.Aggregate(ctx, pipeline, opts)
```

```go
// ❌ 错误示例：$match 在 $group 后面（无法走索引）
pipeline := mongo.Pipeline{
    {{Key: "$group", Value: bson.M{"_id": "$status", "count": bson.M{"$sum": 1}}}},
    {{Key: "$match", Value: bson.M{"_id": "published"}}}, // 太晚了，已全表扫描
}
```

---

## 七、写操作规范

### 7.1 更新操作

```go
// ✅ 使用 $set 更新指定字段，禁止用 ReplaceOne 覆盖整个文档（除非明确需要）
update := bson.M{
    "$set": bson.M{
        "name":       req.Name,
        "updated_at": time.Now(),
    },
}
r.col.UpdateOne(ctx, bson.M{"_id": id}, update)

// ✅ 计数器用 $inc，禁止先查再写（有并发竞争）
r.col.UpdateOne(ctx, bson.M{"_id": postID}, bson.M{
    "$inc": bson.M{"view_count": 1},
})

// ✅ 数组追加用 $push，有去重需求用 $addToSet
r.col.UpdateOne(ctx, bson.M{"_id": userID}, bson.M{
    "$addToSet": bson.M{"tags": newTag},
})

// ❌ 禁止：先 FindOne 查出来改掉再 ReplaceOne 写回（有并发竞争，且多一次网络往返）
```

### 7.2 软删除规范

```go
// ✅ 本项目统一使用软删除，禁止物理删除业务数据
update := bson.M{
    "$set": bson.M{
        "is_deleted": true,
        "deleted_at": time.Now(),
    },
}

// 所有查询必须默认过滤已删除数据
filter := bson.M{
    "author_id":  authorID,
    "is_deleted": bson.M{"$ne": true}, // 或 false
}
```

### 7.3 事务使用规范

```go
// ✅ 跨集合写操作需要事务时（如创建文章同时更新用户统计）
session, err := client.StartSession()
if err != nil {
    return fmt.Errorf("start session: %w", err)
}
defer session.EndSession(ctx)

_, err = session.WithTransaction(ctx, func(sessCtx mongo.SessionContext) (interface{}, error) {
    // 操作1：插入文章
    _, err := postCol.InsertOne(sessCtx, post)
    if err != nil {
        return nil, err
    }
    // 操作2：更新用户文章计数
    _, err = userCol.UpdateOne(sessCtx,
        bson.M{"_id": userID},
        bson.M{"$inc": bson.M{"post_count": 1}},
    )
    return nil, err
})
```

> **事务使用原则**：单集合操作不需要事务（MongoDB 单文档操作本身是原子的）；跨集合写操作才考虑事务；事务会降低吞吐量，能用单文档嵌入替代的尽量嵌入。

---

## 八、集合命名与配置规范

### 8.1 集合名规范

| 规则 | 示例 |
|------|------|
| 全小写，复数形式，下划线分隔 | `users`、`blog_posts`、`token_blacklist` |
| 集合名统一在 Repository 初始化处声明，禁止硬编码字符串散落各处 | `const collUsers = "users"` |

### 8.2 本项目集合规划

| 集合名 | 用途 |
|--------|------|
| `users` | 用户账号信息 |
| `blog_posts` | 博客文章 |
| `documents` | 文档/知识库 |
| `token_blacklist` | JWT refresh token 黑名单 |
| `site_config` | 站点配置项（单文档） |

### 8.3 连接配置（config.yaml）

```yaml
mongodb:
  uri: "mongodb://admin:secret@localhost:27017"
  database: "my_cloud_db"
  max_pool_size: 20
  min_pool_size: 5
  connect_timeout_seconds: 10
  operation_timeout_seconds: 5
```

---

## 九、禁止事项清单

| # | 禁止行为 | 原因 |
|---|----------|------|
| 1 | `$match` 不在聚合管道首位 | 无法走索引，触发全表扫描 |
| 2 | 在大集合上使用 `$skip` 深分页 | 响应时间指数级上升 |
| 3 | 循环内单条 Insert/Update | 应使用 `InsertMany` / `BulkWrite` |
| 4 | 每次请求创建新 `*mongo.Client` | 连接池耗尽，性能崩溃 |
| 5 | DB 操作不传 `context.Context` | 无法超时控制，服务挂死 |
| 6 | 物理删除业务数据（`DeleteOne`）| 数据不可追溯，统一使用软删除 |
| 7 | 先查再改再写（Read-Modify-Write）| 并发竞争，应使用原子操作符 (`$set`/`$inc`) |
| 8 | 无限增长的数组字段 | 文档膨胀超出 16MB 限制 |
| 9 | 在 `$lookup` 的目标集合上缺少索引 | 等同于嵌套循环，性能极差 |
| 10 | 未经 `explain()` 验证就上线聚合查询 | 生产事故风险 |
| 11 | 敏感字段不排除直接返回 | 安全漏洞 |
| 12 | 时间字段用 `int64` 存储 | 应统一用 `time.Time`，便于比较和格式化 |
