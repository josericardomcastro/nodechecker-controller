# NodeChecker Controller

Native kubernetes controller that allow you to check features and connections of kubernetes nodes.

## How to use

#### Installing via chart

Clone the last version of the project and use helm to install the controller into your cluster.

```
$ git clone https://github.com/josericardomcastro/nodechecker-controller.git
$ cd nodechecker-controller/chart
$ kubectl create namespace nodechecker-controller
$ helm install nodechecker-controller . -n nodechecker-controller
```

#### Creating nodecheckers

The controller works with the custom CRD `kind: NodeChecker` to create nodecheckers.

Nodecheckers are rules to validate node connectivity to another nodes, external host or clusters endpoints. To validate node features, it possible to create rules to check if the node has `AVX`, `GPU` or another feature.

Examples:
- [Connection checker](./manifests/sample-connection.yaml)
- [Feature checker](./manifests/sample-feature.yaml)

See the chart, for the [full documentation](./chart/README.md).

## Development instructions

### Requirements

- Kubernetes 1.19+
- go 1.16


### Commands

#### Generate APIs

After modifying the `pkg/apis/nodecontroller/v1/types.go` file always run the following command to update the generated code for that resource type:

```
:~$ make generate-api 
``` 

#### Generate CRD Manifest 

The CRD manifests can be generated and updated with the following command:

```
:~$ make generate-crd 
``` 