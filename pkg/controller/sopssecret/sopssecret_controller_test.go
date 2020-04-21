package sopssecret

import (
	"context"
	"testing"

	"k8s.io/client-go/tools/record"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"

	craftypathv1alpha1 "github.com/craftypath/sops-operator/pkg/apis/craftypath/v1alpha1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes/scheme"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

type FakeDecryptor struct{}

func (f *FakeDecryptor) Decrypt(fileName string, encrypted string) ([]byte, error) {
	return []byte("unencrypted"), nil
}

func init() {
	logf.SetLogger(zap.New(zap.UseDevMode(true)))
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
	recorder := record.NewFakeRecorder(1)
	r := newReconcileSopSecret(s, recorder, sopsSecret)

	res, err := r.Reconcile(req)
	require.NoError(t, err)
	assert.False(t, res.Requeue)

	secret := &corev1.Secret{}
	err = r.client.Get(context.Background(), req.NamespacedName, secret)
	require.NoError(t, err)
	assert.Equal(t, secret.Data["test.yaml"], []byte("dW5lbmNyeXB0ZWQ="))
	event := <-recorder.Events
	assert.Equal(t, event, "Normal Created Created secret: test-secret")
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
	recorder := record.NewFakeRecorder(2)
	r := newReconcileSopSecret(s, recorder, sopsSecret)

	res, err := r.Reconcile(req)
	require.NoError(t, err)
	assert.False(t, res.Requeue)
	event := <-recorder.Events
	assert.Equal(t, event, "Normal Created Created secret: test-secret")

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
	event = <-recorder.Events
	assert.Equal(t, event, "Normal Updated Updated secret: test-secret")
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
	recorder := record.NewFakeRecorder(1)
	r := newReconcileSopSecret(s, recorder, secret, sopsSecret)

	_, err := r.Reconcile(req)
	require.NoError(t, err)
	event := <-recorder.Events
	assert.Contains(t, event, "Secret already exists and not owned by sops-operator")
}

func newReconcileSopSecret(s *runtime.Scheme, recorder *record.FakeRecorder, objs ...runtime.Object) *ReconcileSopsSecret {
	cl := fake.NewFakeClientWithScheme(s, objs...)
	return &ReconcileSopsSecret{
		client:    cl,
		scheme:    s,
		recorder:  recorder,
		decryptor: &FakeDecryptor{},
	}
}
