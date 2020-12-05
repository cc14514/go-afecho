# go-afecho

## 关于 AF_INET 和 AF_UNIX

进程间通信在不强调性能的情况下是一个很简单的问题，
通常会选用 `RPC(AF_INET)` 来实现，其通信原理大致如下：

![AF_INET](https://gitee.com/cc14514/statics/raw/master/utopia/images/af/AF_INET.jpg)

可以看到每个数据包都会经过 `TCP/IP` 协议栈，性能损耗主要在此，
在不改变接口的情况下使用 `IPC(AF_UNIX)` 来实现进程间通信效率会有很大提升，
如下图所示 `AF_UNIX` 直接通过内核缓冲区`copy`数据，省去了`TCP/IP`协议栈的工作

![AF_UNIX](https://gitee.com/cc14514/statics/raw/master/utopia/images/af/AF_UNIX.jpg)

内核为两种方式提供了相同的接口: `socket / bind / connect / accept / send / recv` 
处理方式也是完全一致的：`select / poll / epoll` 

两种方式可以无成本的轻松进行切换，`AF_UNIX`特别适合异构模块本地进程间通信，
下文对此进行了简单的测试，使用 `golang` 编写了 `echo server`，并提供了 `golang` 和 `java` 两种客户端

## 启动 echoserver

```bash
go-afecho$> go run server/main.go
--> tcp :  <nil> [::]:12345
--> ipc :  <nil> /tmp/afecho.ipc
```

## 使用 client 测试

测试 50w 条 "HelloWorld" 消息，对比两种通道的处理能力

### go


>多次测试数据会有所波动，可以看出用 go 实现的 AF_UNIX / AF_INET 
>处理相同的工作性能相差 5 倍以上（理论上应该是10倍甚至更多）, 具体测试脚本如下：

```bash
go-afecho$> go run client/go/main.go --help
  -a string
    	unix: /tmp/afecho.ipc , tcp: localhost:12345 (default "/tmp/afecho.ipc")
  -c int
    	total count of message (default 100000)
  -m string
    	HelloWorld (default "HelloWorld")
  -n string
    	netrowk : tcp / unix (default "unix")
```

#### AF_UNIX

client 端：

```bash
go-afecho$> go run client/go/main.go -n unix -c 500000 -m HelloWorld
normal quit.
net=unix , msg.size=10 , r=500000 ,w=500000 , time=1.430734597s avg=349470.825021/s

//server 输出：
--accept--> /tmp/afecho.ipc
<--close-- /tmp/afecho.ipc  500000 1.449517214s 344942.43681330985
```

#### AF_INET

client 端：

```bash
go-afecho$> go run client/go/main.go -n tcp -a localhost:12345 -c 500000 -m HelloWorld
normal quit.
net=tcp , msg.size=10 , r=500000 ,w=500000 , time=8.105451805s avg=61686.875948/s

//server 输出：
--accept--> 127.0.0.1:12345 127.0.0.1:54276
<--close-- 127.0.0.1:12345 127.0.0.1:54276 500000 8.127940457s 61516.19867852091
```

### java

>与`golang`不同，java不能直接做系统调用，需要使用`JNI/JNR`方式调用`c`编写的库来中转一下，
>过程中应会损失一些性能，完成下面的测试需要先使用 `maven` 编译 `java client` 具体操作如下：

```bash
$> cd go-afecho/client/java/echo-tester
$> mvn package
......
[INFO] ------------------------------------------------------------------------
[INFO] BUILD SUCCESS
[INFO] ------------------------------------------------------------------------
[INFO] Total time:  3.627 s
[INFO] Finished at: 2020-12-05T11:26:31+08:00
[INFO] ------------------------------------------------------------------------
```

#### AF_UNIX

```bash
$> java -jar target/echo-tester-1.0-SNAPSHOT.jar unix 500000
1607138984673
done : net=unix , msg.size=12 , total=500000 , time=1543ms , avg=324044.069994/s

// server 输出：
--accept--> /tmp/afecho.ipc
<--close-- /tmp/afecho.ipc  500000 1.542095908s 324234.0488721406
```

#### AF_INET

```
$> java -jar target/echo-tester-1.0-SNAPSHOT.jar tcp 500000
done : net=tcp , msg.size=12 , total=500000 , time=4830ms , avg=103519.668737/s

// server 输出:
--accept--> 127.0.0.1:12345 127.0.0.1:54359
<--close-- 127.0.0.1:12345 127.0.0.1:54359 500000 4.832523878s 103465.60361061915
```

## 结果

|Client|AF_UNIX|AF_INET|AF_UNIX/AF_INET|
|--|--|--|--|
|go|349470/s|61686/s|5.7|
|java|324044|103519/s|3.1|

结果只为表明`AF_UNIX`速度更快，看上去 go 的 tcp 处理比 java 慢纯属测试代码的问题😄

* 测试代码：https://github.com/cc14514/go-afecho
