# todotask Docker + GitHub Actions CI/CD 部署方案

> 技术栈：Vue3 + Vite / Go / MongoDB / pnpm Monorepo  
> 部署方式：GitHub Actions + appleboy/ssh-action 全自动化  
> 目标环境：阿里云 / 腾讯云 Ubuntu 22.04

---

## 一、方案概述

### 整体流程

```
git push main
    → GitHub Actions 触发
    → 并行构建前端镜像（Vue3+Nginx）+ 后端镜像（Go+Alpine）
    → 推送到 GitHub Container Registry（GHCR，免费）
    → appleboy/ssh-action SSH 进服务器
    → docker compose pull + up -d（滚动重启，MongoDB 数据卷不受影响）
```

**全程无需手动 SSH，push 代码后约 3~5 分钟自动完成部署。**

### 最终目录结构

需要在项目中新增/修改以下文件：

```
todotask/
├── packages/
│   ├── backend/
│   │   ├── Dockerfile          ← 修改（多阶段构建）
│   │   └── ...
│   └── frontend/
│       ├── Dockerfile          ← 修改（多阶段构建）
│       ├── nginx.conf          ← 确认/修改
│       └── ...
├── docker-compose.yml          ← 新增（根目录）
├── .env.example                ← 新增
└── .github/
    └── workflows/
        └── deploy.yml          ← 新增
```

---

## 二、需要新增 / 修改的文件

### 2.1 `packages/backend/Dockerfile`（修改）

使用多阶段构建，构建阶段编译 Go 二进制，运行阶段使用极小的 Alpine 镜像，最终镜像体积约 15MB。

```dockerfile
# ===== 构建阶段 =====
FROM golang:1.22-alpine AS builder
WORKDIR /app

# 先复制依赖文件，利用 Docker 层缓存
COPY go.mod go.sum ./
RUN go mod download

# 复制源码并构建
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build \
    -ldflags='-w -s' \
    -o server ./cmd/server/main.go

# ===== 运行阶段 =====
FROM alpine:3.19
WORKDIR /app

# 安装证书（HTTPS 请求需要）
RUN apk --no-cache add ca-certificates tzdata

COPY --from=builder /app/server .

ENV TZ=Asia/Shanghai

EXPOSE 8080
CMD ["./server"]
```

---

### 2.2 `packages/frontend/Dockerfile`（修改）

两阶段构建：Node 环境打包 Vue3 静态文件，再用 Nginx 镜像托管，最终镜像约 30MB。

```dockerfile
# ===== 构建阶段 =====
FROM node:20-alpine AS builder
WORKDIR /app

# 安装 pnpm
RUN npm install -g pnpm

# 复制依赖文件
COPY package.json pnpm-lock.yaml ./
RUN pnpm install --frozen-lockfile

# 复制源码并构建
COPY . .
RUN pnpm build

# ===== 运行阶段 =====
FROM nginx:alpine

# 复制打包产物
COPY --from=builder /app/dist /usr/share/nginx/html

# 复制 Nginx 配置
COPY nginx.conf /etc/nginx/conf.d/default.conf

EXPOSE 80
CMD ["nginx", "-g", "daemon off;"]
```

---

### 2.3 `packages/frontend/nginx.conf`（确认/修改）

前端 Nginx 作为反向代理，把 `/api/*` 请求转发到 Go 后端容器，同时处理 Vue Router 的 history 模式。

```nginx
server {
    listen 80;
    server_name _;

    root /usr/share/nginx/html;
    index index.html;

    # Vue Router history 模式支持
    location / {
        try_files $uri $uri/ /index.html;
    }

    # 反向代理到 Go 后端（容器名 backend，端口 8080）
    location /api/ {
        proxy_pass http://backend:8080;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_connect_timeout 30s;
        proxy_read_timeout 60s;
    }
}
```

> ⚠️ `proxy_pass` 里的 `backend` 是 docker compose 的服务名，Docker 网络内部会自动解析，不需要填 IP。

---

