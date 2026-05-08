# 生产构建：从 Monorepo 根目录构建 frontend 工作区
FROM node:20.10.0-alpine AS builder
WORKDIR /workspace

ENV CI=true
RUN npm install -g pnpm@10.33.0

# 先复制依赖描述文件以利用构建缓存
COPY package.json pnpm-lock.yaml pnpm-workspace.yaml ./
COPY packages/frontend/package.json ./packages/frontend/package.json

RUN pnpm install --filter ./packages/frontend... --frozen-lockfile

# 复制前端源码并构建
COPY packages/frontend ./packages/frontend
RUN pnpm --filter ./packages/frontend build

FROM nginx:alpine
COPY --from=builder /workspace/packages/frontend/dist /usr/share/nginx/html
COPY packages/frontend/nginx.conf /etc/nginx/conf.d/default.conf
EXPOSE 80
CMD ["nginx", "-g", "daemon off;"]
