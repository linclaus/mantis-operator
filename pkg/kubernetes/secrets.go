package kubernetes

import (
	"context"
	"log"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func (f *Framework) GetSecret(namespace, secretName string) (*v1.Secret, error) {
	cm, err := f.KubeClient.CoreV1().Secrets(namespace).Get(context.TODO(), secretName, metav1.GetOptions{})
	if err != nil {
		log.Fatalln("failed to get secret:", err)
	}
	return cm, err
}

func (f *Framework) UpdateSecret(namespace string, cm *v1.Secret) {
	f.KubeClient.CoreV1().Secrets(namespace).Update(context.TODO(), cm, metav1.UpdateOptions{})
}
