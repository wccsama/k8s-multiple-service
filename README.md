# k8s-multiple-service
集群service 同步，从一个集群import到另外的集群，实现思路参考k8s import service。需要在部署controller的yaml文件中指定待导出集群的kubeconfig
与 导出至某个集群的kubeconfig。

build:
make build

原理说明：
修改endpoints

实践例子：
TODO