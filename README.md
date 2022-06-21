# Capa(cloud application api): To be the high-level api layer for all application runtime.

Let the code achieve "write once, run anywhere".

With the help of the Capa project, your applications have the ability to run across clouds and hybrid clouds with small changes.

详细文档请见 [capa-java](https://github.com/capa-cloud/capa-java)

----------------------------------------------------

## Motivation

### Sidecar or SDK ?

如上述文档所述，Capa采用SDK模型。

### Why Runtime as Sidecar ?

capa-runtime为插件性质，不会影响Capa SDK的运行，目前作为实验性质的项目。

当安装capa-runtime后，将会作为sidecar运行。从而提供SDK模型无法支持的API能力。

## Feature

### 7层HTTP流量拦截

基于Iptables，拦截http流量。

> 多个iptables怎么注入？顺序？

实验性质功能，会有影响主链路的风险。

### Actor API

参考 dapr 的 actor 设计。

actor 并不适合集成到SDK中运行，故通过runtime提供actor api。

### Binding API

参考 dapr 的 binding 设计。

binding 作为拓展性质的同外部系统的交互方式，可基于runtime提供弱依赖的交互。

### SaaS API

SaaS api 可以作为实验性质，在runtime中进行提供

----------------------------------------------------

## Archi

### A、控制面

尽可能少的引入依赖项，Capa作为基础能力的聚合层，不再引入额外的控制面。

使用 configuration 组件作为控制面配置下发方式。

### B、形态

#### sidecar

+ 流量拦截
+ grpc交互

#### proxyless



### C、协商机制

应用同Proxy进行协商，以决定 形态。