# Gj-galaxy
= 分布式消息同步系统

## 内部组件
* socket：提供客户端统一的访问接口，实现tcp+udp调度的方案
* restful：提供对外接口用于修改或访问资源信息
* room：基于房间的消息同步，以房间为单位进行分布式设计
* scene：基于场景设计的同步模块
    * sync：同步逻辑处理，支持中间件
* platform：平台功能集成
    * app：已应用为单位进行功能划分

## Server插件机制
* 提供服务的Register接口，在preRun之后Run前进行扩展服务的注册，注意顺序
* 服务可以将需要的config提供出来，进行全局config配置
* 服务名称全局唯一，服务注册有顺序
* 服务生命周期：
    * GetConfig=>提供配置信息，json标志用于配置文件注入
    * OnCreate=>服务初始化，按顺序执行，验证config，初始化内部组件，并提供依赖组件到server
    * OnStart=>已初始化完毕，所有服务依次开启，和注册顺序相反
    * OnStop=>接收到退出命令，所有服务依次关闭，和注册顺序相反

## web服务扩展
* 全局web提供了http服务
* 可以按需添加路由扩展

## socket服务扩展
* 全局socket服务提供了tcp+udp的调度方案
* 可以按namespace的方式注册到消息中

## 场景扩展
* 提供全同步和范围同步
* 事件直接同步，结果计算同步
* 进行道具数验证：个数，有效时间，道具类型（物品栏道具，状态效果，血量，能量值等），道具状态（可用，不可用）
* 服务端基于事件进行效果处理：返回道具的更改信息
* 服务端进行事件注册：事件，对象，目标对象（们：范围），不注册的事件不计算结果
* 服务端进行道具状态监听：当道具达到触发条件，调用监听，触发有依赖，或递归
