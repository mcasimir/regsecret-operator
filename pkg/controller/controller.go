package controller

import (
	"encoding/base64"
	"encoding/json"
	"fmt"

	log "github.com/sirupsen/logrus"

	"sync"
	"time"

	"k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
)

type NamespaceController struct {
	informer cache.SharedIndexInformer
	client   *kubernetes.Clientset
	config   *Config
}

func New(client *kubernetes.Clientset, config *Config) *NamespaceController {
	controller := &NamespaceController{
		config: config,
	}

	informer := cache.NewSharedIndexInformer(
		&cache.ListWatch{
			ListFunc: func(options metav1.ListOptions) (runtime.Object, error) {
				options.LabelSelector = config.NamespaceSelector
				return client.Core().Namespaces().List(options)
			},
			WatchFunc: func(options metav1.ListOptions) (watch.Interface, error) {
				options.LabelSelector = config.NamespaceSelector
				return client.Core().Namespaces().Watch(options)
			},
		},
		&v1.Namespace{},
		3*time.Minute,
		cache.Indexers{cache.NamespaceIndex: cache.MetaNamespaceIndexFunc},
	)

	informer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc:    controller.createSecretOnAdd,
		UpdateFunc: controller.createSecretOnUpdate,
	})

	controller.client = client
	controller.informer = informer

	return controller
}

func (c *NamespaceController) Run(stopCh <-chan struct{}, wg *sync.WaitGroup) {
	defer wg.Done()
	wg.Add(1)
	go c.informer.Run(stopCh)
	<-stopCh
}

func (c *NamespaceController) createSecretOnAdd(obj interface{}) {
	namespaceObj := obj.(*v1.Namespace)
	c.createSecret(namespaceObj.Name)
}

func (c *NamespaceController) createSecretOnUpdate(oldObj, obj interface{}) {
	namespaceObj := obj.(*v1.Namespace)
	c.createSecret(namespaceObj.Name)
}

func (c *NamespaceController) createSecret(namespaceName string) {

	secret, err := c.newSecretFromCredentials(namespaceName)
	if err != nil {
		log.Error(err.Error)
	}

	_, err = c.client.CoreV1().Secrets(namespaceName).Create(secret)

	if err != nil {
		statusError, isStatusError := err.(*apierrors.StatusError)

		if isStatusError && statusError.Status().Reason == metav1.StatusReasonAlreadyExists {
			log.Debugf("Skip secret %s for namespace %s: %s", secret.Name, namespaceName, err.Error())
			return
		}

		log.Warnf("Failed to create Secret %s for Namespace %s: %s", secret.Name, namespaceName, err.Error())
	} else {
		log.Infof("Created Secret %s for Namespace %s", secret.Name, namespaceName)
	}
}

type dockercfgjson struct {
	Auths map[string]dockercfgjsonAuth `json:"auths"`
}

type dockercfgjsonAuth struct {
	Username string `json:"username"`
	Password string `json:"password"`
	Email    string `json:"email"`
	Auth     string `json:"auth"`
}

func (c *NamespaceController) newSecretFromCredentials(namespaceName string) (*v1.Secret, error) {
	dockercfgjsonStruct := dockercfgjson{Auths: map[string]dockercfgjsonAuth{}}

	for uri, auth := range c.config.Credentials {
		base64Auth := base64.StdEncoding.EncodeToString([]byte(auth.Username + ":" + auth.Password))
		dockercfgjsonStruct.Auths[uri] = dockercfgjsonAuth{
			Username: auth.Username,
			Password: auth.Password,
			Email:    auth.Email,
			Auth:     base64Auth,
		}
	}

	dockercfgjsonBytes, err := json.Marshal(&dockercfgjsonStruct)
	if err != nil {
		return nil, fmt.Errorf("JSON marshal error creating Secret %s for Namespace %s: %s", c.config.SecretName, namespaceName, err.Error())
	}

	secret := &v1.Secret{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Secret",
			APIVersion: "core/v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Namespace: namespaceName,
			Name:      c.config.SecretName,
		},
		Data: map[string][]byte{
			".dockerconfigjson": dockercfgjsonBytes,
		},
		Type: "kubernetes.io/dockerconfigjson",
	}

	return secret, nil
}
