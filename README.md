# k8s-multiple-service
集群service 同步，从一个集群import到另外的集群，实现思路参考k8s import service。需要在部署controller的yaml文件中指定待导出集群的kubeconfig
与 导出至某个集群的kubeconfig。

build:
make build

原理说明：
修改endpoints，让原本网络互通的两个集群实现service的共享。将集群A下的某个namespace的所有service 同步到集群B的若干指定namespace，这些配置在可以在CRD中修改，
controller 会监控CRD资源的创建、删除、更新事件，创建，更新，删除相应service。

实践例子：
TODO