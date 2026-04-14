# Go 后端开发规范 · Skill

> **适用范围**：`packages/backend/` 下所有 Go 代码
> **权威来源**：Uber Go Style Guide · Google Go Style Guide · Effective Go · 本项目 `docs/architecture.md`
> **强制级别**：AI 在修改任何后端代码前，必须先加载并严格遵守本文件。

---

## 一、项目技术栈速查

| 类别 | 选型 | 说明 |
|------|------|------|
| Web 框架 | Gin v1.9+ | HTTP 路由与中间件 |
| 数据库驱动 | mongo-driver（官方） | 手写 Repository 层，禁止引入 ODM |
| 配置管理 | Viper | 读取 `configs/config.yaml` + 环境变量 |
| 日志 | Zap | 结构化日志，禁止用 `fmt` 打日志 |
| 认证 | JWT 双 Token | access（短期）+ refresh（长期，黑名单存 MongoDB） |
| Go 版本 | 1.22+ | 可使用泛型、`errors.Join`、`slices` 标准库 |

---

## 二、项目目录结构规范

```
packages/backend/
├── cmd/
│   └── server/
│       └── main.go          # 唯一入口，只做初始化和启动，禁止含业务逻辑
├── internal/                # 禁止被外部包导入
│   ├── handler/             # HTTP 处理器，只做请求解析和响应，不写业务逻辑
│   ├── service/             # 业务逻辑层，编排多个 repository 调用
│   ├── repository/          # 数据库操作层，只负责 MongoDB CRUD
│   ├── model/               # 数据结构定义（struct + BSON/JSON tag）
│   └── middleware/          # Gin 中间件（JWT、CORS、限流、日志）
├── pkg/                     # 可被外部包导入的公共工具
│   ├── config/              # Viper 配置加载，对外暴露统一 Config struct
│   ├── logger/              # Zap 初始化，对外暴露 *zap.Logger 单例
│   └── response/            # 统一 HTTP 响应结构（见第六节）
└── configs/
    └── config.yaml          # 应用配置文件，密钥类配置通过 .env 注入
```

**核心原则**：依赖方向只能向内（handler → service → repository），禁止反向依赖。

---

## 三、命名规范

### 3.1 文件命名
- 全部小写，多词用下划线：`user_handler.go`、`auth_service.go`
- 测试文件以 `_test.go` 结尾，与被测文件同目录

### 3.2 包命名
- 全小写，简短，无下划线：`handler`、`service`、`repository`
- 禁止用包名重复类型名：用 `model.User`，不用 `model.UserModel`

### 3.3 变量和函数命名
| 类型 | 规则 | 示例 |
|------|------|------|
| 导出符号 | PascalCase | `GetUserByID` |
| 未导出符号 | camelCase | `parseToken` |
| 接口 | 动词+er（单方法）或描述性名词 | `Reader`、`UserRepository` |
| 错误变量 | `Err` 前缀 | `ErrUserNotFound` |
| 常量 | PascalCase（导出）/ camelCase（未导出） | `MaxRetryCount` |
| 循环变量 | 简短：`i`、`k`、`v`、`u`（user） | — |

### 3.4 ID 缩写
- 缩写词全大写：`userID`（非 `userId`）、`URL`、`HTTP`、`JSON`
- MongoDB 主键字段统一用 `ID primitive.ObjectID`，JSON tag 为 `"id"`

---

## 四、代码规范

### 4.1 错误处理（最高优先级）

```go
// ✅ 正确：显式处理，使用 %w 包装保留调用链
result, err := repo.FindUser(ctx, id)
if err != nil {
    return fmt.Errorf("service.GetUser: %w", err)
}

// ✅ 正确：可匹配的哨兵错误
var ErrUserNotFound = errors.New("user not found")

// ❌ 禁止：忽略错误
result, _ := repo.FindUser(ctx, id)

// ❌ 禁止：在正常业务流中使用 panic
panic("something went wrong")

// ❌ 禁止：日志 + return err（重复处理）
if err != nil {
    logger.Error("find user failed", zap.Error(err))
    return err  // 调用方还会再打一次日志
}
```

**错误选型决策表**（来自 Uber Guide）：

| 需要匹配错误？ | 错误信息 | 使用方式 |
|---|---|---|
| 否 | 静态字符串 | `errors.New("msg")` |
| 否 | 动态内容 | `fmt.Errorf("context: %w", err)` |
| 是 | 静态字符串 | `var Err = errors.New("msg")` 导出变量 |
| 是 | 动态内容 | 自定义 `error` 类型 |

**一个错误只处理一次**：要么记录日志，要么向上返回，不能两者都做。

### 4.2 接口定义原则

- 接口定义在**使用方**（consumer），不在实现方
- 保持接口小：单方法接口优先
- 编译期校验接口实现：

