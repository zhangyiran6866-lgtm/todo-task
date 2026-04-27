# 前端开发规范 Skill

> **Skill 路径**：`agent/frontend/skill.md`
> **生效范围**：所有 `packages/frontend/` 下的代码，以及任何涉及前端的改动
> **执行优先级**：每次生成/修改前端代码，**必须全部遵循本文件规范，无例外**

---

## 一、技术栈约束

- **框架**：Vue 3，必须使用 `<script setup lang="ts">` 组合式 API，**禁止** Options API
- **语言**：TypeScript，**禁止** 写 `.js` 业务文件
- **样式**：Tailwind CSS 为主（≥80%），Less 处理复杂嵌套/动态计算场景，**两者不混用行内**
- **状态管理**：Pinia（全局）/ VueUse `createInjectionState`（页面级局部）
- **包管理器**：pnpm（**禁止** npm / yarn）

---

## 二、Vue 组件规范

### 2.1 文件结构顺序（必须遵守）

```vue
<!-- 固定顺序：script → template → style -->
<script setup lang="ts">
// 1. 外部库 import（vue、vue-router、第三方库）
// 2. 内部模块 import（组件、composables、types、utils、api）
// 3. defineProps / defineEmits / defineExpose
// 4. 响应式状态（ref / reactive / computed）
// 5. 生命周期钩子
// 6. 函数定义（先 private，再 public/handler）
</script>

<template>
  <!-- 单根节点，语义化标签 -->
</template>

<style lang="less" scoped>
/* 仅写 Tailwind 覆盖不了的复杂样式 */
</style>
```

### 2.2 Props 规范

```typescript
// ✅ 正确：用 interface + withDefaults
interface Props {
  title: string
  count?: number
  isVisible?: boolean
}
const props = withDefaults(defineProps<Props>(), {
  count: 0,
  isVisible: true
})

// ❌ 错误：用运行时声明或 any
defineProps({ title: String, count: Number })
```

### 2.3 Emits 规范

```typescript
// ✅ 正确：类型化 emit
const emit = defineEmits<{
  change: [value: string]
  'update:modelValue': [value: boolean]
}>()

// ❌ 错误：字符串数组形式
const emit = defineEmits(['change', 'update:modelValue'])
```

### 2.4 组件命名

- 文件名：**PascalCase**（`UserCard.vue`、`NavHeader.vue`）
- 模板中使用：**PascalCase**（`<UserCard />`，**禁止** `<user-card />`）
- 必须是多词组合，**禁止** 单词组件名（如 `Card.vue` → `UserCard.vue`）

### 2.5 template 规范

- 单根节点（无需 `<template>` 包裹时用语义化标签，如 `<main>`、`<section>`、`<article>`）
- 属性过多时每个属性独占一行，末尾属性不带逗号
- `v-if` 和 `v-for` **不能同级使用**，需用 `<template>` 拆分
- `v-for` 必须绑定 `:key`，key 值优先使用唯一业务 ID，**禁止** 使用 index

```vue
<!-- ✅ 正确 -->
<template v-if="isLogin">
  <UserCard
    v-for="user in users"
    :key="user.id"
    :user="user"
    @click="handleUserClick(user)"
  />
</template>

<!-- ❌ 错误 -->
<div v-if="isLogin" v-for="user in users" :key="index">...</div>
```

---

## 三、TypeScript 规范

### 3.1 类型声明

```typescript
// ✅ 优先类型推导，减少冗余标注
const count = ref(0)              // 自动推导 Ref<number>
const name = ref('')              // 自动推导 Ref<string>

// ✅ 接口描述数据结构
interface UserInfo {
  id: string
  name: string
  avatar?: string
  createdAt: number
}

// ✅ type 用于联合类型/工具类型
type Theme = 'dark' | 'light' | 'system'
type Nullable<T> = T | null

// ❌ 禁止 any，必要时用 unknown + 类型收窄
const result: any = fetch(url) // 禁止！
```

### 3.2 接口返回类型推导

```typescript
import { getUserInfo } from '@/api/user'

// ✅ 自动同步接口返回类型，无需手动维护
const userInfo: Awaited<ReturnType<typeof getUserInfo>> = await getUserInfo(id)
```

### 3.3 非空断言

```typescript
// ❌ 过度使用非空断言
const el = document.getElementById('app')!

// ✅ 先判断再使用
const el = document.getElementById('app')
if (!el) return
```

### 3.4 枚举使用

```typescript
// ✅ 使用 const enum 节省编译产物
const enum RouteNames {
  Home = 'home',
  Blog = 'blog',
  Lab = 'lab'
}
```

