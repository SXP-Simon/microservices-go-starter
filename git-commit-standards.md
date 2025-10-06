# Git提交规范

## 提交信息格式

我们采用基于[Conventional Commits](https://www.conventionalcommits.org/zh-hans/)的提交信息格式，但使用中文描述。

### 格式
```
<类型>(<范围>): <描述>

[可选的正文]

[可选的脚注]
```

### 类型
- `feat`: 新功能
- `fix`: 修复bug
- `docs`: 文档更新
- `style`: 代码格式化（不影响功能）
- `refactor`: 重构代码（既不是新功能也不是修复）
- `perf`: 性能优化
- `test`: 添加或修改测试
- `chore`: 构建过程或辅助工具的变动
- `ci`: CI/CD相关配置

### 范围
- `api-gateway`: API网关
- `driver-service`: 司机服务
- `trip-service`: 行程服务
- `payment-service`: 支付服务
- `shared`: 共享组件
- `events`: 事件处理
- `websocket`: WebSocket处理
- `infra`: 基础设施配置

### 示例

#### 新功能
```
feat(events): 添加RabbitMQ事件发布器实现

- 实现RabbitMQPublisher接口
- 添加事件和命令发布方法
- 支持自动重连机制
```

#### 修复bug
```
fix(driver-service): 修复司机注册时的内存泄漏问题

- 确保在连接关闭时正确清理资源
- 添加连接池管理
```

#### 文档更新
```
docs(readme): 更新项目启动说明

- 添加RabbitMQ依赖说明
- 更新环境变量配置文档
```

#### 重构
```
refactor(trip-service): 重构行程服务的事件处理逻辑

- 将事件发布逻辑抽取到独立模块
- 简化服务层的职责
```

## 提交频率

- 每完成一个小功能就提交一次
- 保持每次提交的变更量适中
- 确保每次提交都能通过测试
- 避免将不相关的修改混在一个提交中

## 分支策略

- `main`: 主分支，始终保持稳定可运行状态
- `develop`: 开发分支，集成最新功能
- `feature/*`: 功能分支，开发特定功能
- `hotfix/*`: 热修复分支，修复紧急问题

## 提交前检查清单

- [ ] 代码已通过所有测试
- [ ] 提交信息符合规范
- [ ] 没有调试代码和注释
- [ ] 代码格式化完成
- [ ] 相关文档已更新