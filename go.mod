module github.com/k8s-multiple-service

go 1.14

require (
	k8s.io/apiextensions-apiserver v0.18.4
	k8s.io/apimachinery v0.18.4
	k8s.io/code-generator v0.18.4
	k8s.io/klog v1.0.0
)

replace k8s.io/code-generator v0.18.4 => k8s.io/code-generator v0.0.0-20190612205613-18da4a14b22b
