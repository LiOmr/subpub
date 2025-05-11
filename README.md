# subpub
обеспечивает два rpc-метода: publish (unary) и subscribe (server-stream) позволяя публиковать строковые события по ключу и асинхронно доставлять их неограниченному числу подписчиков с сохранением порядка fifo и без взаимных блокировок.
## 1. сборка
```
go build -o pubsub ./cmd/server         
```

## 2. запуск (конфиг в configs/config.yaml)
```
./pubsub
```

## 3. пример
1) подписка
```
grpcurl -plaintext -d '{"key":"demo"}' localhost:8080 api.PubSub/Subscribe
```

2) публикация 
```
grpcurl -plaintext -d '{"key":"demo","data":"hello"}' localhost:8080 api.PubSub/Publish
```

(появится {"data":"hello"})

## паттерны
- Dependency Injection (через конструкторы) , см internal/service/pubsub.go, internal/handler/grpc/server.go, cmd/server/main.go
- Graceful Shutdown, см cmd/server/main.go
