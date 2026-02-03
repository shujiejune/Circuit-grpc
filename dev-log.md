1. A microservice in gRPC
In a monorepo structure like this, a microservice spreads across the directory.
It is a vertical slice composed of 3 parts:
- contract (api/proto/user): the public API interface
  - blueprint (user.proto): the Swagger/OpenAPI spec + route definitions
  - data structure (user.pb.go): the DTOs
  - connection logic (user_grpc.pb.go): the routing logic, e.g. e.POST("/login", userService.Login)
- logic (internal/modules/users): the private implementation
  - controller (grpc_handler.go)
  - business logic (service.go)
  - database layer (repository.go)
- entry point (cmd/user): the executable that starts the server

2. Why have User models in both api/proto/user/user.proto and internal/modules/users/domain/user.go?
The proto model (user.proto) is the public view (wire format).
- what the frontend and other microservices see
- optimized for network transmission (JSON/Binary)
- hides sensitive data (e.g. password_hash, deleted_at)
The domain model (user.go) is the private view (database entity).
- represents database schema
- includes ORM tags (GORM or db tags)
- holds sensitive fields
