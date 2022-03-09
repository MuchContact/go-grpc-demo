# gRPC

## GO 语言示例

该示例中所有用到的代码文件，可以通过下面的方式获得

```bash
$ git clone https://github.com/crazygit/go-grpc-demo.git
```

### 环境准备

#### 安装`Protocol buffer`编译器

mac下可以直接使用下面的命令安装编译器，其它平台可以[参考文档](https://github.com/protocolbuffers/protobuf#protocol-compiler-installation)。

```bash
$ brew install protobuf
```

检查编译器安装是否成功

```bash
$ protoc --version
libprotoc 3.17.3
```

#### 安装`Protocol buffer`编译器的`Go`插件

```bash
go get github.com/gogo/protobuf/proto
go get github.com/gogo/protobuf/protoc-gen-gogofaster
go get github.com/gogo/protobuf/gogoproto
```

安装好之后，配置`PATH`环境变量

```bash
$ export PATH="$PATH:$(go env GOPATH)/bin"
```


### 创建项目

```bash
$ mkdir go-grpc-demo
$ cd go-grpc-demo
$ go mod init github.com/crazygit/go-grpc-demo
```

### 创建约定的服务文件
`protos`文件由于要被服务端和客户端共同使用，一般情况下，可以[用单独的一个库来保存](https://www.bugsnag.com/blog/libraries-for-grpc-services)，这里为了便于演示，都保存在`protos`目录下。

`protos/greeting.proto`

```protobuf
syntax = "proto3";

// 生成的go代码使用的包信息
option go_package = "gen/greeting";

// protocol buffers定义的包信息，与生成的代码没有关系
package helloworld;

// The greeting service definition.
service Greeter {
  // Sends a greeting
  rpc SayHello (HelloRequest) returns (HelloReply) {}
}

// The request message containing the user's name.
message HelloRequest {
  string name = 1;
}

// The response message containing the greetings
message HelloReply {
  string message = 1;
}
```

上面的文件定一个了一个`Greeter`服务，里面有一个`SayHello`方法，它接收`HelloRequest`作为参数，返回`HelloReply`

### 编译`protos/greeting.proto`文件，生成让服务端和客户端调用的代码

这里我们的服务端都客户端代码都是用`Go`实现。所以使用`--gogofaster_out`文件。

```bash
$ protoc -I.:${GOPATH}/src  --gogofaster_out=plugins=grpc:. protos/greeting.proto
```

执行完上面的步骤之后，在`gen`目录下生成了一个文件。

```
$ tree gen
gen
└── greeting
    ├── greeting.pb.go
```

**PS**:

`gen`目录一般可以添加到`.gitignore`文件里忽略掉。这里处于演示目的都一并提交在代码库里。

### 编写服务端代码

编写服务端代码`cmd/geeting_server/main.go`

```go
package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net"

	pb "github.com/crazygit/go-grpc-demo/gen/greeting"
	"google.golang.org/grpc"
)

var (
	port = flag.Int("port", 50051, "The server port")
)

// server is used to implement helloworld.GreeterServer.
type server struct {
	pb.UnimplementedGreeterServer
}

func (s *server) SayHello(ctx context.Context, in *pb.HelloRequest) (*pb.HelloReply, error) {
	return &pb.HelloReply{Message: "Hello " + in.GetName()}, nil
}

func main() {
	flag.Parse()
	lis, err := net.Listen("tcp", fmt.Sprintf("localhost:%d", *port))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	s := grpc.NewServer()
	pb.RegisterGreeterServer(s, &server{})
	log.Printf("server listening at %v", lis.Addr())
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
```

### 编写客户端代码

编写客户端代码`cmd/geeting_client/main.go`

```go
package main

import (
	"context"
	"flag"
	"log"
	"time"

	"google.golang.org/grpc"
	pb "github.com/crazygit/go-grpc-demo/gen/greeting"
)

const (
	defaultName = "world"
)

var (
	addr = flag.String("addr", "localhost:50051", "the address to connect to")
	name = flag.String("name", defaultName, "Name to greet")
)

func main() {
	flag.Parse()
	// Set up a connection to the server.
	conn, err := grpc.Dial(*addr, grpc.WithInsecure())
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	defer conn.Close()
	c := pb.NewGreeterClient(conn)

	// Contact the server and print out its response.
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	r, err := c.SayHello(ctx, &pb.HelloRequest{Name: *name})
	if err != nil {
		log.Fatalf("could not greet: %v", err)
	}
	log.Printf("Greeting: %s", r.GetMessage())
}
```

### 分别运行服务端和客户端

#### 运行服务端
服务端默认监听在`50051`端口

```bash
$ go mod tidy
$ go run cmd/greeting_server/main.go
2021/12/01 11:32:19 server listening at 127.0.0.1:50051
```

#### 运行客户端

```bash
$ go run cmd/greeting_client/main.go
2021/12/01 11:32:26 Greeting: Hello world
```

### 调试gPRC
参考:

[gRPCui: Don’t gRPC without it!](https://www.fullstory.com/blog/grpcui-dont-grpc-without-it/)

在调试`HTTP`接口时。可以使用`postman`这样的工具方便接口调用。gPRC里也有类似的调试工具:

* [grpcurl](https://github.com/fullstorydev/grpcurl)
* [grpcui](https://github.com/fullstorydev/grpcui)

两个调试工具的作者都是同一个人，区别在于`grpcurl`是命令行模式，而`grpcui`有图形界面。

这里主要介绍`grpcui`的使用。

#### 安装`grpcui`

```bash
$ go install github.com/fullstorydev/grpcui/cmd/grpcui@latest
```

#### 启动服务

首先，需要确保我们的`greeting_server`仍然处于运行状态. 因为`grpcui`在一启动的时候会去连接我们的服务。`localhost:50051`为`greetings_server`的监听地址。

```bash
$ grpcui -plaintext localhost:50051
Failed to compute set of methods to expose: server does not support the reflection API
```

可以看到启动失败，错误提示信息为`Failed to compute set of methods to expose: server does not support the reflection API`。
意思是我们的服务不支持[server reflection](https://github.com/grpc/grpc/blob/master/src/proto/grpc/reflection/v1alpha/reflection.proto).

为什么会这样？想想使用`postman`时的场景，在我们不知道服务端接口定义的情况下，我们可以导入类似`swagger.json`文件，告诉`postman`服务端接口的调用方式。

我们目前还没有告诉`grpcui`,哪里去找它需要的类似`swagger.json`这样的文件。很容易想到，我们前面不是定义了`protos/greeting.proto`文件，它的作用和`swagger.json`是一样的。

```bash
$ grpcui -proto greeting.proto -import-path=protos/ -plaintext localhost:50051
```

选项解释

```
-proto value
    	The name of a proto source file. Source files given will be used to
    	determine the RPC schema instead of querying for it from the remote
    	server via the gRPC reflection API. May specify more than one via
    	multiple -proto flags. Imports will be resolved using the given
    	-import-path flags. Multiple proto files can be specified by specifying
    	multiple -proto flags. It is an error to use both -protoset and -proto
    	flags.

-import-path value
    	The path to a directory from which proto sources can be imported,
    	for use with -proto flags. Multiple import paths can be configured by
    	specifying multiple -import-path flags. Paths will be searched in the
    	order given. If no import paths are given, all files (including all
    	imports) must be provided as -proto flags, and grpcui will attempt to
    	resolve all import statements from the set of file names given.
```

执行上面的命令后。会在本地打开浏览器，并自动弹出下面的界面。界面的使用比较简单，可以参考[插件文档](https://github.com/fullstorydev/grpcui)。

![gPRC UI](https://images.stdcdn.com/2021/12/2eff8921aa969de008f1441aad63d3d6.png)


### 在服务端启用`reflection API`

再来看看我们之前直接通过命令启动`grpcui`的报错信息。

```bash
$ grpcui -plaintext localhost:50051
Failed to compute set of methods to expose: server does not support the reflection API
```

虽然在前面的步骤中，我们通过指定`-proto`和`-import-path`选项让报错不再提示。但是实际上并没有解决报错里提到的问题。

想想我们在前面使用插件的时候，必须要将`protos/greeting.proto`文件信息提供给`grpcui`，`grpcui`才能使用，实际开发时，这个文件的内容可能经常改变的，`grpcui`难道要不断的从我们这里获取新版本的文件吗？这样也太不方便了，因此错误信息里提到的[`Server Reflection API`](https://github.com/grpc/grpc/blob/master/doc/server-reflection.md)便派上用场了，开启这个服务之后, `grpcui`即使在没有`protos/greeting.proto`文件的情况下，也可以方便的调用服务端的接口。

#### 修改服务端的代码

开启这个服务也比较简单，参考[RPC Server Reflection Tutorial](https://github.com/grpc/grpc-go/blob/master/Documentation/server-reflection-tutorial.md), 只需要在服务端添加几句代码就可以了。客户端不用做任何调整。

```diff
--- a/cmd/greeting_server/main.go
+++ b/cmd/greeting_server/main.go
@@ -9,6 +9,7 @@ import (

 	pb "github.com/crazygit/go-grpc-demo/gen/greeting"
 	"google.golang.org/grpc"
+	"google.golang.org/grpc/reflection"
 )

 var (
@@ -32,6 +33,8 @@ func main() {
 	}
 	s := grpc.NewServer()
 	pb.RegisterGreeterServer(s, &server{})
+	// Register reflection service on gRPC server.
+	reflection.Register(s)
 	log.Printf("server listening at %v", lis.Addr())
 	if err := s.Serve(lis); err != nil {
 		log.Fatalf("failed to serve: %v", err)
```

修改完成后，重新启动服务端

```bash
$ go mod tidy
$ go run cmd/greeting_server/main.go
```

再打开`grpcui`

```bash
$ grpcui -plaintext localhost:50051
```

调试工具成功启动，报错信息消失不见了。

最后，我们总结下`grpcui`的使用方法:

1. 如果服务端启用了`Server Reflection`，则可以直接启动调试工具。
2. 如果服务端没有启用`Server Reflection`，那么我们可以通过指定`protos`文件来使用。
3. 还有一种前面没有介绍过的，就是`grpcui`还支持使用指定的[`protoset-files`](https://github.com/fullstorydev/grpcui#proto-source-files)来使用.具使用方式参考[文档](https://github.com/fullstorydev/grpcui#proto-source-files)即可。
