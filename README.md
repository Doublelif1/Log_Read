# Log_Read
     基于zookeeper,kafka,elasticsearch,kibana,etcd,Gin框架的日志收集系统。(Gin,Logagent,Logtransfer)(超详细中文注释)
## 架构图
 <div align="center">
  <img src="https://github.com/Doublelif1/Log_Read/blob/master/readme_photo/a.jpg">
  </div>

## 1.	启动系统
        启动zookeeper,kafka,elasticsearch,kibana,etcd

## 2.	修改配置文件
       在Log_agent和Log_transfer两个文件夹中的config修改以上需要连接的系统的地址

## 3.	启动Log_agent和Log_transfer (运行对应main文件)
    启动Log_agent 和Log_transfer时会显示_对应的_需要被拉取数据的etcd的key
  <div align="center">
  <img src="https://github.com/Doublelif1/Log_Read/blob/master/readme_photo/e1.png">
  </div>

   <div align="center">
  <img src="https://github.com/Doublelif1/Log_Read/blob/master/readme_photo/e2.png">
  </div>


## 4.	启动Log_agent和Log_transfer完成之后 就可以实现：
        通过对etcd对应key-value数据的更改:
        (1) 实现实时改变_要将消息发往kafka指定topic的日志文件(Log_agent)  
        (2) 实现实时改变_要从kafka中拉取消息的topic.将topic下的消息读取出来存储到elasticsearch (Log_tranfer)
  
## 5.	为了方便修改etcd的key-value
        我编写了极简的（后期有时间会对前端进行优化）基于Gin框架的客户端,来方便对etcd上的key-value数据进行增删改查:
   <div align="center">
  <img src="https://github.com/Doublelif1/Log_Read/blob/master/readme_photo/web.png">
  </div>
  
        运行步骤:
           (1)启动Gin目录下的main文件
           (2)启动Gin/templates/etcd.html文件
           (3)在网页中实时更改etcd的数据
 

## 6．最终实现
         最后通过在日志文件中插入数据 , 就可在elasticsearch-kibana中可视化的搜索以及查看数据信息.

# 总结
         本系统的优点为:
         1.	Log_agent和Log_transfer可以单独使用,只需要改变对应系统的etcd的key值即可实现 
         2.	Log_agent可实现同时读取多个日志文件的消息发送到kafka
         3.	Log_transfer可实现同时读取kafka中多个topic的消息发送到elasticsearch中以方便进行可视化查找收集信息.
         4.  通过客户端网页便捷的对etcd数据进行修改
