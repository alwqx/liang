# Liang: Kubernetes Scheduer Extender

该项目基于[kratos v0.6.0](https://github.com/go-kratos/kratos/tree/v1.0.0)构建，详情参考文档[kratos docs v1](https://v1.go-kratos.dev/#/kratos-tool)

# Env
prometheus: 39.105.5.227:9090 admin/adminlwq
node_exporter: 39.105.5.227:9100 prom/prom

# 改进点
## TOPSIS-MAX
1. 网络负载部分的计算逻辑有问题，k8s-master的带宽是0.5Gbps，三个节点的网速都比master节点大，在计算负载和过滤时不会因为上限80%被过滤掉，影响结果判断

# Reference
- [prom go SDK](https://github.com/prometheus/client_golang)