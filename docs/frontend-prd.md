# 前端 PRD 文档

> **框架**：Vue 3 + TypeScript | **风格**：Tech-Noir 沉浸式科技风
> **参见**：`agent/frontend/skill.md` · `agent/frontend/ui-skill.md`

---

## 一、页面清单

| 路由 | 组件 | 功能说明 | 鉴权 |
|------|------|----------|------|
| `/` | `HomeView` | 首页（沉浸式欢迎页） | ❌ |
| `/login` | `LoginView` | 用户登录 | ❌ |
| `/register` | `RegisterView` | 用户注册 | ❌ |
| `/tasks` | `TasksView` | 任务列表主界面 | ✅ |
| `/tasks/:id` | `TaskDetailView` | 任务详情/编辑 | ✅ |
| `/profile` | `ProfileView` | 个人信息设置 | ✅ |
| `*` | `NotFoundView` | 404 页面 | ❌ |

---

## 二、核心功能需求

### 2.1 首页（HomeView）

- 沉浸式星空/粒子背景（tsparticles），100~150 粒子，慢速漂移
- Hero 区域：项目名称 + 标语 + 单一主按钮
- GSAP Timeline 控制元素入场（`power4.out`，stagger 0.1s）
- 首页作为统一入口，不因已登录状态自动跳转；重新打开浏览器或访问根路径时优先展示首页
- 已登录用户仅展示「立即开始」按钮，点击后跳转 `/tasks`
- 未登录用户仅展示「去登录」按钮，点击后跳转 `/login`

---

### 2.2 登录 / 注册页

- 毛玻璃卡片表单（`bg-glass`）+ 霓虹边框
- 登录页和注册页需提供「返回首页」入口
- 表单验证：邮箱格式、密码最少 8 位
- 错误提示：下方行内红色提示文本
- 登录成功后：存储 token → 跳转 `/tasks`
- 注册成功后：显示成功提示 → 跳转 `/login`
- 已登录用户访问 `/login` 或 `/register` 时，由路由守卫跳转回 `/`
- 邮箱等输入框触发浏览器自动填充时，背景、文字和光标样式需保持 Tech-Noir 深色输入态

---

### 2.3 任务列表页（TasksView）— 核心页面

**布局**：
- 顶部导航栏：Logo + 多语言切换 + 主题色切换 + 退出登录
- 左侧筛选栏：状态 / 优先级 / 搜索关键词
- 主内容区：任务卡片列表（虚拟滚动，`vue-virtual-scroller`）

**任务卡片（TaskCard.vue）**：
- 显示：标题、优先级标签（颜色区分）、状态、截止日期
- 悬停：向上偏移 2px + 霓虹边框高亮
- 操作：点击进入详情，右键/长按显示快捷菜单（修改状态、删除）

**新建任务**：
- 页面右下角浮动「+」按钮
- 点击弹出侧边抽屉表单（含标题、描述、优先级、截止日期）
- 提交后列表顶部插入新卡片（乐观更新）

**分页**：游标分页，滚动到底部自动加载更多

---

### 2.4 任务详情页（TaskDetailView）

- 展示全部字段，支持行内编辑（点击字段即可编辑）
- 底部：删除按钮（需二次确认弹窗）
- 修改后需调用 `PATCH /tasks/:id` 接口

---

### 2.5 个人信息页（ProfileView）

- 展示邮箱（只读）、昵称（可编辑）
- 语言切换下拉：中文 / English
- 主题色选择器：4 种霓虹色块点选
- 修改密码区域：旧密码 + 新密码 + 确认新密码

---

## 三、全局状态（Pinia Stores）

| Store | 持久化 | 数据 |
|-------|--------|------|
| `useAuthStore` | ✅ localStorage | `accessToken`, `refreshToken`, `user` |
| `useTaskStore` | ❌ | `tasks[]`, `filters`, `nextCursor` |
| `useThemeStore` | ✅ localStorage | `theme`, `language` |

---

## 四、多语言（vue-i18n）

- 支持 `zh` / `en` 双语
- 语言文件：`src/locales/zh.ts` + `src/locales/en.ts`
- 切换语言后实时生效，无需刷新页面

---

## 五、主题色系统

通过 CSS 变量切换，4 种霓虹主题：

| 主题 | `--neon` 颜色 |
|------|--------------|
| Cyan（默认） | `#00f3ff` |
| Purple | `#bc13fe` |
| Green | `#39ff14` |
| Pink | `#ff00e5` |

---

## 六、性能要求

- 首屏白屏时间 < 2s
- 任务列表支持 1000+ 条数据流畅滚动（虚拟列表）
- 路由全部懒加载

---

## 七、禁止事项

| # | 禁止 |
|---|------|
| 1 | Options API |
| 2 | `any` 类型 |
| 3 | index 作为 `v-for` key |
| 4 | 内联 `style`（除动态计算值） |
| 5 | `npm` / `yarn`（pnpm only） |
