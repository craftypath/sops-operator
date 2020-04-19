package sopssecret

import (
	"context"
	"testing"

	craftypathv1alpha1 "github.com/craftypath/sops-operator/pkg/apis/craftypath/v1alpha1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes/scheme"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

type FakeDecryptor struct{}

func (f *FakeDecryptor) Decrypt(fileName string, encrypted string) ([]byte, error) {
	return []byte("unencrypted"), nil
}

var (
	name      = "test-secret"
	namespace = "test-namespace"
	req       = reconcile.Request{
		NamespacedName: types.NamespacedName{
			Name:      name,
			Namespace: namespace,
		},
	}
)

func TestCreate(t *testing.T) {
	sopsSecret := &craftypathv1alpha1.SopsSecret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Spec: craftypathv1alpha1.SopsSecretSpec{
			StringData: map[string]string{"test.yaml": "encrypted"},
		},
	}

	s := scheme.Scheme
	s.AddKnownTypes(craftypathv1alpha1.SchemeGroupVersion, sopsSecret)
	r := newReconcileSopSecret(s, sopsSecret)

	res, err := r.Reconcile(req)
	require.NoError(t, err)
	assert.False(t, res.Requeue)

	secret := &corev1.Secret{}
	err = r.client.Get(context.Background(), req.NamespacedName, secret)
	require.NoError(t, err)
	assert.Equal(t, secret.Data["test.yaml"], []byte("dW5lbmNyeXB0ZWQ="))
}

func TestUpdate(t *testing.T) {
	sopsSecret := &craftypathv1alpha1.SopsSecret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
	}

	s := scheme.Scheme
	s.AddKnownTypes(craftypathv1alpha1.SchemeGroupVersion, sopsSecret)
	r := newReconcileSopSecret(s, sopsSecret)

	res, err := r.Reconcile(req)
	require.NoError(t, err)
	assert.False(t, res.Requeue)

	secret := &corev1.Secret{}
	err = r.client.Get(context.Background(), req.NamespacedName, secret)
	require.NoError(t, err)
	assert.Empty(t, secret.Labels)

	sopsSecret.Labels = map[string]string{
		"foo": "42",
	}
	err = r.client.Update(context.Background(), sopsSecret)
	require.NoError(t, err)

	res, err = r.Reconcile(req)
	require.NoError(t, err)
	assert.False(t, res.Requeue)
	err = r.client.Get(context.Background(), req.NamespacedName, secret)
	require.NoError(t, err)
	assert.Equal(t, secret.Labels["foo"], "42")
}

func TestExistingSecretNotOwnedByUs(t *testing.T) {
	secret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:              name,
			Namespace:         namespace,
			CreationTimestamp: metav1.Now(),
		},
	}

	sopsSecret := &craftypathv1alpha1.SopsSecret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
	}

	s := scheme.Scheme
	s.AddKnownTypes(corev1.SchemeGroupVersion, secret)
	s.AddKnownTypes(craftypathv1alpha1.SchemeGroupVersion, sopsSecret)
	r := newReconcileSopSecret(s, secret, sopsSecret)

	_, err := r.Reconcile(req)
	require.Error(t, err)
	assert.Equal(t, "secret already exists and not owned by sops-operator", err.Error())
}

func newReconcileSopSecret(s *runtime.Scheme, objs ...runtime.Object) *ReconcileSopsSecret {
	cl := fake.NewFakeClientWithScheme(s, objs...)
	return &ReconcileSopsSecret{
		client:    cl,
		scheme:    s,
		decryptor: &FakeDecryptor{},
	}
}