### 2.4 `docker-compose.yml`（根目录新增）

统一管理三个服务：frontend / backend / mongodb。

```yaml
version: '3.8'

services:
  frontend:
    image: ghcr.io/你的GitHub用户名/todotask-frontend:latest
    container_name: todotask-frontend
    ports:
      - "80:80"
    depends_on:
      - backend
    restart: unless-stopped
    networks:
      - todotask-net

  backend:
    image: ghcr.io/你的GitHub用户名/todotask-backend:latest
    container_name: todotask-backend
    expose:
      - "8080"
    environment:
      - MONGODB_URI=mongodb://mongo:27017/todotask
      - GIN_MODE=release
      # 如有其他环境变量在此添加，例如 JWT_SECRET=xxx
    depends_on:
      - mongo
    restart: unless-stopped
    networks:
      - todotask-net

  mongo:
    image: mongo:7
    container_name: todotask-mongo
    expose:
      - "27017"
    volumes:
      - mongo_data:/data/db
    restart: unless-stopped
    networks:
      - todotask-net

volumes:
  mongo_data:  # 数据持久化，重启不丢失

networks:
  todotask-net:
    driver: bridge
```

> ⚠️ 将 `你的GitHub用户名` 替换为实际的 GitHub 用户名（小写）。

---

### 2.5 `.env.example`（根目录新增）

提交到 Git 的环境变量模板，实际的 `.env` 文件加入 `.gitignore` 不提交。

```env
# GitHub 仓库所有者（你的 GitHub 用户名，小写）
GITHUB_REPOSITORY_OWNER=your-github-username

# MongoDB 连接地址（docker compose 内部通信，无需修改）
MONGODB_URI=mongodb://mongo:27017/todotask

# 其他后端环境变量（按需添加）
# JWT_SECRET=xxx
# API_KEY=xxx
```

---

### 2.6 `.github/workflows/deploy.yml`（新增）

核心 CI/CD 文件，push 到 main 分支时触发，完成构建、推送、部署全流程。

```yaml
name: Build & Deploy

on:
  push:
    branches: [main]
  workflow_dispatch:  # 支持手动触发

env:
  REGISTRY: ghcr.io
  IMAGE_PREFIX: ghcr.io/${{ github.repository_owner }}

jobs:
  build-and-deploy:
    runs-on: ubuntu-latest
    permissions:
      contents: read
      packages: write  # 推送到 GHCR 必须

    steps:
      - name: Checkout
        uses: actions/checkout@v4

      - name: Login to GHCR
        uses: docker/login-action@v3
        with:
          registry: ghcr.io
          username: ${{ github.actor }}
          password: ${{ secrets.GITHUB_TOKEN }}

      # ===== 构建并推送前端镜像 =====
      - name: Build & push frontend
        uses: docker/build-push-action@v5
        with:
          context: ./packages/frontend
          push: true
          tags: |
            ${{ env.IMAGE_PREFIX }}/todotask-frontend:latest
            ${{ env.IMAGE_PREFIX }}/todotask-frontend:${{ github.sha }}

      # ===== 构建并推送后端镜像 =====
      - name: Build & push backend
        uses: docker/build-push-action@v5
        with:
          context: ./packages/backend
          push: true
          tags: |
            ${{ env.IMAGE_PREFIX }}/todotask-backend:latest
            ${{ env.IMAGE_PREFIX }}/todotask-backend:${{ github.sha }}

      # ===== SSH 进服务器自动部署 =====
      - name: Deploy to server
        uses: appleboy/ssh-action@v1
        with:
          host: ${{ secrets.SERVER_HOST }}
          username: ${{ secrets.SERVER_USER }}
          key: ${{ secrets.SSH_PRIVATE_KEY }}
          script: |
            cd /opt/todotask
            docker compose pull
            docker compose up -d --remove-orphans
            docker image prune -f
```

> 💡 将 GHCR 包设为 Public 后（第三部分第4步），服务器拉取镜像不需要登录，deploy.yml 无需额外的 GHCR 登录步骤，配置最简洁。

