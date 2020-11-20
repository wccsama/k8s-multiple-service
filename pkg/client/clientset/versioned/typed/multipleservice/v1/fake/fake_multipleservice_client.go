// Code generated by client-gen. DO NOT EDIT.

package fake

import (
	v1 "github.com/k8s-multiple-service/pkg/client/clientset/versioned/typed/multipleservice/v1"
	rest "k8s.io/client-go/rest"
	testing "k8s.io/client-go/testing"
)

type FakeWccV1 struct {
	*testing.Fake
}

func (c *FakeWccV1) ServiceExports(namespace string) v1.ServiceExportInterface {
	return &FakeServiceExports{c, namespace}
}

// RESTClient returns a RESTClient that is used to communicate
// with API server by this client implementation.
func (c *FakeWccV1) RESTClient() rest.Interface {
	var ret *rest.RESTClient
	return ret
}