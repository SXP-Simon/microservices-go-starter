# gRPC Proto 使用说明规范

## 概述

本文档旨在为团队成员提供关于项目中 gRPC 和 Protocol Buffers (Proto) 的使用指南。通过本文档，您将了解如何定义、实现和使用 gRPC 服务。

## 目录

1. [Proto 文件基础](#proto-文件基础)
2. [项目中的 Proto 文件结构](#项目中的-proto-文件结构)
3. [生成 Go 代码](#生成-go-代码)
4. [实现 gRPC 服务](#实现-grpc-服务)
5. [实现 gRPC 客户端](#实现-grpc-客户端)
6. [最佳实践](#最佳实践)
7. [常见问题](#常见问题)

## Proto 文件基础

### 什么是 Protocol Buffers?

Protocol Buffers (Proto) 是 Google 开发的一种语言无关、平台无关、可扩展的序列化结构数据的方法。它比 XML 更小、更快、更简单。

### 基本语法

```protobuf
syntax = "proto3";  // 指定使用 proto3 语法

package package_name;  // 包名，防止命名冲突

option go_package = "path/to/go/package;package_name";  // Go 语言的包路径

// 定义服务
service ServiceName {
  rpc MethodName(RequestMessage) returns (ResponseMessage);
}

// 定义消息
message RequestMessage {
  string field_name = 1;  // 字段类型 字段名 = 字段编号
  int32 number = 2;
  bool flag = 3;
}

message ResponseMessage {
  string result = 1;
  int32 code = 2;
}
```

### 字段规则

- **字段编号**: 每个字段必须有唯一的编号，1-15 使用一个字节，16-2047 使用两个字节
- **字段类型**: 支持 string, int32, int64, bool, float, double, enum, message 等
- **重复字段**: 使用 `repeated` 关键字表示数组或列表

## 项目中的 Proto 文件结构

本项目包含以下 Proto 文件：

### 1. driver.proto

位置: [`proto/driver.proto`](proto/driver.proto)

定义了司机相关的服务:

```protobuf
service DriverService {
  rpc RegisterDriver(RegisterDriverRequest) returns (RegisterDriverResponse);
  rpc UnRegisterDriver(RegisterDriverRequest) returns (RegisterDriverResponse);
}
```

主要消息类型:
- `RegisterDriverRequest`: 注册司机请求
- `RegisterDriverResponse`: 注册司机响应
- `Driver`: 司机信息
- `Location`: 位置信息

### 2. trip.proto

位置: [`proto/trip.proto`](proto/trip.proto)

定义了行程相关的服务:

```protobuf
service TripService {
  rpc PreviewTrip(PreviewTripRequest) returns (PreviewTripResponse);
  rpc CreateTrip(CreateTripRequest) returns (CreateTripResponse);
}
```

主要消息类型:
- `PreviewTripRequest`: 预览行程请求
- `PreviewTripResponse`: 预览行程响应
- `CreateTripRequest`: 创建行程请求
- `CreateTripResponse`: 创建行程响应
- `Trip`: 行程信息
- `Route`: 路线信息
- `RideFare`: 车费信息

### 3. payment.proto

位置: [`proto/payment.proto`](proto/payment.proto)

定义了支付相关的服务:

```protobuf
service PaymentService {
  rpc CreatePaymentSession(CreatePaymentSessionRequest) returns (CreatePaymentSessionResponse);
  rpc GetPaymentByTripId(GetPaymentByTripIdRequest) returns (GetPaymentByTripIdResponse);
  rpc ProcessWebhook(WebhookRequest) returns (WebhookResponse);
}
```

主要消息类型:
- `CreatePaymentSessionRequest`: 创建支付会话请求
- `CreatePaymentSessionResponse`: 创建支付会话响应
- `GetPaymentByTripIdRequest`: 根据行程ID获取支付记录请求
- `GetPaymentByTripIdResponse`: 根据行程ID获取支付记录响应
- `WebhookRequest`: Webhook 请求
- `WebhookResponse`: Webhook 响应

## 生成 Go 代码

### 前置条件

确保已安装以下工具:
- protoc: Protocol Buffers 编译器
- protoc-gen-go: Go 代码生成插件
- protoc-gen-go-grpc: gRPC Go 代码生成插件

### 生成命令

项目提供了 Makefile 来简化代码生成过程:

```bash
make generate-proto
```

这个命令会执行以下操作:
```bash
protoc \
  --proto_path=proto \
  --go_out=. \
  --go-grpc_out=. \
  proto/driver.proto proto/payment.proto proto/trip.proto
```

### 生成的文件

生成后的 Go 文件位于 `shared/proto/` 目录下:
- `shared/proto/driver/`: 司机服务相关代码
- `shared/proto/trip/`: 行程服务相关代码
- `shared/proto/payment/`: 支付服务相关代码

每个服务目录包含:
- `{service}.pb.go`: 消息类型的 Go 实现
- `{service}_grpc.pb.go`: gRPC 服务和客户端的 Go 实现

## 实现 gRPC 服务

### 基本结构

```go
package grpc

import (
  "context"
  "google.golang.org/grpc"
  "google.golang.org/grpc/codes"
  "google.golang.org/grpc/status"
  pb "path/to/generated/proto"
)

// 实现服务结构体
type serviceHandler struct {
  pb.UnimplementedServiceNameServer  // 嵌入未实现的服务器，确保向前兼容
  service ServiceType  // 业务逻辑服务
}

// 创建处理器
func NewGRPCHandler(server *grpc.Server, service ServiceType) *serviceHandler {
  handler := &serviceHandler{
    service: service,
  }
  
  // 注册服务
  pb.RegisterServiceNameServer(server, handler)
  
  return handler
}

// 实现服务方法
func (h *serviceHandler) MethodName(ctx context.Context, req *pb.RequestMessage) (*pb.ResponseMessage, error) {
  // 参数验证
  if req.GetField() == "" {
    return nil, status.Errorf(codes.InvalidArgument, "字段不能为空")
  }
  
  // 调用业务逻辑
  result, err := h.service.BusinessMethod(ctx, req.GetField())
  if err != nil {
    return nil, status.Errorf(codes.Internal, "处理失败: %v", err)
  }
  
  // 构建响应
  return &pb.ResponseMessage{
    Result: result,
    Code:   200,
  }, nil
}
```

### 错误处理

使用 gRPC 预定义的状态码:

```go
import "google.golang.org/grpc/codes"

// 常见状态码
codes.InvalidArgument  // 无效参数
codes.NotFound        // 资源未找到
codes.PermissionDenied // 权限不足
codes.Unauthenticated // 未认证
codes.Internal        // 内部错误
codes.Unavailable     // 服务不可用
```

### 示例: 司机服务实现

参考: [`services/driver-service/grpc_handler.go`](services/driver-service/grpc_handler.go)

```go
type grpcHandler struct {
  pb.UnimplementedDriverServiceServer
  service *Service
}

func (h *grpcHandler) RegisterDriver(ctx context.Context, req *pb.RegisterDriverRequest) (*pb.RegisterDriverResponse, error) {
  driver, err := h.service.RegisterDriver(req.GetDriverID(), req.GetPackageSlug())
  if err != nil {
    return nil, status.Errorf(codes.Internal, "failed to register driver")
  }

  return &pb.RegisterDriverResponse{
    Driver: driver,
  }, nil
}
```

## 实现 gRPC 客户端

### 基本结构

```go
package grpc_clients

import (
  "google.golang.org/grpc"
  "google.golang.org/grpc/credentials/insecure"
  "os"
  pb "path/to/generated/proto"
)

type serviceClient struct {
  Client pb.ServiceNameClient
  conn   *grpc.ClientConn
}

func NewServiceClient() (*serviceClient, error) {
  // 从环境变量获取服务地址
  serviceURL := os.Getenv("SERVICE_URL")
  if serviceURL == "" {
    serviceURL = "default-service:port"
  }

  // 创建连接
  conn, err := grpc.NewClient(serviceURL, grpc.WithTransportCredentials(insecure.NewCredentials()))
  if err != nil {
    return nil, err
  }

  // 创建客户端
  client := pb.NewServiceNameClient(conn)

  return &serviceClient{
    Client: client,
    conn:   conn,
  }, nil
}

func (c *serviceClient) Close() {
  if c.conn != nil {
    c.conn.Close()
  }
}
```

### 使用客户端

```go
// 创建客户端
client, err := NewServiceClient()
if err != nil {
  log.Fatalf("Failed to create client: %v", err)
}
defer client.Close()

// 调用服务方法
ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
defer cancel()

req := &pb.RequestMessage{
  Field: "value",
}

resp, err := client.Client.MethodName(ctx, req)
if err != nil {
  log.Fatalf("Failed to call method: %v", err)
}

log.Printf("Response: %v", resp)
```

### 示例: 行程服务客户端

参考: [`services/api-gateway/grpc_clients/trip_client.go`](services/api-gateway/grpc_clients/trip_client.go)

## 最佳实践

### 1. Proto 文件设计

- **使用明确的字段名**: 避免缩写，使用描述性名称
- **保持向后兼容**: 不要重用字段编号，谨慎删除字段
- **使用合理的字段编号**: 1-15 用于频繁出现的字段
- **添加注释**: 为服务和消息添加清晰的注释

### 2. 错误处理

- **使用适当的状态码**: 根据错误类型选择正确的 gRPC 状态码
- **提供有意义的错误消息**: 错误消息应该清晰、有用
- **记录错误**: 在服务端记录详细的错误信息

### 3. 连接管理

- **复用连接**: 避免为每个请求创建新连接
- **正确关闭连接**: 使用 defer 确保连接被正确关闭
- **设置超时**: 为客户端请求设置合理的超时时间

### 4. 上下文传递

- **传递上下文**: 始终传递 context.Context 参数
- **使用超时**: 为长时间运行的操作设置超时
- **传递元数据**: 使用 context 传递请求 ID、用户信息等

### 5. 测试

- **单元测试**: 为服务处理器编写单元测试
- **集成测试**: 测试客户端和服务器的交互
- **模拟服务**: 使用模拟服务进行测试

## 常见问题

### Q: 如何更新 Proto 文件?

A: 
1. 修改 `.proto` 文件
2. 运行 `make generate-proto` 重新生成代码
3. 更新相关的服务实现和客户端代码
4. 重新编译和部署服务

### Q: 如何处理版本兼容性?

A: 
- 不要更改现有字段的编号
- 添加新字段时使用新的编号
- 保留已弃用的字段，标记为 `reserved`
- 使用 `optional` 关键字标识可选字段

### Q: 如何调试 gRPC 服务?

A: 
- 使用 grpcurl 工具测试服务
- 启用 gRPC 日志记录
- 使用 grpc-ui 进行可视化调试
- 检查网络连接和防火墙设置

### Q: 如何处理大型消息?

A: 
- 考虑使用流式传输
- 分页处理大量数据
- 压缩消息内容
- 调整 gRPC 消息大小限制

## 参考资料

- [Protocol Buffers 官方文档](https://developers.google.com/protocol-buffers)
- [gRPC Go 官方文档](https://grpc.io/docs/languages/go/quickstart/)
- [gRPC 最佳实践](https://grpc.io/docs/guides/)