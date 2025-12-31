# Docker 镜像构建与推送指南

本指南教你如何将本地修改的 new-api 代码构建成 Docker 镜像并推送到 Docker Hub。

## 前提条件

1. 已安装 Docker Desktop
2. 已登录 Docker Hub 账户（你的用户名是 `demokt`）

## 步骤一：登录 Docker Hub

打开终端（PowerShell 或 CMD），执行：

```bash
docker login
```

输入你的 Docker Hub 用户名和密码（或 Access Token）。

> **提示**：如果你已经在 Docker Desktop 中登录，可以跳过这一步。

## 步骤二：构建 Docker 镜像

### 方法 A：使用国内镜像源构建（推荐，速度更快）

进入 new-api 目录，执行：

```bash
cd new-api

# 构建镜像并打标签
# 格式：docker build -f <Dockerfile路径> -t <用户名>/<仓库名>:<标签> .
docker build -f Dockerfile.cn -t demokt/new-api:latest .
```

### 方法 B：使用标准 Dockerfile 构建

```bash
cd new-api
docker build -t demokt/new-api:latest .
```

### 构建参数说明

| 参数 | 说明 |
|------|------|
| `-f Dockerfile.cn` | 指定使用国内镜像源的 Dockerfile |
| `-t demokt/new-api:latest` | 指定镜像标签：`用户名/仓库名:版本` |
| `.` | 构建上下文目录（当前目录） |

## 步骤三：推送镜像到 Docker Hub

构建完成后，推送镜像：

```bash
docker push demokt/new-api:latest
```

## 步骤四（可选）：添加版本标签

为了更好地管理版本，建议同时推送版本标签：

```bash
# 添加版本标签（例如 v1.0.1）
docker tag demokt/new-api:latest demokt/new-api:v1.0.1

# 推送版本标签
docker push demokt/new-api:v1.0.1
```

## 完整命令示例（一键执行）

### Windows PowerShell

```powershell
cd new-api
docker build -f Dockerfile.cn -t demokt/new-api:latest .
docker push demokt/new-api:latest
```

### 带版本号的完整流程

```powershell
cd new-api

# 设置版本号
$VERSION = "v1.0.1-url-fix"

# 构建镜像
docker build -f Dockerfile.cn -t demokt/new-api:latest -t demokt/new-api:$VERSION .

# 推送所有标签
docker push demokt/new-api:latest
docker push demokt/new-api:$VERSION
```

## 验证推送结果

推送成功后，访问 [Docker Hub](https://hub.docker.com/r/demokt/new-api/tags) 查看镜像。

## 常见问题

### Q1: 构建失败，提示网络错误

尝试使用国内镜像源的 Dockerfile：

```bash
docker build -f Dockerfile.cn -t demokt/new-api:latest .
```

### Q2: 推送失败，提示 `denied: requested access to the resource is denied`

重新登录 Docker Hub：

```bash
docker logout
docker login
```

### Q3: 构建很慢

1. 确保使用 `Dockerfile.cn`（国内镜像源）
2. 首次构建会下载依赖，后续构建会使用缓存

### Q4: 如何清理本地旧镜像

```bash
# 查看本地镜像
docker images | grep new-api

# 删除指定镜像
docker rmi demokt/new-api:旧版本

# 清理所有未使用的镜像
docker image prune
```

## 在其他服务器上使用

其他服务器可以直接拉取你推送的镜像：

```bash
docker pull demokt/new-api:latest
docker run -d --name new-api -p 3000:3000 demokt/new-api:latest
```

或者在 docker-compose.yml 中使用：

```yaml
version: '3'
services:
  new-api:
    image: demokt/new-api:latest
    ports:
      - "3000:3000"
    volumes:
      - ./data:/data