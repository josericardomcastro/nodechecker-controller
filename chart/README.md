# NodeChecker Controller Chart

Chart of installation of nodechecker-controller, native kubernetes controller that checks features and connections of kubernetes nodes


## Installing via chart

Clone the last version of the project and use helm to install the controller into your cluster.

```
$ git clone https://github.com/josericardomcastro/nodechecker-controller.git
$ cd nodechecker-controller/chart
$ kubectl create namespace nodechecker-controller
$ helm install nodechecker-controller . -n nodechecker-controller
```

## Creating nodecheckers

The controller works with the custom CRD `kind: NodeChecker` to create nodecheckers.

Nodecheckers are rules to validate node connectivity to another nodes, external host or clusters endpoints. To validate node features, it possible to create rules to check if the node has `AVX`, `GPU` or another feature.

Examples:
- [Connection checker](../manifests/sample-connection.yaml)
- [Feature checker](../manifests/sample-feature.yaml)


### Connection checker

Validate node connectivity to another nodes, external host or clusters endpoints


Parameter | Description 
--------- | ----------- 
`type`    | Type of checker (`connection` or `feature`)
`schedule`| Scheduler for check execution,  it takes a Cron format string, such as `"* * * * *"` or `"@hourly"`
`sourceNodes.matchLabels` | Labels selector for identify the source node of the connection, set `{}` for select all nodes


#### Cluster Node Connectivity

Check connections to a node in the same cluster.

```
apiVersion: nodecontroller.io/v1
kind: NodeChecker
metadata:
  name: worker-outbound
spec:
  type: connection
  schedule: "* * * * *"
  sourceNodes:
    matchLabels:
      node-role.kubernetes.io/worker: "true"
  connectionCheck:
    clusterEndpoints: []
    external: []
    clusterNodes:
      - name: masters-apiserver
        matchLabels:
          node-role.kubernetes.io/control-plane: "true"
        port: 443
        protocol: TCP
        syncLabels: 
          connection.nodechecker/controlplane-nodes: "true"
```

Parameter | Description 
--------- | ----------- 
`clusterNodes` | Array for rules to check connectivity to cluster nodes 
`clusterNodes[].name`    | Name of the checker rule
`clusterNodes[].matchLabels` | Labels selector for identify the destination nodes
`clusterNodes[].port` | Port of the destination nodes
`clusterNodes[].protocol` | Protocol of the connection (`TCP` or `UDP` )
`clusterNodes[].syncLabels` | Labels setted for the `source node` if connection works, if empty `syncLabels: {}` nothing will happen


#### External Host Connectivity

Check connections to a external existing host (localhost, xpto-vm or google.com)

```
apiVersion: nodecontroller.io/v1
kind: NodeChecker
metadata:
  name: worker-outbound
spec:
  type: connection
  schedule: "* * * * *"
  sourceNodes:
    matchLabels:
      node-role.kubernetes.io/worker: "true"
  connectionCheck:
    clusterNodes: []
    clusterEndpoints: []
    external:
      - name: local-network-probe
        host: localhost
        port: 9099
        protocol: TCP
        syncLabels:
          connection.nodechecker/local-network-probe: "true"
```

Parameter | Description 
--------- | ----------- 
`external` | Array for rules to check connectivity to a host
`external[].name`    | Name of the checker rule
`external[].host` | Destination host to check connectivity
`external[].port` | Port of the destination host
`external[].protocol` | Protocol of the connection (`TCP` or `UDP` )
`external[].syncLabels` | Labels setted for the `source node` if connection works, if empty `syncLabels: {}` nothing will happen

#### Cluster Endpoints Connectivity

Check connections to Endpoints

```
apiVersion: nodecontroller.io/v1
kind: NodeChecker
metadata:
  name: worker-outbound
spec:
  type: connection
  schedule: "* * * * *"
  sourceNodes:
    matchLabels:
      node-role.kubernetes.io/worker: "true"
  connectionCheck:
    clusterNodes: []
    external: []
    clusterEndpoints:
      - name: access-endpoint-dns
        endpoint:
          name: kube-dns
          namespace: kube-system
        syncLabels:
          connection.nodechecker/endpoint-dns: "true"
```

Parameter | Description 
--------- | ----------- 
`clusterEndpoints` | Array for rules to check endpoints
`clusterEndpoints[].name`    | Name of the checker rule
`clusterEndpoints[].endpoint.name` | Endpoint name
`clusterEndpoints[].endpoint.namespace` | Endpoint namespace
`clusterEndpoints[].syncLabels` | Labels setted for the `source node` if connection works, if empty `syncLabels: {}` nothing will happen


### Feature checker

Validate node features

```
apiVersion: nodecontroller.io/v1
kind: NodeChecker
metadata:
  name: feature-avx
spec:
  type: feature
  schedule: "* * * * *"
  sourceNodes:
    matchLabels:
      node-role.kubernetes.io/worker: "true"
  featureCheck:
    command: |
      if grep -q avx /proc/cpuinfo; then
        exit 0
      else
        exit 1
      fi
    syncLabels:
      feature.nodechecker/avx: "true"
```


Parameter | Description 
--------- | ----------- 
`type`    | Type of checker (`connection` or `feature`)
`schedule`| Scheduler for check execution,  it takes a Cron format string, such as `"* * * * *"` or `"@hourly"`
`sourceNodes.matchLabels` | Labels selector for identify the source node of the connection, set `{}` for select all nodes
`featureCheck.command` | Command to check an existing feature. The command need to end with a `zero (0) exit` status for succeeded or `one (1)` for failed.
`featureCheck.syncLabels` | Labels setted for the `source node` if connection works, if empty `syncLabels: {}` nothing will happen