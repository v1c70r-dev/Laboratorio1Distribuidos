# Laboratorio1Distribuidos
Laboratorio 1 Sistemas distribuidos

# Go
- Actualizar paquete de dependencias en go
    * go mod tidy

# Rabbitmq
- Levantar docker de rabbitmq: 
    docker run -it --rm --name rabbitmq -p 5672:5672 -p 15672:15672 rabbitmq:3.10-management

# grpc proto
- Generar archivos de compilación desde message.proto
    protoc --go_out=. --go_opt=paths=source_relative --go-grpc_out=. --go-grpc_opt=paths=source_relative proto/message.proto

#### Aparte
- crear modulo en go: 
    go mod init <modulename>
    go mod tidy     
- Actualizar paquete de dependencias en go
    * go mod tidy
- Arreglar problema del archivo de proto:
    export PATH="$PATH:$(go env GOPATH)/bin/"