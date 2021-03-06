// Code generated by lister-gen. DO NOT EDIT.

package v1

import (
	v1 "github.com/k8s-multiple-service/pkg/api/multipleservice/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/tools/cache"
)

// ServiceExportLister helps list ServiceExports.
type ServiceExportLister interface {
	// List lists all ServiceExports in the indexer.
	List(selector labels.Selector) (ret []*v1.ServiceExport, err error)
	// ServiceExports returns an object that can list and get ServiceExports.
	ServiceExports(namespace string) ServiceExportNamespaceLister
	ServiceExportListerExpansion
}

// serviceExportLister implements the ServiceExportLister interface.
type serviceExportLister struct {
	indexer cache.Indexer
}

// NewServiceExportLister returns a new ServiceExportLister.
func NewServiceExportLister(indexer cache.Indexer) ServiceExportLister {
	return &serviceExportLister{indexer: indexer}
}

// List lists all ServiceExports in the indexer.
func (s *serviceExportLister) List(selector labels.Selector) (ret []*v1.ServiceExport, err error) {
	err = cache.ListAll(s.indexer, selector, func(m interface{}) {
		ret = append(ret, m.(*v1.ServiceExport))
	})
	return ret, err
}

// ServiceExports returns an object that can list and get ServiceExports.
func (s *serviceExportLister) ServiceExports(namespace string) ServiceExportNamespaceLister {
	return serviceExportNamespaceLister{indexer: s.indexer, namespace: namespace}
}

// ServiceExportNamespaceLister helps list and get ServiceExports.
type ServiceExportNamespaceLister interface {
	// List lists all ServiceExports in the indexer for a given namespace.
	List(selector labels.Selector) (ret []*v1.ServiceExport, err error)
	// Get retrieves the ServiceExport from the indexer for a given namespace and name.
	Get(name string) (*v1.ServiceExport, error)
	ServiceExportNamespaceListerExpansion
}

// serviceExportNamespaceLister implements the ServiceExportNamespaceLister
// interface.
type serviceExportNamespaceLister struct {
	indexer   cache.Indexer
	namespace string
}

// List lists all ServiceExports in the indexer for a given namespace.
func (s serviceExportNamespaceLister) List(selector labels.Selector) (ret []*v1.ServiceExport, err error) {
	err = cache.ListAllByNamespace(s.indexer, s.namespace, selector, func(m interface{}) {
		ret = append(ret, m.(*v1.ServiceExport))
	})
	return ret, err
}

// Get retrieves the ServiceExport from the indexer for a given namespace and name.
func (s serviceExportNamespaceLister) Get(name string) (*v1.ServiceExport, error) {
	obj, exists, err := s.indexer.GetByKey(s.namespace + "/" + name)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, errors.NewNotFound(v1.Resource("serviceexport"), name)
	}
	return obj.(*v1.ServiceExport), nil
}
