# 自动化测试文档

> **测试策略**：API 集成测试（Python）+ 后端单元测试（Go）
> **脚本位置**：`tests/` 目录（Phase 5 创建）

---

## 一、测试分层

| 类型 | 工具 | 覆盖范围 | 执行时机 |
|------|------|----------|----------|
| 单元测试 | Go `testing` | Service 层业务逻辑 | 每次 commit |
| API 集成测试 | Python `requests` + `pytest` | 所有 HTTP 接口 | Phase 5 + 回归测试 |
| 手动测试 | Apifox | 新接口联调 | 接口开发完成后 |

---

## 二、Python 测试脚本规范

### 2.1 目录结构

```
tests/
├── conftest.py          # 全局配置（base_url、共享 token）
├── helpers.py           # 公共工具函数
├── test_auth.py         # 认证模块测试
├── test_tasks.py        # 任务模块测试
├── test_users.py        # 用户模块测试
└── requirements.txt     # 依赖：pytest requests
```

---

### 2.2 `conftest.py` 示例

```python
import pytest
import requests

BASE_URL = "http://localhost:8080/api/v1"

@pytest.fixture(scope="session")
def base_url():
    return BASE_URL

@pytest.fixture(scope="session")
def auth_token(base_url):
    """注册并获取测试用 token（每次测试会话只执行一次）"""
    resp = requests.post(f"{base_url}/auth/register", json={
        "email": "test_auto@example.com",
        "password": "TestPass123",
        "nickname": "自动化测试"
    })
    # 用户已存在时登录
    if resp.json().get("code") == 409:
        resp = requests.post(f"{base_url}/auth/login", json={
            "email": "test_auto@example.com",
            "password": "TestPass123"
        })
    return resp.json()["data"]["access_token"]

@pytest.fixture
def headers(auth_token):
    return {"Authorization": f"Bearer {auth_token}"}
```

---

### 2.3 `test_auth.py` 示例

```python
def test_register_success(base_url):
    """注册新用户成功"""
    import time
    resp = requests.post(f"{base_url}/auth/register", json={
        "email": f"user_{int(time.time())}@test.com",
        "password": "TestPass123",
        "nickname": "测试用户"
    })
    assert resp.status_code == 200
    assert resp.json()["code"] == 0
    assert "access_token" in resp.json()["data"]


def test_login_wrong_password(base_url):
    """密码错误返回 401"""
    resp = requests.post(f"{base_url}/auth/login", json={
        "email": "test_auto@example.com",
        "password": "WrongPass"
    })
    assert resp.status_code == 401


def test_get_me_without_token(base_url):
    """未携带 token 返回 401"""
    resp = requests.get(f"{base_url}/users/me")
    assert resp.status_code == 401
```

---

### 2.4 `test_tasks.py` 示例

```python
def test_create_task(base_url, headers):
    """创建任务成功"""
    resp = requests.post(f"{base_url}/tasks", json={
        "title": "自动化测试任务",
        "priority": "high"
    }, headers=headers)
    assert resp.status_code == 200
    data = resp.json()["data"]
    assert data["title"] == "自动化测试任务"
    return data["id"]


def test_list_tasks(base_url, headers):
    """获取任务列表"""
    resp = requests.get(f"{base_url}/tasks", headers=headers)
    assert resp.status_code == 200
    assert "items" in resp.json()["data"]


def test_delete_task_forbidden(base_url, headers):
    """不能删除他人任务（需准备另一 token 测试）"""
    # TODO: Phase 5 补充
    pass
```

---

## 三、测试用例清单

### 认证模块

| # | 用例 | 期望结果 |
|---|------|----------|
| A-01 | 正常注册 | 200，返回双 token |
| A-02 | 邮箱已注册重复注册 | 409 |
| A-03 | 邮箱格式非法 | 400 |
| A-04 | 正常登录 | 200，返回双 token |
| A-05 | 密码错误登录 | 401 |
| A-06 | 刷新 Token | 200，返回新 access_token |
| A-07 | 使用已登出 Token | 401 |

### 任务模块

| # | 用例 | 期望结果 |
|---|------|----------|
| T-01 | 创建任务（标题必填） | 200 |
| T-02 | 创建任务（缺少标题） | 400 |
| T-03 | 获取任务列表（无筛选）| 200，data.items 为数组 |
| T-04 | 按状态筛选任务 | 200，仅返回对应状态 |
| T-05 | 获取单个任务 | 200 |
| T-06 | 获取他人任务 | 404 |
| T-07 | 更新任务标题 | 200 |
| T-08 | 删除任务（软删除） | 200，再次获取返回 404 |
| T-09 | 删除他人任务 | 403 |

---

## 四、运行方式

```bash
# 安装依赖
pip install pytest requests

# 确保后端服务运行中
cd tests
pytest -v                       # 运行所有测试
pytest test_auth.py -v          # 只运行认证模块
pytest -k "test_create" -v      # 按关键词过滤
```

---

## 五、CI 集成（Phase 5）

```yaml
# .github/workflows/test.yml（示例）
- name: Run API Tests
  run: |
    cd tests
    pip install pytest requests
    pytest -v --tb=short
```

---

## 六、测试数据清理

- 测试账号统一使用 `test_auto@example.com`
- 测试产生的任务需在测试结束后清理（使用 pytest fixture `yield` + teardown）
- 生产环境**禁止**运行自动化测试脚本
