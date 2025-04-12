package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	chinzziiv1 "github.com/Chinzzii/my-k8s-operator/api/v1"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
)

var (
	scheme   = runtime.NewScheme()
	setupLog = ctrl.Log.WithName("setup")
)

func init() {
	utilruntime.Must(chinzziiv1.AddToScheme(scheme))
}

type reconciler struct {
	client.Client                 // Client knows how to perform CRUD operations on Kubernetes objects.
	scheme        *runtime.Scheme // Scheme knows how to convert between different Kubernetes object types.
	kubeClient    *kubernetes.Clientset
}

// Reconcile reads that state of the cluster for a StaticPage object and makes changes based on the state read and what is in the StaticPage.Spec.
// The Controller will requeue the Request to be processed again after the specified duration.
func (r *reconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := log.FromContext(ctx).WithValues("staticpage", req.NamespacedName)
	log.Info("reconciling staticpage")

	// AppsV1 retrieves the AppsV1Client, which is used to interact with the Kubernetes API server for managing deployments.
	deploymentsClient := r.kubeClient.AppsV1().Deployments(req.Namespace)
	// CoreV1 retrieves the CoreV1Client, which is used to interact with the Kubernetes API server for managing core resources like configmaps.
	cmClient := r.kubeClient.CoreV1().ConfigMaps(req.Namespace)

	// get the staticpage name from the request
	// staticpage name is the same as the deployment and configmap name
	staticPageName := "staticpage-" + req.Name

	// get the staticpage object from the request
	var staticPage chinzziiv1.StaticPage
	err := r.Client.Get(ctx, req.NamespacedName, &staticPage)
	if err != nil {
		if k8serrors.IsNotFound(err) { // staticpage not found, we can delete the resources
			err = deploymentsClient.Delete(ctx, staticPageName, metav1.DeleteOptions{})
			if err != nil {
				return ctrl.Result{}, fmt.Errorf("couldn't delete deployment: %s", err)
			}
			err = cmClient.Delete(ctx, staticPageName, metav1.DeleteOptions{})
			if err != nil {
				return ctrl.Result{}, fmt.Errorf("couldn't delete configmap: %s", err)
			}
			return ctrl.Result{}, nil
		}
		return ctrl.Result{}, err
	}

	deployment, err := deploymentsClient.Get(ctx, staticPageName, metav1.GetOptions{})
	// if the deployment is not found, we need to create it
	if err != nil {
		if k8serrors.IsNotFound(err) {
			// create the configmap object
			cmObj := getConfigMapObject(staticPageName, staticPage.Spec.Contents)
			_, err = cmClient.Create(ctx, cmObj, metav1.CreateOptions{})
			if err != nil && !k8serrors.IsAlreadyExists(err) {
				return ctrl.Result{}, fmt.Errorf("couldn't create configmap: %s", err)
			}

			// create the deployment object
			deploymentObj := getDeploymentObject(staticPageName, staticPage.Spec.Image, staticPage.Spec.Replicas)
			_, err := deploymentsClient.Create(ctx, deploymentObj, metav1.CreateOptions{})
			if err != nil {
				return ctrl.Result{}, fmt.Errorf("couldn't create deployment: %s", err)
			}

			log.Info("new staticpage with name " + staticPageName + " created")
			return ctrl.Result{}, nil
		} else {
			return ctrl.Result{}, fmt.Errorf("deployment get error: %s", err)
		}
	}
	// if deployment is found, let's see if we need to update it
	if int(*deployment.Spec.Replicas) != staticPage.Spec.Replicas {
		deployment.Spec.Replicas = int32Ptr(int32(staticPage.Spec.Replicas))
		_, err := deploymentsClient.Update(ctx, deployment, metav1.UpdateOptions{})
		if err != nil {
			return ctrl.Result{}, fmt.Errorf("couldn't update deployment: %s", err)
		}
		log.Info("staticpage with name " + staticPageName + " updated")
		return ctrl.Result{}, nil
	}

	log.Info("staticpage " + staticPageName + " is up-to-date")

	// Result contains the result of a Reconciler invocation.
	return ctrl.Result{}, nil
}

func main() {
	var (
		config *rest.Config // Kubernetes config
		err    error
	)

	kubeconfigFilePath := filepath.Join(homedir.HomeDir(), ".kube", "config")  // kubeconfig file path
	if _, err := os.Stat(kubeconfigFilePath); errors.Is(err, os.ErrNotExist) { // kubeconfig file not found
		config, err = rest.InClusterConfig() // use in-cluster config
		if err != nil {
			panic(err.Error())
		}
	} else { // kubeconfig file found
		config, err = clientcmd.BuildConfigFromFlags("", kubeconfigFilePath) // use kubeconfig file
		if err != nil {
			panic(err.Error())
		}
	}

	// kubernetes clientset
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		panic(err.Error())
	}

	ctrl.SetLogger(zap.New()) // set up logger

	// returns a new Manager for creating Controllers.
	mgr, err := ctrl.NewManager(config, ctrl.Options{
		Scheme: scheme, // scheme for the manager
	})
	if err != nil {
		setupLog.Error(err, "unable to start manager")
		os.Exit(1)
	}

	// returns a new controller builder that will be started by the provided Manager.
	err = ctrl.NewControllerManagedBy(mgr).
		// defines the type of Object being *reconciled*, and configures the ControllerManagedBy to respond to create / delete / update events by *reconciling the object*.
		For(&chinzziiv1.StaticPage{}).

		// builds the Application Controller, which is responsible for reconciling the StaticPage object.
		Complete(&reconciler{
			Client:     mgr.GetClient(),
			scheme:     mgr.GetScheme(),
			kubeClient: clientset,
		})
	if err != nil {
		setupLog.Error(err, "unable to create controller")
		os.Exit(1)
	}

	// start the manager, which starts the controller and watches for events.
	setupLog.Info("starting manager")
	// starts all registered Controllers and blocks until the context is cancelled.
	if err := mgr.Start(ctrl.SetupSignalHandler()); err != nil {
		setupLog.Error(err, "error running manager")
		os.Exit(1)
	}
}

func getDeploymentObject(name string, image string, replicas int) *appsv1.Deployment {
	return &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{Name: name},
		Spec: appsv1.DeploymentSpec{
			Replicas: int32Ptr(int32(replicas)),
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"app": "staticpage",
				},
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"app": "staticpage",
					},
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name:  "staticpage",
							Image: image,
							Ports: []corev1.ContainerPort{
								{
									Name:          "http",
									Protocol:      corev1.ProtocolTCP,
									ContainerPort: 80,
								},
							},
							VolumeMounts: []corev1.VolumeMount{
								{
									Name:      "contents",
									MountPath: "/usr/share/nginx/html",
								},
							},
						},
					},
					Volumes: []corev1.Volume{
						{
							Name: "contents",
							VolumeSource: corev1.VolumeSource{
								ConfigMap: &corev1.ConfigMapVolumeSource{
									LocalObjectReference: corev1.LocalObjectReference{
										Name: name,
									},
								},
							},
						},
					},
				},
			},
		},
	}
}

func getConfigMapObject(name, contents string) *corev1.ConfigMap {
	return &corev1.ConfigMap{
		TypeMeta: metav1.TypeMeta{
			Kind:       "ConfigMap",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
		},
		Data: map[string]string{
			"index.html": contents,
		},
	}
}

func int32Ptr(i int32) *int32 { return &i }
