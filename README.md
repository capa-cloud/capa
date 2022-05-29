# Capa(cloud application api): To be the high-level api layer for all application runtime.

Let the code achieve "write once, run anywhere".

With the help of the Capa project, your applications have the ability to run across clouds and hybrid clouds with small changes.

详细文档请见 [capa-java](https://github.com/capa-cloud/capa-java)

## Motivation

### Sidecar or SDK ?

如上述文档所述，Capa采用SDK模型。

### Why Runtime as Sidecar ?

capa-runtime为插件性质，不会影响Capa SDK的运行，目前作为实验性质的项目。

当安装capa-runtime后，将会作为sidecar运行。从而提供SDK模型无法支持的API能力。

## Feature

### Actor API

参考 dapr 的 actor 设计。

actor 并不适合集成到SDK中运行，故通过runtime提供actor api。


