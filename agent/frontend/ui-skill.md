# UI/UX & Motion 规范 · Tech-Noir

> **Skill 路径**：`agent/frontend/ui-skill.md`
> **适用范围**：首页及后续所有具由“沉浸式/科技感”需求的页面

## 一、视觉调色盘 (Thematic Palette)

本项目采用动态响应式的主题色系统，通过 CSS 变量绑定。

| 变量 | 默认 (Cyan) | Purple | Green | Pink |
|------|-----------|--------|-------|------|
| `--neon` | `#00f3ff` | `#bc13fe` | `#39ff14` | `#ff00e5` |
| `--neon-glow` | `rgba(0, 243, 255, 0.4)` | `rgba(188, 19, 254, 0.4)` | `rgba(57, 255, 20, 0.4)` | `rgba(255, 0, 229, 0.4)` |

### Tailwind 扩展类
- `text-neon`: 使用主题强调色文字。
- `border-neon`: 主题色边框。
- `shadow-neon`: 主题色霓虹外发光沉淀。
- `bg-glass`: 深色半透明磨砂背景 (`backdrop-blur-md bg-white/5`)。

---

## 二、动画原则 (GSAP & Framer)

### 2.1 入场时序 (Entrance Stagger)
所有首屏元素必须使用 **GSAP Timeline** 进行编排。
- **Ease**: `power4.out`
- **Duration**: `0.8s` - `1.2s`
- **Stagger**: `0.1s` (项与项之间)

```javascript
// 标准入场示例
gsap.from(".hero-title", { 
  y: 30, 
  opacity: 0, 
  duration: 1, 
  ease: "power4.out" 
});
```

### 2.2 呼吸感 (Micro-interactions)
- **呼吸效果**: 状态徽章圆点使用 `animate-pulse` 或 GSAP 循环缩放。
- **悬停位移**: 卡片悬停时向上偏移 `2px`，边框不透明度增加。

---

## 三、Shadcn-vue 适配

- 所有 UI 组件（Button, Card）必须覆盖其默认的 `border-radius` 为较小的 `rounded-md` 或 `rounded-lg`，以保持硬核科技感。
- 按钮默认使用 `glass` 样式，主按钮使用 `shadow-neon` 增强视觉重心。

---

## 四、粒子背景规范

- **密度**: 100 - 150 粒子（宽屏）。
- **连线距离**: `150px` 阈值。
- **流速**: `0.5px/s` 慢速漂移。
