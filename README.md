# Liang
>A Kubernetes Scheduer Extender with two Customed Scheduling Algorithms BNP and CMDN.

Kubernetes has become the de-facto container cluster management system with its rich enterprise and production-level features. However, Kubernetes' default scheduling algorithms are designed for general scenarios. For customed scenarios, this project proposes a scheduler extender named `Liang` and implements two example algorithms BNP and CMDN.

![liang architecture](assets/images/liang-arc.png)

Config file for kubernetes scheduler is:
```json
{
    "kind": "Policy",
    "apiVersion": "v1",
    "extenders": [
        {
            "urlPrefix": "http://localhost:8000/v1",
            "prioritizeVerb": "prioritizeVerb",
            "weight": 1,
            "enableHttps": false,
            "httpTimeout": 1000000000,
            "nodeCacheCapable": true,
            "ignorable": false
        }
    ]
}
```

# Quick Start
**Note:** This section used default data for run liang, it's just a demo. You should not change default config in `config/` directions.

1. build binary
   ```shell
    go build cmd/main.go
   ```
2. run binary in shell
   ```shell
    ./main -conf configs -log.v 7
   ```
3. request `localhost:8000/v1/prioritizeVerb` use cURL
    ```shell
    curl --location --request POST 'localhost:8000/v1/prioritizeVerb' \
    --header 'Content-Type: application/json' \
    --data-raw '{
        "Pod": {
            "metadata": {
                "creationTimestamp": null,
                "annotations": {
                    "LiangNetIO": "80"
                }
            },
            "spec": {
                "containers": null
            },
            "status": {}
        },
        "Nodes": null,
        "NodeNames": [
            "node1",
            "node2",
            "node3"
        ]
    }'
    ```

You will see sample output like:
```json
[{"Host":"node1","Score":0},{"Host":"node2","Score":51},{"Host":"node3","Score":100}]
```

# Customed Kubernetes Scheduling Algorithm
## Balanced NetIO Priority (BNP)
BNP adds network IO resource request and combines the network information of candidate nodes to select the best node. BNP makes the overall network IO usage of the cluster more balanced and reduces the container deployment time.

![](assets/images/bnp-net-by-pods.png)

## CMDN
Multi-criteria resources scheduling algorithm CMDN is based on TOPSIS decision algorithm. The CMDN algorithm takes CPU utilization, memory utilization, disk IO, network IO and NIC bandwidth of candidate nodes into account. It then scores nodes comprehensively using TOPSIS algorithm which brings two scheduling effects of multi-criteria resource balancing and compactness.

![](assets/images/cmdn-min-cpu-by-pods.png)

The experiments show that BNP Algorithm improves the balance level of cluster  network IO, prevents nodes from network IO bottlenecks, and also reduces the container deployment time by 32%. The CMDN Algorithm can balance the multi-criteria resource utilization such as CPU, Memory, disk IO and network IO of the cluster nodes in balancing policy. It also reduces container deployment time by 21%. The CMDN Algorithm can schedule containers to the nodes with high multidi-criteria resource utilization in the compact policy which achieves the expected results.

# Deploy
1. compile binary
   ```shell
   go build cmd/main.go
   ```
2. change configs in configs/applications.toml if needed
3. run binary
   ```shell
    ./main -conf configs
   ```

# Reference
- [prom go SDK](https://github.com/prometheus/client_golang)
- [kratos v0.6.0](https://github.com/go-kratos/kratos/tree/v1.0.0)
- [kratos v1 docs](https://v1.go-kratos.dev/#/kratos-tool)