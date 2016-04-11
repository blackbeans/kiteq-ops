# kiteq-ops

#### 简介
    - 提供kiteq的系统指标的图标展示
    - 节点存活
    - kiteq系统指标告警(需要对接自己的告警平台)
    
#### 对接告警平台：
    实现[kiteq-ops/app/models/alarms/alarm_entry.go](alarm_entry.go)的WrapAlaramParams方法，期望返回报警系统的完整的URL。即可实现和报警系统对接
    
#### 安装：
    sh build.sh

    revel run kiteq-ops

#### demo

![image](./doc/home.png)

![image](./doc/hours.png)

![image](./doc/days.png)
