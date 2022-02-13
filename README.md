Simple MQTT Broker 
============

## 关于
Golang MQTT Broker, Version 3.1.1, and Compatible
for [eclipse paho client](https://github.com/eclipse?utf8=%E2%9C%93&q=mqtt&type=&language=) and mosquitto-client

## 运行
```bash
$ go get github.com/sestack/smq
$ cd $GOPATH/github.com/sestack/smq
$ go run main.go
```

## 用法:
~~~
Usage: smq [options]

Options:
    -c,  --config <file>              Configuration file
    -v, --version                     Show version  
    -h, --help                        Show this message  
~~~

### 配置文件
~~~
#集群名
cluster: smq
#节点id
broker: f7637a07-a081-42d0-bd56-6643b4ee522b
#节点名
name: windows
#工作池大小
worker: 4096
#监听地址
host: 0.0.0.0
#监听端口
port: 1883
#tls设置
tls:
  enable: false
  verify: false
  cert: ssl/server/cert.pem
  key: ssl/server/key.pem
#etcd服务器配置
etcd:
  endpoints:
    - 127.0.0.1:2379
#http服务配置
http:
  enable: true
  host: 0.0.0.0
  port: 8888
#桥接设置
bridge:
  enable: false
  type: kafka
  kafka:
    addr:
      - 127.0.0.1:9092
#通知设置
notify:
  enable: false
  type: grpc
  server: 127.0.0.1:9999
#是否允许匿名客户端连接
auth:
  enable: false
#是否开启访问控制
acl:
  enable: false
#日志级别配置
log:
  level: info
~~~

### 特点和未来

* 支持 QOS 0,1和2

* 支持集群

* 支持分组共享订阅

* 支持保留消息

* 支持遗嘱消息

* 支持TLS/SSL

* 支持客户端认证

* 支持访问控制

* 支持Kafka桥接 

* 支持HTTP接口
	* http://127.0.0.1:8888/api/v1/swagger/index.html

* 支持agent模式

### 集群
~~~
集群功能通过etcd来实现，每个节点注册并监听etcd，节点与节点之间无需建立连接。
~~~

### 分组共享订阅
~~~
                                       [s1]
           msg1                      /
[smq]  ------>  "$share/g1/topic"    - [s2] got msg1
         |                           \
         |                             [s3]
         | msg1
          ---->  "$share/g2/topic"   --  [s4]
                                     \
                                      [s5] got msg1
~~~

### 桥接
~~~
smq实现了将客户端发布给topic的消息直接桥接的对应的kafka topic上。
添加桥接映射：
curl -X 'POST' \
  'http://127.0.0.1:8888/api/v1/bridge/' \
  -H 'accept: application/json' \
  -H 'x-token: <token>' \
  -d '{"topic": "event","queue": "event"}'
  
接收消息：
[smq]  ------>  "$queue/topic"   ------> “kafka topic”
~~~

### agent模式
~~~
agent1				  |----------------------> NodeServer							
  	   \			  |-------------------------|   
		\	          |                         |
agent2 ------------	[smq] ----> kafka ----> JobServer			
		/	    |agents/agent1		 
	   / 	    |agents/agent2			
agent3		   |agents/agent3		

smq作为一个agent连接管理端来使用，用来集中管理agent连接。agent上下线消息可以直接通知到NodeServer。JobServer可以通过agent订阅的topic发送作业给对于的agent，agent执行完成后将消息可以直接写入kafka，JobServer读取kafka里的消息就可以完成异步任务的执行了。                                     
smq在ConnectPacket包后面添加了一个字段可以带上客户端的信息。默认客户端结构体：
type ClientInfo struct {
	HostName     string `json:"host_name"`
	IntranetIP   string `json:"intranet_ip"`
	PublicIP     string `json:"public_ip"`
	Platform     string `json:"platform"`
	Arch         string `json:"arch"`
	OS           string `json:"os"`
	AgentVersion string `json:"agent_version"`
}

agent订阅：
只要客户端订阅或者取消订阅了topic:agents/<clientid>在开启通知的情况下会触发OnAgentConnect或者OnAgentDisconnected消息用于通知服务器agent上线或者离线。
[smq]  ------>  "agents/<clientid>"

agent上下线通知消息格式：
{
  "id": "0e481c4c-9394-4006-9291-fe71a4fe0744",
  "node_id": "f7637a07-a081-42d0-bd56-6643b4ee522b",
  "host_name": "pi.example.com",
  "connect_ip": "192.168.1.100",
  "intranet_ip": "192.168.1.100",
  "public_ip": "",
  "platform": "linux",
  "arch": "aarch64",
  "os": "CentOS Stream release 8",
  "agent_version": "v0.0.1",
  "user_name": "",
  "proto_name": "MQTT",
  "proto_ver": 4,
  "keep_alive": 5,
  "clean_start": true,
  "status": 1,
  "timestamp": 1644739215
}
~~~

### 通知
~~~
目前实现了http和grpc的通知，接口如下:
type Notification interface {
	//客户端上线通知
	OnClientConnect(data interface{}) bool
	//客户端离线通知
	OnClientDisconnected(data interface{}) bool
	//订阅通知
	OnSubscribe(data interface{}) bool
	//取消订阅通知
	OnUnSubscribe(data interface{}) bool
	//Agent上线通知
	OnAgentConnect(data interface{}) bool
	//Agent离线通知
	OnAgentDisconnected(data interface{}) bool
}
~~~

## License

* Apache License Version 2.0


## Reference

* hmq.(https://github.com/fhmq/hmq)
* emqx.(https://github.com/emqx/emqx)
