package service

import (
	"fmt"

	v1alpha1 "github.com/jetstack-experimental/navigator/pkg/apis/navigator/v1alpha1"
	k8sErrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	corelisters "k8s.io/client-go/listers/core/v1"
	"k8s.io/client-go/tools/record"
)

type Interface interface {
	Sync(*v1alpha1.CassandraCluster) error
}

type defaultCassandraClusterServiceControl struct {
	kubeClient    kubernetes.Interface
	serviceLister corelisters.ServiceLister
	recorder      record.EventRecorder
}

var _ Interface = &defaultCassandraClusterServiceControl{}

func NewControl(
	kubeClient kubernetes.Interface,
	serviceLister corelisters.ServiceLister,
	recorder record.EventRecorder,
) Interface {
	return &defaultCassandraClusterServiceControl{
		kubeClient:    kubeClient,
		serviceLister: serviceLister,
		recorder:      recorder,
	}
}

func (e *defaultCassandraClusterServiceControl) Sync(cluster *v1alpha1.CassandraCluster) error {
	svc := ServiceForCluster(cluster)
	client := e.kubeClient.CoreV1().Services(svc.Namespace)
	existingSvc, err := e.serviceLister.Services(svc.Namespace).Get(svc.Name)
	if k8sErrors.IsNotFound(err) {
		_, err = client.Create(svc)
		return err
	}
	if err != nil {
		return err
	}
	if !metav1.IsControlledBy(existingSvc, cluster) {
		return fmt.Errorf("Duplicate service detected: %s", svc)
	}
	_, err = client.Update(svc)
	return err
}