```go
// 在 repository 包内实现，在 service 包定义接口并校验
var _ UserRepository = (*mongoUserRepository)(nil)
```

### 4.3 Context 传递规范

```go
// ✅ 正确：ctx 始终是第一个参数
func (r *userRepo) FindByID(ctx context.Context, id primitive.ObjectID) (*model.User, error)

// ❌ 禁止：将 ctx 存入 struct
type Service struct {
    ctx context.Context // 禁止
}
```

- 所有涉及 I/O 的函数（MongoDB 查询、HTTP 调用）必须接收并传递 `context.Context`
- Gin handler 使用 `c.Request.Context()` 获取 ctx，不要直接传 `c`

### 4.4 并发规范

```go
// ✅ 正确：defer 解锁，防止忘记
mu.Lock()
defer mu.Unlock()

// ✅ 正确：goroutine 必须能退出，防止泄漏
go func() {
    defer wg.Done()
    select {
    case <-ctx.Done():
        return
    case msg := <-ch:
        process(msg)
    }
}()

// ❌ 禁止：启动后无法取消的 goroutine（fire-and-forget）
go func() {
    for { doWork() }  // 无法退出
}()
```

### 4.5 切片和 Map

```go
// ✅ 正确：预分配容量
users := make([]User, 0, len(docs))

// ✅ 正确：在边界处拷贝，防止外部修改内部状态
func (s *Service) GetUsers() []User {
    result := make([]User, len(s.users))
    copy(result, s.users)
    return result
}
```

### 4.6 避免 init()

- 禁止在 `init()` 中启动 goroutine
- 禁止在 `init()` 中做有副作用的操作（如连接数据库、读取配置）
- 初始化逻辑统一放在 `main()` 或显式的 `New*()` 构造函数中

---

## 五、分层规范（Gin + MongoDB）

### 5.1 Handler 层职责

```go
// handler/user_handler.go
type UserHandler struct {
    svc service.UserService  // 依赖接口，不依赖具体实现
}

func NewUserHandler(svc service.UserService) *UserHandler {
    return &UserHandler{svc: svc}
}

func (h *UserHandler) GetUser(c *gin.Context) {
    id := c.Param("id")
    objID, err := primitive.ObjectIDFromHex(id)
    if err != nil {
        response.BadRequest(c, "invalid id format")
        return
    }

    user, err := h.svc.GetUser(c.Request.Context(), objID)
    if err != nil {
        if errors.Is(err, service.ErrUserNotFound) {
            response.NotFound(c, "user not found")
            return
        }
        response.InternalError(c, "failed to get user")
        return
    }

    response.OK(c, user)
}
```

**Handler 只允许**：解析请求参数 → 调用 service → 映射错误到 HTTP 状态码 → 返回响应

### 5.2 Service 层职责

```go
// service/user_service.go
var ErrUserNotFound = errors.New("user not found")

type UserService interface {
    GetUser(ctx context.Context, id primitive.ObjectID) (*model.User, error)
    CreateUser(ctx context.Context, req CreateUserRequest) (*model.User, error)
}

type userService struct {
    repo   repository.UserRepository
    logger *zap.Logger
}
```

**Service 只允许**：业务逻辑编排 → 调用 repository → 业务错误定义

### 5.3 Repository 层职责

```go
// repository/user_repository.go
type UserRepository interface {
    FindByID(ctx context.Context, id primitive.ObjectID) (*model.User, error)
    FindByEmail(ctx context.Context, email string) (*model.User, error)
    Insert(ctx context.Context, user *model.User) error
    UpdateByID(ctx context.Context, id primitive.ObjectID, update bson.M) error
    DeleteByID(ctx context.Context, id primitive.ObjectID) error
}
```

**MongoDB 操作规范**：
- 查找单条记录时，`ErrNoDocuments` 映射为业务的 `ErrXxxNotFound`
- 所有写操作使用 `context` 超时控制
- 禁止在 repository 层拼接业务逻辑判断

```go
// ✅ 正确：将 mongo.ErrNoDocuments 转换为业务错误
func (r *mongoUserRepo) FindByID(ctx context.Context, id primitive.ObjectID) (*model.User, error) {
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

### 5.4 Model 层规范

```go
// model/user.go
type User struct {
    ID        primitive.ObjectID `bson:"_id,omitempty" json:"id"`
    Email     string             `bson:"email"         json:"email"`
    Password  string             `bson:"password"      json:"-"`    // 禁止返回给前端
    CreatedAt time.Time          `bson:"created_at"    json:"created_at"`
    UpdatedAt time.Time          `bson:"updated_at"    json:"updated_at"`
}
```

- 所有 struct 字段必须同时声明 `bson` 和 `json` tag
- 敏感字段（密码、Token）的 json tag 使用 `"-"`
- 时间字段统一用 `time.Time`，不用 `int64` 时间戳

---

## 六、统一响应结构

所有 API 响应必须通过 `pkg/response` 包输出，禁止直接调用 `c.JSON`。

```go
// pkg/response/response.go
type Response struct {
    Code    int         `json:"code"`
    Message string      `json:"message"`
    Data    interface{} `json:"data,omitempty"`
}

