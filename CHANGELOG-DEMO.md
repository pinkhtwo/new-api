# New API Demo 版更新日志

> 本文档记录了从原作者仓库 (QuantumNous/new-api) fork 后的所有本地修改和优化内容

## ✅ 已清理的无效代码

以下早期设计失误的代码已被清理：

| 文件 | 已清理内容 |
|------|----------|
| `relay/channel/gemini/adaptor.go` | 移除了渠道类型判断的死代码（这些代码永远不会被执行） |

---

## 📊 变更统计

| 项目 | 数量 |
|------|------|
| 修改提交数 | 11 |
| 新增文件 | 7 |
| 修改文件 | 13 |
| 新增代码行 | +1433 |
| 删除代码行 | -300 |

---

## 🎯 主要功能特性

### 1. URL 防呆设计 (核心特性)

#### 功能描述
- **智能路径规范化**：自动处理 URL 格式问题，避免因末尾斜杠、多余空格、双斜杠等导致的请求失败
- **API 端点智能提取**：支持用户在 URL 中添加任意前缀后仍能正确路由到 API 端点
- **SillyTavern 兼容**：修复当 SillyTavern 使用带尾部斜杠的 URL 时产生双斜杠路径导致路由失败的问题

#### 相关文件
| 文件 | 变更类型 | 说明 |
|------|---------|------|
| [`common/path_normalize.go`](common/path_normalize.go) | 新增 | 路径规范化核心逻辑，包含 API 端点智能提取 |
| [`common/path_normalize_test.go`](common/path_normalize_test.go) | 新增 | 路径规范化单元测试 (272 行) |
| [`common/embed-file-system.go`](common/embed-file-system.go) | 新增 | 嵌入式文件系统支持 |
| [`main.go`](main.go) | 修改 | HTTP 层添加 PathNormalizeHandler 进行路径规范化 |
| [`middleware/auth.go`](middleware/auth.go) | 修改 | 认证中间件集成路径规范化 |
| [`middleware/distributor.go`](middleware/distributor.go) | 修改 | 分发中间件适配 |
| [`router/relay-router.go`](router/relay-router.go) | 修改 | 路由器重构，支持防呆路径 |
| [`router/web-router.go`](router/web-router.go) | 修改 | Web 路由器适配 |
| [`relay/common/relay_utils.go`](relay/common/relay_utils.go) | 修改 | 添加请求路径规范化和 /v1 前缀自动补全 |
| [`relay/constant/relay_mode.go`](relay/constant/relay_mode.go) | 修改 | Relay 模式常量适配 |

#### 示例
```
# 以下 URL 都能正确路由到 /v1/chat/completions：
/ABC/v1/chat/completions
/我是奶龙/v1beta/models/gemini-pro:generateContent
/test/v1/models
//chat/completions (双斜杠修正)
```

---

### 2. SillyTavern + Google AI Studio 支持

#### 功能描述
- **Gemini 模型列表优化**：Google AI Studio 接口只返回 Gemini 渠道类型的模型
- **格式兼容**：优化 Gemini 模型列表返回格式，支持 SillyTavern 拉取自定义模型
- **渠道类型识别**：根据渠道类型自动选择正确的请求/响应格式

#### 相关文件
| 文件 | 变更类型 | 说明 |
|------|---------|------|
| [`controller/model.go`](controller/model.go) | 修改 | 添加按渠道类型过滤模型的逻辑 |
| [`model/ability.go`](model/ability.go) | 修改 | 新增 `GetGroupEnabledModelsByChannelType` 和 `GetAllGroupsEnabledModelsByChannelType` 函数 |

#### 新增函数

```go
// GetGroupEnabledModelsByChannelType 获取指定分组下指定渠道类型的启用模型
// 用于 Gemini 接口只返回 Gemini 渠道类型的模型
func GetGroupEnabledModelsByChannelType(group string, channelType int) []string

// GetAllGroupsEnabledModelsByChannelType 获取所有分组下指定渠道类型的启用模型（用于 auto 分组）
func GetAllGroupsEnabledModelsByChannelType(groups []string, channelType int) []string
```

---

### 3. Gemini 渠道防呆设计

#### 功能描述
- **OpenAI 兼容反代支持**：当 Gemini 渠道配置为 OpenAI 兼容反代时，正确使用 OpenAI 格式而非 Gemini 原生格式
- **认证头自动切换**：根据渠道类型自动选择 `x-goog-api-key` 或 `Authorization: Bearer`
- **响应处理适配**：上游返回 OpenAI 格式时使用 OpenAI adaptor 处理响应

#### 相关文件
| 文件 | 变更类型 | 说明 |
|------|---------|------|
| [`relay/channel/gemini/adaptor.go`](relay/channel/gemini/adaptor.go) | 修改 | 添加渠道类型判断，实现防呆逻辑 |

#### 核心代码变更

