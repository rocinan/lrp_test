
# <div align="center"><a title="NexT website repository"> Lrp</div>

<p align="center">
  A lightweight reverse proxy SDK
<br>
  <img src="https://img.shields.io/badge/golang-v1.15.5-blue">
  <img src="https://img.shields.io/cii/level/2625?style=flat-square" title="Core Infrastructure Initiative Best Practices">
<br>
</p>

## 整体结构

![lrp](http://lev-nas.oss-cn-shenzhen.aliyuncs.com/pic_bed/lrp.png)

## 通讯协议

```
+-------+-------+----------+
|  Len  |  CMD  |   DATA   |
+-------+-------+----------+
|   8   |   1   | Len - 9  |
+-------+-------+----------+
````
- len : payload长度，int 占8字节
- cmd :
    - 0x00 : client  向  server 发起反向代理请求可用端口
    - 0x01 : server  向  client 返回请求结果
    - 0x02 : client <->  server 数据传输
    - 0x03 : client <->  server 处理链接释放

### 0x00 请求可用端口
```
+-------+
|  CMD  |
+-------+
|   1   |
+-------+

[0]
```

### 0x01 Client 接收请求结果
```
+-------+---------+-------+
|  CMD  |  STATUS |  PORT |
+-------+---------+-------+
|   1   |    1    |   2   |
+-------+---------+-------+
```
- status: 0 -> false | 1 -> true 

### 0x02 传输转发数据

```
+-------+-------+--------+
|  CMD  |  CID  |  DATA  |
+-------+-------+--------+
|   1   |   12  | LEN-13 |
+-------+-------+--------+
```
- cmd： 0x02
- cid： 标识传输的id 12字节大小
- data：传输的代理数据

### 0x03 处理链接释放
```
+-------+-------+--------+
|  CMD  |  TYPE |   CID  |
+-------+-------+--------+
|   1   |   1   |    12  |
+-------+-------+--------+
```
- cmd:  0x03
- type: 0 -> 关闭一条连接 | 1 -> 关闭客户端
- cid : 标识连接的id
