package operator

import (
	"context"
	"testing"

	appsv1alpha1 "github.com/3scale/3scale-operator/pkg/apis/apps/v1alpha1"
	appsv1 "github.com/openshift/api/apps/v1"
	imagev1 "github.com/openshift/api/image/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes/scheme"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
)

func TestRedisBackendDCReconcilerCreate(t *testing.T) {
	var (
		appLabel       = "someLabel"
		name           = "example-apimanager"
		namespace      = "operator-unittest"
		trueValue      = true
		wildcardDomain = "test.3scale.net"
		tenantName     = "someTenant"
		log            = logf.Log.WithName("operator_test")
	)
	apimanager := &appsv1alpha1.APIManager{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Spec: appsv1alpha1.APIManagerSpec{
			APIManagerCommonSpec: appsv1alpha1.APIManagerCommonSpec{
				AppLabel:                     &appLabel,
				ImageStreamTagImportInsecure: &trueValue,
				ResourceRequirementsEnabled:  &trueValue,
				WildcardDomain:               wildcardDomain,
				TenantName:                   &tenantName,
			},
		},
	}
	s := scheme.Scheme
	s.AddKnownTypes(appsv1alpha1.SchemeGroupVersion, apimanager)
	err := imagev1.AddToScheme(s)
	if err != nil {
		t.Fatal(err)
	}
	err = appsv1.AddToScheme(s)
	if err != nil {
		t.Fatal(err)
	}

	// Objects to track in the fake client.
	objs := []runtime.Object{}

	// Create a fake client to mock API calls.
	cl := fake.NewFakeClient(objs...)
	clientAPIReader := fake.NewFakeClient(objs...)

	baseReconciler := NewBaseReconciler(cl, clientAPIReader, s, log)
	baseLogicReconciler := NewBaseLogicReconciler(baseReconciler)
	baseAPIManagerLogicReconciler := NewBaseAPIManagerLogicReconciler(baseLogicReconciler, apimanager)
	reconciler := NewRedisReconciler(baseAPIManagerLogicReconciler)
	_, err = reconciler.Reconcile()
	if err != nil {
		t.Fatal(err)
	}

	cases := []struct {
		testName string
		objName  string
		obj      runtime.Object
	}{
		{"backendRedisDC", "backend-redis", &appsv1.DeploymentConfig{}},
		{"backendRedisService", "backend-redis", &v1.Service{}},
		{"backendRedisCM", "redis-config", &v1.ConfigMap{}},
		{"backendRedisPVC", "backend-redis-storage", &v1.PersistentVolumeClaim{}},
		{"backendRedisIS", "backend-redis", &imagev1.ImageStream{}},
		{"systemRedisDC", "system-redis", &appsv1.DeploymentConfig{}},
		{"systemRedisPVC", "system-redis-storage", &v1.PersistentVolumeClaim{}},
		{"systemRedisIS", "system-redis", &imagev1.ImageStream{}},
		{"systemRedisService", "system-redis", &v1.Service{}},
	}

	for _, tc := range cases {
		t.Run(tc.testName, func(subT *testing.T) {
			obj := tc.obj
			namespacedName := types.NamespacedName{
				Name:      tc.objName,
				Namespace: namespace,
			}
			err = cl.Get(context.TODO(), namespacedName, obj)
			// object must exist, that is all required to be tested
			if err != nil {
				subT.Errorf("error fetching object %s: %v", tc.objName, err)
			}
		})
	}
}