```go
// GetRequestURL - 防呆设计
if info.ChannelType != channelconstant.ChannelTypeGemini {
    // 使用 OpenAI 兼容格式的 URL
    return relaycommon.GetFullRequestURL(info.ChannelBaseUrl, info.RequestURLPath, info.ChannelType), nil
}

// SetupRequestHeader - 防呆设计
if info.ChannelType != channelconstant.ChannelTypeGemini {
    req.Set("Authorization", "Bearer "+info.ApiKey)
} else {
    req.Set("x-goog-api-key", info.ApiKey)
}

// ConvertOpenAIRequest - 防呆设计
if info.ChannelType != channelconstant.ChannelTypeGemini {
    return request, nil  // 直接返回原始 OpenAI 请求
}

// DoResponse - 防呆设计
if info.ChannelType != channelconstant.ChannelTypeGemini {
    openaiAdaptor := openai.Adaptor{}
    return openaiAdaptor.DoResponse(c, resp, info)
}
```

---

### 4. 渠道测试优化

#### 功能描述
- **智能端点类型选择**：根据渠道类型自动选择最优的测试端点格式
- **Gemini 渠道原生测试**：使用 Gemini 原生格式测试 Gemini 渠道
- **请求类型适配**：根据 RelayMode 选择正确的转换函数

#### 相关文件
| 文件 | 变更类型 | 说明 |
|------|---------|------|
| [`controller/channel-test.go`](controller/channel-test.go) | 修改 | 重构渠道测试逻辑，支持多种端点类型 |

#### 新增函数

```go
// endpointTypeToRelayFormat 将端点类型转换为 RelayFormat
// 用于渠道测试时根据渠道类型选择正确的请求格式
func endpointTypeToRelayFormat(endpointType constant.EndpointType) types.RelayFormat
```

---

## 🔧 运维与部署

### 新增部署文件

| 文件 | 说明 |
|------|------|
| [`Dockerfile.cn`](Dockerfile.cn) | 中国区 Docker 镜像构建文件 (使用国内镜像加速) |
| [`docker-compose.local.yml`](docker-compose.local.yml) | 本地开发 Docker Compose 配置 |
| [`build-and-push.bat`](build-and-push.bat) | Windows 一键构建和推送脚本 |
| [`DOCKER_BUILD_GUIDE.md`](DOCKER_BUILD_GUIDE.md) | Docker 构建指南文档 |

### docker-compose.local.yml 特点
- 支持本地开发环境快速部署
- 预配置 MySQL + Redis
- 支持数据持久化

---

## 📝 文档更新

### README.md 改进
- 添加 Demo 版优化特性说明
- 更新项目链接指向 fork 仓库
- 添加 SillyTavern + Google AI Studio 使用说明
- 更新贡献指南和问题反馈链接

---

## 🐛 Bug 修复

| 提交 | 修复内容 |
|------|---------|
| `962039c4` | 修复 Gemini 渠道通过 OpenAI 接口调用失败的问题 |
| `dedbc05d` | 修复渠道测试时 Gemini 渠道使用 OpenAI 格式的问题 |

---

## 📋 完整提交历史

| 提交哈希 | 日期 | 类型 | 说明 |
|---------|------|------|------|
| `c4172fe6` | 2025-12-20 | feat | 优化Gemini模型列表返回格式，支持SillyTavern拉取自定义模型 |
| `2832e0d3` | 2025-12-21 | feat | URL防呆设计 - 支持双斜杠路径规范化 |
| `8028eb2d` | 2025-12-21 | feat | 增强URL防呆设计 - 支持任意路径前缀自动提取API端点 |
| `38bd262e` | 2025-12-21 | docs | 更新 README 适配魔改版项目 |
| `c331bbd5` | 2025-12-21 | docs | 更新 README 为 Demo 版，添加优化特性说明 |
| `89d0adff` | 2025-12-31 | feat | URL防呆设计 - 支持任意路径前缀自动提取API端点 |
| `1ff54a55` | 2025-12-31 | feat | Google AI Studio 接口只返回 Gemini 渠道类型的模型 |
| `dedbc05d` | 2026-01-05 | fix | 修复渠道测试时 Gemini 渠道使用 OpenAI 格式的问题 |
| `962039c4` | 2026-01-05 | fix | 修复 Gemini 渠道通过 OpenAI 接口调用失败的问题 |
| `c4274f24` | 2026-01-05 | merge | 合并上游更新：保留本地Demo版优化特性，兼容新函数签名 |
| `3994e9a9` | 2026-01-06 | merge | 合并上游全部更新：修复默认请求体大小限制为128MB |

---

## 🔀 给作者提交 PR 的建议

### 建议分拆的 PR

基于功能独立性，建议分拆为以下几个 PR：

#### PR 1: URL 防呆设计
**优先级**: ⭐⭐⭐⭐⭐ (高)

