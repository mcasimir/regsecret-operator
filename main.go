package main

import (
	"encoding/json"
	"flag"

	log "github.com/sirupsen/logrus"

	"os"
	"os/signal"
	"path/filepath"
	"sync"
	"syscall"

	"github.com/mcasimir/regsecret-operator/pkg/controller"
	"github.com/mcasimir/regsecret-operator/pkg/logging"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

type appConfig struct {
	Logger  logging.Config       `json:"logger"`
	Secrets []*controller.Config `json:"secrets"`
}

func main() {
	config := &appConfig{}
	err := json.Unmarshal([]byte(os.Getenv("REGSECRET_OPERATOR_CONFIG")), config)
	if err != nil {
		panic(err.Error())
	}

	logging.Setup(config.Logger)

	sigs := make(chan os.Signal, 1) // Create channel to receive OS signals
	stop := make(chan struct{})     // Create channel to receive stop signal

	signal.Notify(sigs, os.Interrupt, syscall.SIGTERM, syscall.SIGINT) // Register the sigs channel to receieve SIGTERM

	wg := &sync.WaitGroup{} // Goroutines can add themselves to this to be waited on so that they finish

	runOutsideCluster := flag.Bool("run-outside-cluster", false, "Set this flag when running outside of the cluster.")
	flag.Parse()
	// Create clientset for interacting with the kubernetes cluster
	clientset, err := newClientSet(*runOutsideCluster)

	if err != nil {
		panic(err.Error())
	}

	for _, secretConfig := range config.Secrets {
		err = secretConfig.Validate()
		if err != nil {
			panic(err.Error())
		}

		go (func(c controller.Config) {
			log.Debugf("Goroutine started for secret %s", c.SecretName)
			controller.New(clientset, &c).Run(stop, wg)
		})(*secretConfig)
	}

	log.Infof("Regsecret Operator started")

	<-sigs // Wait for signals (this hangs until a signal arrives)
	log.Infof("Shutting down...")

	close(stop) // Tell goroutines to stop themselves
	wg.Wait()   // Wait for all to be stopped
}

func newClientSet(runOutsideCluster bool) (*kubernetes.Clientset, error) {
	kubeConfigLocation := ""

	if runOutsideCluster == true {
		homeDir := os.Getenv("HOME")
		kubeConfigLocation = filepath.Join(homeDir, ".kube", "config")
	}

	// use the current context in kubeconfig
	config, err := clientcmd.BuildConfigFromFlags("", kubeConfigLocation)

	if err != nil {
		return nil, err
	}

	return kubernetes.NewForConfig(config)
}
