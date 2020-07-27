package exoscale

import (
	"context"
	"reflect"
	"testing"

	"github.com/stretchr/testify/require"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/fake"
)

const testServiceID = "009e9cb6-147e-4a33-ab93-7651f246cb5c"

func newFakeService() *v1.Service {
	return &v1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:        "service-name",
			Namespace:   metav1.NamespaceDefault,
			Annotations: map[string]string{},
		},
	}
}

// newFakeService takes a service and returns a copy of the service called "current"
// and a pointer to the service called "modified".
// TestNewServicePatcher test firstly when the Service was not modified
// if the current service and the modified service are equal.
// And secondly when the Service was modified
// if the current service and the modified service are unequal.
func TestNewServicePatcher(t *testing.T) {
	ctx := context.Background()
	clientset := fake.NewSimpleClientset()

	service := newFakeService()
	patcher := newServicePatcher(ctx, clientset, service)

	if !reflect.DeepEqual(patcher.current, patcher.modified) {
		t.Errorf("patcher.current and patcher.modified must be equal")
	}

	service.ObjectMeta.Annotations[annotationLoadBalancerID] = testServiceID

	if reflect.DeepEqual(patcher.current, patcher.modified) {
		t.Errorf("patcher.current and patcher.modified must be unequal")
	}
}

func TestPatch(t *testing.T) {
	ctx := context.Background()
	clientset := fake.NewSimpleClientset()

	service := newFakeService()
	serviceAdded, err := clientset.CoreV1().Services(metav1.NamespaceDefault).Create(ctx, service, metav1.CreateOptions{})
	require.NoError(t, err)

	svcID := serviceAdded.ObjectMeta.Annotations[annotationLoadBalancerID]
	require.Empty(t, svcID)

	patcher := newServicePatcher(ctx, clientset, service)
	service.ObjectMeta.Annotations[annotationLoadBalancerID] = testServiceID
	err = patcher.Patch()
	require.NoError(t, err)

	serviceFinal, err := clientset.CoreV1().Services(metav1.NamespaceDefault).Get(ctx, service.Name, metav1.GetOptions{})
	require.NoError(t, err)

	svcID = serviceFinal.ObjectMeta.Annotations[annotationLoadBalancerID]
	require.Equal(t, svcID, testServiceID)
}
