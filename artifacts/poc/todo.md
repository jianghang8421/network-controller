- 创建workload ，去掉podid
- macvlan subnet ，ip range
- 逗号分隔，pod 多个ip 多个mac
- pod update
- cni static-ipam del
- docs flannel默认出口 eth0
- docs 配置权限 macvlansubnets macvlanips
- owner reference
- ip 正则， 192.168.1.1-192.168.1.1-xxxxx
- 校验pod数量


# 前端

## 扁平网络部分

- 重复vlan前端校验规则

macvlansubnet的list接口，可以传递使用lableSelector 筛选master和vlan，当已经存在vlan与master都同名的item时，不允许创建

- 创建subnet中增加ip range

UI加入一个可选的ip段标签栏，允许输入两个值start-end，发送给后端请求时，转换进新增的ranges字段中：

```
"spec": {}
        "cidr": "192.168.56.0/24",
        "master": "enp0s9",
        "mode": "bridge",
        "ranges": [
            {
                "rangeEnd": "192.168.56.20",
                "rangeStart": "192.168.56.10"
            },
            {
                "rangeEnd": "192.168.56.40",
                "rangeStart": "192.168.56.30"
            }
        ]
    }

- 
```



## 创建workload部分

- 查看deployment的信息显示mac为 n/a

mac字段为空字符串时，前端可以补auto，跟ip保持一致。（后端现在兼容了”“和”auto“两种）

- ”CIDR“字段变更

CIDR 字段 变更为IP，支持三种情况：

```
1. 空，不填，这时候跟之前一样，前端补充为”auto“
2. 192.168.1.1 格式的单个ip
3. 多个ip，后端需要的形式为 ”192.168.1.1-192.168.1.2-192.168.1.3“ ，用”-“分隔的多个ip
多个ip在ui上的交互展现待考虑

```

- 修改spec template中的annotation key

加入namespace前缀了

```
	cidr         macvlan.pandaria.cattle.io/ip
	subnet       macvlan.pandaria.cattle.io/subnet
	mac          macvlan.pandaria.cattle.io/mac
```