---

## 四、命名规范（完整版）

| 场景 | 规范 | 示例 |
|------|------|------|
| 普通变量 | camelCase | `userName`, `pageSize` |
| 常量 | CONSTANT_CASE | `MAX_RETRY_COUNT`, `BASE_URL` |
| 布尔变量 | `is/has/should/can/will` 前缀 | `isLoading`, `hasPermission` |
| 函数/方法 | camelCase + 动宾结构 | `getUserInfo()`, `validateForm()` |
| 异步/请求函数 | `fetch`/`load` 开头 | `fetchUserList()`, `loadPageData()` |
| 事件处理函数 | `handle` 开头 | `handleButtonClick()`, `handleFormSubmit()` |
| 类/构造函数 | PascalCase | `UserService`, `HttpClient` |
| Vue 组件（文件） | PascalCase | `UserCard.vue`, `NavHeader.vue` |
| 目录 | kebab-case | `user-center/`, `nav-header/` |
| 普通 TS 文件 | kebab-case | `format-time.ts`, `use-scroll.ts` |
| Composable | `use` 前缀 + camelCase | `useUserStore.ts`, `useIntersection.ts` |
| Pinia Store | `use` + 名词 + `Store` | `useUserStore`, `useThemeStore` |
| CSS class | BEM + kebab-case | `.user-card__name--active` |
| DOM id | kebab-case，禁数字开头 | `id="submit-btn"` ✅ / `id="1btn"` ❌ |
| 路由 name | kebab-case | `name: 'user-profile'` |
| 事件名（emit） | kebab-case | `emit('update:model-value')` |

---

## 五、目录结构规范

```
src/
├── assets/           ← 静态资源（图片、字体、svg）
│   ├── images/
│   └── fonts/
├── components/       ← 全局通用组件（纯展示/基础组件）
│   └── base/         ← 基础原子组件（BaseButton.vue 等）
├── composables/      ← 自定义 Hooks（useXxx.ts）
├── layouts/          ← 页面布局组件
├── router/           ← 路由配置
│   └── index.ts
├── stores/           ← Pinia stores（useXxxStore.ts）
├── types/            ← 全局 TypeScript 类型定义（*.d.ts）
├── utils/            ← 工具函数（kebab-case.ts）
├── api/              ← 接口请求层
│   └── modules/      ← 按业务模块拆分
├── views/            ← 页面级组件（与路由一一对应）
│   └── home/
│       ├── index.vue           ← 页面入口
│       └── components/         ← 页面私有组件
├── App.vue
├── main.ts
├── style.css         ← 全局样式（Tailwind + 全局 CSS 变量）
└── vite-env.d.ts
```

---

## 六、Composable 规范

```typescript
// src/composables/use-counter.ts
import { ref, computed } from 'vue'

// ✅ 规范：命名导出、返回对象解构友好
export function useCounter(initialValue = 0) {
  const count = ref(initialValue)
  const doubled = computed(() => count.value * 2)
  const isPositive = computed(() => count.value > 0)

  function increment() {
    count.value++
  }

  function reset() {
    count.value = initialValue
  }

  // 统一在函数末尾 return
  return { count, doubled, isPositive, increment, reset }
}
```

规则：
- 文件名 kebab-case（`use-counter.ts`），函数名 camelCase（`useCounter`）
- 必须在组件内调用（响应性绑定组件生命周期）
- 返回值统一在末尾 `return`，便于阅读
- 副作用需在 `onUnmounted` 中清理

---

## 七、样式规范

### 7.1 Tailwind CSS

```vue
<!-- ✅ 优先使用 Tailwind 工具类 -->
<div class="flex items-center gap-4 px-6 py-3 rounded-lg bg-slate-800">

<!-- ✅ 响应式：移动优先 -->
<div class="text-sm md:text-base lg:text-lg">

<!-- ✅ 状态变体 -->
<button class="bg-blue-600 hover:bg-blue-500 active:bg-blue-700 disabled:opacity-50">

<!-- ❌ 禁止内联 style（除动态计算值） -->
<div style="color: red">
```

### 7.2 Less（仅复杂场景）

```less
// ✅ BEM 命名 + 嵌套
.user-card {
  &__header { ... }
  &__body { ... }
  &__footer { ... }
  &--active { ... }
  &--disabled { opacity: 0.5; }
}

// ✅ CSS 变量结合 Less
:root {
  --color-primary: #3b82f6;
}
.btn-primary {
  background-color: var(--color-primary);
}
```