**文件**:
- `common/path_normalize.go` (新增)
- `common/path_normalize_test.go` (新增)
- `main.go` (部分修改)
- `middleware/auth.go` (部分修改)

**说明**: 这是一个独立且通用的功能增强，可以解决用户配置 URL 时的常见错误，提升用户体验。

---

#### PR 2: Gemini 渠道 OpenAI 兼容支持
**优先级**: ⭐⭐⭐⭐ (中高)

**文件**:
- `relay/channel/gemini/adaptor.go` (修改)
- `relay/common/relay_utils.go` (部分修改)

**说明**: 使 Gemini 渠道支持 OpenAI 兼容反代，增加部署灵活性。

---

#### PR 3: Gemini 模型列表按渠道类型过滤
**优先级**: ⭐⭐⭐ (中)

**文件**:
- `controller/model.go` (修改)
- `model/ability.go` (修改)

**说明**: 使 Google AI Studio 接口只返回 Gemini 渠道的模型，优化 SillyTavern 等客户端体验。

---

#### PR 4: 渠道测试端点类型优化
**优先级**: ⭐⭐⭐ (中)

**文件**:
- `controller/channel-test.go` (修改)

**说明**: 优化渠道测试逻辑，支持多种端点类型。

---

### 不建议提交的内容 (项目特定)

以下文件是 Demo 版项目特定的配置和文档，不适合提交给原作者：

- `README.md` (Demo 版本说明)
- `Dockerfile.cn` (中国区镜像配置)
- `docker-compose.local.yml` (本地开发配置)
- `build-and-push.bat` (Windows 构建脚本)
- `DOCKER_BUILD_GUIDE.md` (构建指南)

---

## ⚠️ 待清理的无效代码

### Gemini Adaptor 中的渠道类型判断代码

**问题描述**：`relay/channel/gemini/adaptor.go` 中添加了多处渠道类型判断代码，意图是"当渠道类型不是 Gemini 时使用 OpenAI 格式"。

**为什么是死代码**：

调用链分析：
```
用户请求 → 选择渠道 (ChannelType=Gemini) → 转换 APIType → GetAdaptor(APITypeGemini) → gemini.Adaptor
```

由于 adaptor 的选择是基于 `APIType` 的，而 `APIType` 是从 `ChannelType` 转换来的：
- 如果 `ChannelType == ChannelTypeGemini`，则 `APIType == APITypeGemini`，使用 `gemini.Adaptor`
- 如果 `ChannelType != ChannelTypeGemini`，则不会使用 `gemini.Adaptor`

因此，在 `gemini.Adaptor` 内部判断 `info.ChannelType != ChannelTypeGemini` 永远为 false！

**应该移除的代码**：

```go
// ❌ 死代码 - GetRequestURL 中
if info.ChannelType != channelconstant.ChannelTypeGemini {
    return relaycommon.GetFullRequestURL(info.ChannelBaseUrl, info.RequestURLPath, info.ChannelType), nil
}

// ❌ 死代码 - SetupRequestHeader 中
if info.ChannelType != channelconstant.ChannelTypeGemini {
    req.Set("Authorization", "Bearer "+info.ApiKey)
} else {
    req.Set("x-goog-api-key", info.ApiKey)
}

// ❌ 死代码 - ConvertOpenAIRequest 中
if info.ChannelType != channelconstant.ChannelTypeGemini {
    return request, nil
}

// ❌ 死代码 - DoResponse 中
if info.ChannelType != channelconstant.ChannelTypeGemini {
    openaiAdaptor := openai.Adaptor{}
    return openaiAdaptor.DoResponse(c, resp, info)
}
```

**正确理解**：

当用户使用 OpenAI 接口调用 Gemini 渠道时：
1. 请求路径是 `/v1/chat/completions`
2. 渠道类型是 `ChannelTypeGemini`
3. adaptor 是 `gemini.Adaptor`
4. adaptor 的 `ConvertOpenAIRequest` 方法会将 OpenAI 格式转换为 Gemini 格式
5. 发送给上游的是 Gemini 原生格式

**这才是正确的行为**，原版代码已经实现了这个功能，不需要额外的"防呆"逻辑。

---

## 📌 维护说明

### 与上游同步

```bash
# 获取上游更新
git fetch upstream

# 合并上游更新 (保留本地修改)
git merge upstream/main

# 解决冲突后提交
git push origin main
```

### 关键修改点

在合并上游更新时，需特别注意以下文件的冲突处理：

1. **`relay/channel/gemini/adaptor.go`** - 保留渠道类型判断逻辑
2. **`controller/model.go`** - 保留按渠道类型过滤逻辑
3. **`controller/channel-test.go`** - 保留端点类型选择逻辑
4. **`main.go`** - 保留路径规范化中间件
5. **`middleware/auth.go`** - 保留路径提取逻辑

---

*最后更新: 2026-01-06*