func OK(c *gin.Context, data interface{}) {
    c.JSON(http.StatusOK, Response{Code: 0, Message: "success", Data: data})
}

func BadRequest(c *gin.Context, msg string) {
    c.JSON(http.StatusBadRequest, Response{Code: 400, Message: msg})
}

func NotFound(c *gin.Context, msg string) {
    c.JSON(http.StatusNotFound, Response{Code: 404, Message: msg})
}

func InternalError(c *gin.Context, msg string) {
    c.JSON(http.StatusInternalServerError, Response{Code: 500, Message: msg})
}

func Unauthorized(c *gin.Context, msg string) {
    c.JSON(http.StatusUnauthorized, Response{Code: 401, Message: msg})
}
```

---

## 七、日志规范（Zap）

```go
// ✅ 正确：结构化字段，便于日志系统索引
logger.Info("user created",
    zap.String("user_id", user.ID.Hex()),
    zap.String("email", user.Email),
    zap.Duration("elapsed", elapsed),
)

logger.Error("failed to create user",
    zap.Error(err),
    zap.String("email", req.Email),
)

// ❌ 禁止：字符串拼接或使用 fmt
fmt.Println("user created: " + user.ID.Hex())
log.Printf("error: %v", err)
```

- 日志级别：`Debug`（开发调试）、`Info`（关键业务节点）、`Warn`（可恢复异常）、`Error`（需关注的错误）
- 禁止在 repository 层打日志，日志统一在 service 层或 middleware 层记录
- 生产环境使用 JSON 格式输出，通过 Viper 配置切换

---

## 八、JWT 双 Token 中间件规范

```go
// middleware/auth.go
func JWTAuth(secret string) gin.HandlerFunc {
    return func(c *gin.Context) {
        tokenStr := c.GetHeader("Authorization")
        if tokenStr == "" || !strings.HasPrefix(tokenStr, "Bearer ") {
            response.Unauthorized(c, "missing or invalid token")
            c.Abort()
            return
        }

        claims, err := parseAccessToken(strings.TrimPrefix(tokenStr, "Bearer "), secret)
        if err != nil {
            response.Unauthorized(c, "token expired or invalid")
            c.Abort()
            return
        }

        c.Set("user_id", claims.UserID)
        c.Next()
    }
}
```

- access_token 有效期：15 分钟（配置在 config.yaml）
- refresh_token 有效期：7 天，黑名单存 MongoDB `token_blacklist` 集合
- Token 通过 `Authorization: Bearer <token>` 头传递，不接受 Query 参数方式

---

## 九、Import 分组规范

```go
import (
    // 第一组：标准库
    "context"
    "errors"
    "fmt"

    // 第二组：第三方库（空行分隔）
    "github.com/gin-gonic/gin"
    "go.mongodb.org/mongo-driver/bson"
    "go.uber.org/zap"

    // 第三组：项目内部包（空行分隔）
    "backend/internal/model"
    "backend/pkg/response"
)
```

使用 `goimports` 工具自动管理（保存时自动运行）。

---

## 十、代码格式化与 Lint 规范

### 强制工具
| 工具 | 作用 | 何时运行 |
|------|------|----------|
| `gofmt` / `goimports` | 格式化代码 + 管理 import | 保存时自动运行 |
| `go vet` | 静态检查常见错误 | CI 必须通过 |
| `golangci-lint` | 综合 lint（包含 errcheck、staticcheck 等） | CI 必须通过 |

### CI 检查命令
```bash
go vet ./...
golangci-lint run ./...
go test ./... -race
```

---

## 十一、禁止事项清单

| # | 禁止行为 | 原因 |
|---|----------|------|
| 1 | 在 handler 层写 MongoDB 查询 | 违反分层规范 |
| 2 | 忽略错误返回值（`result, _ := ...`） | 掩盖潜在 bug |
| 3 | 用 `fmt.Println` / `log.Printf` 打日志 | 必须使用 Zap |
| 4 | 直接调用 `c.JSON(...)` 返回响应 | 必须使用 `response` 包 |
| 5 | 在 struct 中存储 `context.Context` | 违反 Go 规范 |
| 6 | 启动无法退出的 goroutine | 导致 goroutine 泄漏 |
| 7 | 在 `init()` 中连接数据库或启动 goroutine | 不可测试，难以控制 |
| 8 | 使用 `panic` 处理业务错误 | 应显式返回 error |
| 9 | 密码等敏感字段出现在 JSON 响应中 | 安全风险 |
| 10 | 在 repository 层写业务判断逻辑 | 职责混淆 |