### 7.3 scoped 使用规则

- 组件私有样式：**必须** 加 `scoped`
- 需要穿透子组件：使用 `:deep(.child-class)`，**禁止** 删除 `scoped`
- 全局样式只写在 `style.css` 或布局组件中

---

## 八、Prettier 配置（`.prettierrc`）

```json
{
  "semi": false,
  "singleQuote": true,
  "tabWidth": 2,
  "printWidth": 120,
  "trailingComma": "none",
  "bracketSpacing": true,
  "arrowParens": "avoid",
  "bracketSameLine": false,
  "quoteProps": "as-needed",
  "proseWrap": "preserve",
  "endOfLine": "auto",
  "overrides": [
    { "files": ["*.vue"], "options": { "tabWidth": 2, "endOfLine": "crlf", "printWidth": 180 } },
    { "files": ["*.tsx", "*.jsx"], "options": { "parser": "typescript", "printWidth": 120 } },
    { "files": ["*.md"], "options": { "printWidth": 80, "proseWrap": "always" } },
    { "files": ["*.json"], "options": { "printWidth": 200, "trailingComma": "none" } }
  ]
}
```

---

## 九、import 顺序规范

```typescript
// 1. Vue 核心
import { ref, computed, onMounted } from 'vue'
import { useRouter, useRoute } from 'vue-router'

// 2. 第三方库
import { storeToRefs } from 'pinia'
import gsap from 'gsap'

// 3. 内部 —— stores
import { useUserStore } from '@/stores/use-user-store'

// 4. 内部 —— composables
import { useCounter } from '@/composables/use-counter'

// 5. 内部 —— 组件
import UserCard from '@/components/UserCard.vue'

// 6. 内部 —— types / utils / api
import type { UserInfo } from '@/types/user'
import { formatDate } from '@/utils/format-date'
import { fetchUserInfo } from '@/api/modules/user'
```

规则：
- 各分组之间空一行
- 类型导入使用 `import type`（减少运行时产物）
- 使用路径别名 `@/` 代替相对路径（超过 2 层时）

---

## 十、性能规范

```typescript
// ✅ 大列表使用虚拟滚动（vue-virtual-scroller）
// ✅ 路由懒加载
const BlogView = () => import('@/views/blog/index.vue')

// ✅ 耗时计算使用 computed 缓存
const sortedList = computed(() => [...list.value].sort(...))

// ✅ Three.js：页面卸载时必须销毁
onUnmounted(() => {
  renderer.dispose()
  scene.clear()
})

// ✅ 事件监听在 onUnmounted 中清理
onMounted(() => window.addEventListener('resize', handleResize))
onUnmounted(() => window.removeEventListener('resize', handleResize))

// ❌ 禁止在 template 中写复杂表达式，抽成 computed
// BAD:   {{ list.filter(i => i.active).map(i => i.name).join(', ') }}
// GOOD:  {{ activeNames }}
```

---

## 十一、Git 提交规范

格式：`type(scope): message`（全小写英文，不加句号）

| type | 含义 |
|------|------|
| `feat` | 新功能 |
| `fix` | Bug 修复 |
| `style` | 代码格式调整（不影响逻辑） |
| `refactor` | 重构（不是新功能也不是修 bug） |
| `perf` | 性能优化 |
| `chore` | 构建/工具/依赖更新 |
| `docs` | 文档更新 |
| `test` | 测试相关 |
| `revert` | 回滚 |

示例：
```
feat(home): add particle starfield hero section
fix(auth): handle refresh token expiry edge case
chore(deps): upgrade vue to 3.5.0
style(UserCard): apply prettier formatting
```

---

## 十二、禁止清单（AI 生成代码请主动规避）

1. ❌ 禁止 Options API（`export default { data(), methods: {} }`）
2. ❌ 禁止 `any` 类型
3. ❌ 禁止 index 作为 `v-for` 的 key
4. ❌ 禁止 `v-if` 与 `v-for` 同级
5. ❌ 禁止内联 `style`（除动态计算值）
6. ❌ 禁止直接操作 DOM（用 `ref` + `onMounted`）
7. ❌ 禁止在 template 中写复杂逻辑（提取为 `computed`）
8. ❌ 禁止忘记清理 Three.js renderer 和事件监听
9. ❌ 禁止单词组件名
10. ❌ 禁止数字开头的 DOM id
11. ❌ 禁止使用 `npm` 或 `yarn`（pnpm only）
12. ❌ 禁止使用相对路径超过 2 层（用 `@/` 别名）
