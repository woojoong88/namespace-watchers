package kubernetes

import (
	"context"
	"fmt"
	"github.com/sirupsen/logrus"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	v1 "k8s.io/client-go/informers/core/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/cache"
	"slices"
	"time"
)

// WatchNamespace watches for new namespaces and creates an echo pod in the namespace
func WatchNamespace(ctx context.Context, excludedNamespaces []string, startTime time.Time) error {
	// Get the in-cluster config - this is used to create the clientset
	config, err := rest.InClusterConfig()
	if err != nil {
		return err
	}

	// Create the clientset to interact with the Kubernetes API
	// Note that this requires the right ServiceAccount and (cluster) Roles
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return err
	}

	// get Informer to watch for new namespaces
	nsInformer := v1.NewNamespaceInformer(
		clientset, 0, cache.Indexers{})

	// Add event handler to the informer
	// For this use-case, we just only need to add `Add` event handler
	nsInformer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			ns := obj.(*corev1.Namespace)

			// skip for excluded namespaces
			if excludedNamespaces != nil {
				if slices.Contains(excludedNamespaces, ns.Name) {
					return
				}
			}

			// skip the namespaces if they are created before this service is started
			if !ns.CreationTimestamp.After(startTime) {
				return
			}

			// create echo-pod - can be running in go-routine for parallel processing
			err = createEchoPod(ctx, ns.Name, clientset)
			if err != nil {
				logrus.Error(err, " failed to create echo pod")
				return
			}
		},
	})

	// prepare to stop the informer
	stopCh := make(chan struct{})
	defer close(stopCh)

	// run the informer
	go nsInformer.Run(stopCh)

	// wait for the cache to sync
	if !cache.WaitForCacheSync(stopCh, nsInformer.HasSynced) {
		err = fmt.Errorf("Timeout waiting for caches to sync")
		logrus.Error(err)
		return err
	}

	// wait for the informer to stop
	<-stopCh
	err = fmt.Errorf("informer stopped")

	return err
}

// createEchoPod creates a pod in the given namespace that echoes the namespace name
func createEchoPod(ctx context.Context, namespace string, clientset *kubernetes.Clientset) error {
	// message to echo
	echoMsg := fmt.Sprintf("echo namespace: %s", namespace)

	// create a pod that echoes the namespace name
	pod := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "echo-pod",
			Namespace: namespace,
		},
		Spec: corev1.PodSpec{
			RestartPolicy: corev1.RestartPolicyNever,
			Containers: []corev1.Container{
				{
					Name:  "echo",
					Image: "busybox:stable",
					Args:  []string{"sh", "-c", echoMsg},
				},
			},
		},
	}

	logrus.Info(fmt.Sprintf("Creating pod %s in namespace %s: spec %+v", pod.Name, pod.Namespace, pod.Spec))

	// check if the pod already exists before creating it
	_, err := clientset.CoreV1().Pods(pod.Namespace).Get(ctx, pod.Name, metav1.GetOptions{})
	if err == nil {
		logrus.Info("echo-pod exists, skipping creation")
		return nil
	}

	// if error is captured but it is not NotFound, return the error
	if !errors.IsNotFound(err) {
		logrus.Error(err, " failed to get pod to check if it already exists")
		return err
	}

	// create the pod
	if _, err := clientset.CoreV1().Pods(pod.Namespace).Create(ctx, pod, metav1.CreateOptions{}); err != nil {
		logrus.Error(err, " failed to create pod")
		return err
	}

	return nil
}