---

## 三、GitHub 上需要做的操作

### 3.1 开启 Actions 写权限

1. 进入仓库页面 → **Settings** → 左侧 **Actions** → **General**
2. 找到 **Workflow permissions**
3. 选择 **Read and write permissions**
4. 点击 Save

---

### 3.2 创建 Personal Access Token（PAT）

用于服务器登录 GHCR 拉取镜像（Actions 自带的 `GITHUB_TOKEN` 只在 Actions 运行期间有效，服务器上无法使用）。

1. 右上角头像 → **Settings** → 最底部 **Developer settings**
2. **Personal access tokens** → **Tokens (classic)** → **Generate new token (classic)**
3. Note 填：`todotask-ghcr-pull`
4. 勾选权限：`read:packages`（只需要这一个）
5. 点击 **Generate token**
6. **立即复制保存**，token 只显示一次！

---

### 3.3 配置 GitHub Secrets

进入仓库 → **Settings** → **Secrets and variables** → **Actions** → **New repository secret**，依次添加以下 4 个：

| Secret 名称 | 填写内容 |
|---|---|
| `SERVER_HOST` | 云服务器公网 IP，例如 `123.456.78.90` |
| `SERVER_USER` | SSH 登录用户名，通常是 `root` 或 `ubuntu` |
| `SSH_PRIVATE_KEY` | SSH 私钥完整内容（第四部分第1步生成） |
| `GHCR_TOKEN` | 上一步生成的 PAT token（暂时备用，设为 Public 后可不用） |

> ⚠️ `SSH_PRIVATE_KEY` 需粘贴完整内容，包括 `-----BEGIN OPENSSH PRIVATE KEY-----` 和 `-----END OPENSSH PRIVATE KEY-----` 两行，私钥本身含换行是正常的。

---

### 3.4 首次构建后将镜像设为 Public

第一次 Actions 构建完成后，GHCR 会生成两个包，需要将它们设为公开：

1. 进入 **GitHub 个人主页** → **Packages**
2. 找到 `todotask-frontend` 和 `todotask-backend`
3. 分别进入 → **Package settings** → 将 **Visibility** 改为 **Public**

设为 Public 后，服务器无需登录即可拉取镜像，deploy.yml 配置最简洁。

---

## 四、服务器上需要做的操作

> 以下命令均在云服务器 Ubuntu 22.04 上执行。

### 4.1 生成 SSH 密钥对（本机执行，不是服务器）

```bash
# 生成专用密钥对
ssh-keygen -t ed25519 -C "github-actions-deploy" -f ~/.ssh/todotask_deploy

# 查看公钥（下一步要用）
cat ~/.ssh/todotask_deploy.pub

# 查看私钥（复制内容填入 GitHub Secrets 的 SSH_PRIVATE_KEY）
cat ~/.ssh/todotask_deploy
```

---

### 4.2 将公钥加入服务器

SSH 登录服务器后执行：

```bash
# 如果 .ssh 目录不存在先创建
mkdir -p ~/.ssh && chmod 700 ~/.ssh

# 将上一步的公钥内容追加进去（替换 <公钥内容> 为实际内容）
echo '<公钥内容>' >> ~/.ssh/authorized_keys
chmod 600 ~/.ssh/authorized_keys
```

在本机验证是否能用新密钥登录：

```bash
ssh -i ~/.ssh/todotask_deploy root@<服务器IP>
```

---

### 4.3 安装 Docker 和 Docker Compose

```bash
# 更新包列表
sudo apt update

# 安装 Docker（官方一键脚本）
curl -fsSL https://get.docker.com | sh

# 将当前用户加入 docker 组（避免每次 sudo）
sudo usermod -aG docker $USER
newgrp docker

# 验证安装
docker --version
docker compose version
```

---

### 4.4 创建项目目录并上传 docker-compose.yml

```bash
# 创建目录
sudo mkdir -p /opt/todotask
sudo chown $USER:$USER /opt/todotask
```

