package prometheus

import (
	"context"
	"fmt"
	"log"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func (f *Framework) GetConfigMap(namespace, configName string) (*v1.ConfigMap, error) {
	cm, err := f.KubeClient.CoreV1().ConfigMaps(namespace).Get(context.TODO(), configName, metav1.GetOptions{})
	if err != nil {
		log.Fatalln("failed to get config map:", err)
	}
	fmt.Printf("name %s\n", cm.GetName())
	fmt.Printf("data %s\n", cm.Data)
	return cm, err
}

func (f *Framework) UpdateConfigMap(namespace string, cm *v1.ConfigMap) {
	f.KubeClient.CoreV1().ConfigMaps(namespace).Update(context.TODO(), cm, metav1.UpdateOptions{})
}

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