在本机将 `docker-compose.yml` 上传到服务器：

```bash
scp -i ~/.ssh/todotask_deploy docker-compose.yml root@<服务器IP>:/opt/todotask/
```

> ⚠️ 上传前先将 `docker-compose.yml` 里的 `你的GitHub用户名` 替换为实际的 GitHub 用户名（小写）。

---

### 4.5 开放防火墙端口

**系统防火墙（服务器内执行）：**

```bash
sudo ufw allow 22     # SSH
sudo ufw allow 80     # HTTP
sudo ufw allow 443    # HTTPS（备用）
sudo ufw enable
sudo ufw status
```

**云控制台安全组：**

1. 登录阿里云 / 腾讯云控制台
2. 找到对应服务器 → **安全组** → **入站规则**
3. 添加规则：TCP 端口 `80`，来源 `0.0.0.0/0`
4. 添加规则：TCP 端口 `22`，来源限制为你的开发机 IP（提升安全性）

---

### 4.6 首次手动启动服务（验证配置）

在第一次 Actions 运行前，手动启动一次验证配置正确：

```bash
cd /opt/todotask

# 拉取镜像并启动（镜像设为 Public 后无需登录）
docker compose pull
docker compose up -d

# 查看容器状态（三个容器都应为 Up）
docker compose ps

# 查看日志
docker compose logs -f
```

---

## 五、验证部署成功

### 5.1 推送代码触发 CI

```bash
git add .
git commit -m "feat: add Docker CI/CD configuration"
git push origin main
```

### 5.2 查看 Actions 运行状态

1. 进入 GitHub 仓库 → 顶部菜单 **Actions**
2. 找到最新一次 **Build & Deploy** 工作流
3. 点击进入查看每个 step 的日志
4. 全部绿色 ✅ 说明部署成功

### 5.3 访问服务

```bash
# 浏览器访问前端
http://<服务器IP>

# 测试后端 API（根据你的实际路由修改）
curl http://<服务器IP>/api/health

# 服务器上查看容器状态
docker compose -f /opt/todotask/docker-compose.yml ps
```

---

## 六、常见问题排查

| 问题现象 | 排查方法 |
|---|---|
| Actions 推送镜像失败 403 | 检查仓库 Settings → Actions → 是否开启 Read and write permissions |
| SSH 连接失败 | 检查 `SSH_PRIVATE_KEY` Secret 是否包含完整私钥（包括首尾 `---` 行） |
| 服务器拉取镜像失败 | 将 GHCR 包设为 Public，或检查 PAT 是否有 `read:packages` 权限 |
| 前端访问 502 Bad Gateway | 检查 `nginx.conf` 的 `proxy_pass` 是否为 `http://backend:8080` |
| MongoDB 数据丢失 | 不要使用 `docker compose down -v`，`-v` 参数会删除数据卷 |
| 容器启动后立即退出 | 执行 `docker compose logs backend` 查看后端报错信息 |

---

## 七、后续扩展：绑定域名 + HTTPS

服务跑起来后，如需绑定域名并开启 HTTPS，推荐使用 **Caddy**（自动申请 SSL 证书，零配置）：

```bash
# 安装 Caddy
sudo apt install -y debian-keyring debian-archive-keyring apt-transport-https
curl -1sLf 'https://dl.cloudsmith.io/public/caddy/stable/gpg.key' | sudo gpg --dearmor -o /usr/share/keyrings/caddy-stable-archive-keyring.gpg
curl -1sLf 'https://dl.cloudsmith.io/public/caddy/stable/debian.deb.txt' | sudo tee /etc/apt/sources.list.d/caddy-stable.list
sudo apt update && sudo apt install caddy
```

编辑 `/etc/caddy/Caddyfile`：

```
yourdomain.com {
    reverse_proxy localhost:80
}
```

然后 `sudo systemctl reload caddy`，Caddy 会自动申请并续期 SSL 证书。记得在云控制台安全组开放 443 端口。