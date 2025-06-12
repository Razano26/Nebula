package controller

import (
	"context"
	"fmt"
	"time"

	"os"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/client-go/discovery"
	"k8s.io/client-go/discovery/cached/memory"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/restmapper"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	cachev1alpha1 "github.com/Razano26/Nebula/operator/api/v1alpha1"
	"helm.sh/helm/v3/pkg/action"
	"helm.sh/helm/v3/pkg/chart/loader"
	"helm.sh/helm/v3/pkg/cli"
	"helm.sh/helm/v3/pkg/getter"
	"helm.sh/helm/v3/pkg/repo"
)

// StunnerIngressReconciler reconciles a StunnerIngress object
type StunnerIngressReconciler struct {
	client.Client
	Scheme   *runtime.Scheme
	Recorder record.EventRecorder
}

// SetupWithManager sets up the controller with the Manager.
func (r *StunnerIngressReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&cachev1alpha1.StunnerIngress{}).
		Complete(r)
}

func (r *StunnerIngressReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := log.FromContext(ctx)

	// Fetch the StunnerIngress instance
	stunnerIngress := &cachev1alpha1.StunnerIngress{}
	if err := r.Get(ctx, req.NamespacedName, stunnerIngress); err != nil {
		if errors.IsNotFound(err) {
			return ctrl.Result{}, nil
		}
		return ctrl.Result{}, err
	}

	// Define the target namespace for resources
	namespace := stunnerIngress.Namespace

	// Check if target service exists
	targetService := &corev1.Pod{}
	targetNamespace := stunnerIngress.Spec.Target.Namespace
	if targetNamespace == "" {
		targetNamespace = namespace
	}

	err := r.Get(ctx, types.NamespacedName{
		Name:      stunnerIngress.Spec.Target.Name,
		Namespace: targetNamespace,
	}, targetService)
	if err != nil {
		if errors.IsNotFound(err) {
			logger.Error(err, "Target service not found")
			// Update status condition
			meta.SetStatusCondition(&stunnerIngress.Status.Conditions, metav1.Condition{
				Type:    "Available",
				Status:  metav1.ConditionFalse,
				Reason:  "TargetServiceNotFound",
				Message: fmt.Sprintf("Target service %s not found in namespace %s", stunnerIngress.Spec.Target.Name, targetNamespace),
			})
			r.Status().Update(ctx, stunnerIngress)
			return ctrl.Result{RequeueAfter: time.Second * 30}, nil
		}
		return ctrl.Result{}, err
	}

	err = r.installHelmChart(ctx, namespace, "https://l7mp.io/stunner/stunner", "stunner", map[string]interface{}{})
	if err != nil {
		return ctrl.Result{}, err
	}

	// Create or update the Service
	service := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      fmt.Sprintf("stunner-svc-%s", stunnerIngress.Name),
			Namespace: namespace,
		},
		Spec: corev1.ServiceSpec{
			Type: corev1.ServiceTypeLoadBalancer,
			Ports: []corev1.ServicePort{
				{
					Port:       getExternalPort(stunnerIngress),
					TargetPort: intstr.FromInt(int(stunnerIngress.Spec.Port)),
					Protocol:   getProtocol(stunnerIngress.Spec.Protocol),
				},
			},
			Selector: map[string]string{
				"app": fmt.Sprintf("stunner-%s", stunnerIngress.Name),
			},
		},
	}

	// Create or update the service
	if err := r.createOrUpdate(ctx, service); err != nil {
		return ctrl.Result{}, err
	}

	// Update status with external addresses
	if len(service.Status.LoadBalancer.Ingress) > 0 {
		var addresses []string
		for _, ingress := range service.Status.LoadBalancer.Ingress {
			if ingress.IP != "" {
				addresses = append(addresses, ingress.IP)
			}
			if ingress.Hostname != "" {
				addresses = append(addresses, ingress.Hostname)
			}
		}
		stunnerIngress.Status.ExternalAddresses = addresses
	}

	// Update status condition
	meta.SetStatusCondition(&stunnerIngress.Status.Conditions, metav1.Condition{
		Type:    "Available",
		Status:  metav1.ConditionTrue,
		Reason:  "ResourcesCreated",
		Message: "All resources have been created successfully",
	})

	if err := r.Status().Update(ctx, stunnerIngress); err != nil {
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}

// Helper functions

func getProtocol(protocol string) corev1.Protocol {
	return corev1.Protocol(protocol)
}

func getExternalPort(stunnerIngress *cachev1alpha1.StunnerIngress) int32 {
	if stunnerIngress.Spec.ExternalPort != nil {
		return *stunnerIngress.Spec.ExternalPort
	}
	return stunnerIngress.Spec.Port
}

func (r *StunnerIngressReconciler) createOrUpdate(ctx context.Context, obj client.Object) error {
	if err := r.Create(ctx, obj); err != nil {
		if errors.IsAlreadyExists(err) {
			return r.Update(ctx, obj)
		}
		return err
	}
	return nil
}

func (r *StunnerIngressReconciler) installHelmChart(ctx context.Context, namespace string, chartPath string, releaseName string, values map[string]interface{}) error {
	logger := log.FromContext(ctx)
	logger.Info("Installing Helm chart", "namespace", namespace, "chartPath", chartPath, "releaseName", releaseName)

	// Create action configuration
	actionConfig := new(action.Configuration)

	// Get REST config from controller-runtime
	config, err := ctrl.GetConfig()
	if err != nil {
		return fmt.Errorf("failed to get Kubernetes config: %w", err)
	}

	// Initialize action configuration with Kubernetes client
	err = actionConfig.Init(
		&helmRESTClientGetter{config: config, namespace: namespace},
		namespace,
		os.Getenv("HELM_DRIVER"),
		logger.Info,
	)
	if err != nil {
		return fmt.Errorf("failed to initialize Helm action configuration: %w", err)
	}

	// Create install or upgrade action
	helmClient := action.NewUpgrade(actionConfig)
	helmClient.Install = true // Upgrade if exists, install if doesn't exist
	helmClient.Namespace = namespace
	helmClient.Wait = true
	helmClient.Timeout = 300 * time.Second

	// Set up repository
	settings := cli.New()

	// Create repository entry
	chartRepo := &repo.Entry{
		Name: "stunner",
		URL:  "https://l7mp.io/stunner",
	}

	// Create chart repository
	chartRepository, err := repo.NewChartRepository(chartRepo, getter.All(settings))
	if err != nil {
		return fmt.Errorf("failed to create chart repository: %w", err)
	}

	// Download index file
	indexFile, err := chartRepository.DownloadIndexFile()
	if err != nil {
		return fmt.Errorf("failed to download index file: %w", err)
	}

	// Load chart
	chartRequested, err := loader.Load(indexFile)
	if err != nil {
		return fmt.Errorf("failed to load Helm chart: %w", err)
	}

	// Run upgrade/install
	_, err = helmClient.Run(releaseName, chartRequested, values)
	if err != nil {
		return fmt.Errorf("failed to install/upgrade Helm chart %s: %w", releaseName, err)
	}

	logger.Info("Helm chart installed successfully", "namespace", namespace, "releaseName", releaseName)
	return nil
}

// helmRESTClientGetter implements the genericclioptions.RESTClientGetter interface
// which is required by Helm to interact with the Kubernetes cluster
type helmRESTClientGetter struct {
	config    *rest.Config
	namespace string
}

// ToRESTConfig returns the REST client configuration
func (c *helmRESTClientGetter) ToRESTConfig() (*rest.Config, error) {
	return c.config, nil
}

// ToDiscoveryClient returns a discovery client
func (c *helmRESTClientGetter) ToDiscoveryClient() (discovery.CachedDiscoveryInterface, error) {
	config, err := c.ToRESTConfig()
	if err != nil {
		return nil, err
	}

	// Create a discovery client
	discoveryClient, err := discovery.NewDiscoveryClientForConfig(config)
	if err != nil {
		return nil, err
	}

	// Create a cached discovery client
	return memory.NewMemCacheClient(discoveryClient), nil
}

// ToRESTMapper returns a REST mapper
func (c *helmRESTClientGetter) ToRESTMapper() (meta.RESTMapper, error) {
	discoveryClient, err := c.ToDiscoveryClient()
	if err != nil {
		return nil, err
	}

	mapper := restmapper.NewDeferredDiscoveryRESTMapper(discoveryClient)
	return mapper, nil
}

// ToRawKubeConfigLoader returns a ClientConfig
func (c *helmRESTClientGetter) ToRawKubeConfigLoader() clientcmd.ClientConfig {
	loadingRules := clientcmd.NewDefaultClientConfigLoadingRules()
	configOverrides := &clientcmd.ConfigOverrides{}
	return clientcmd.NewNonInteractiveDeferredLoadingClientConfig(loadingRules, configOverrides)
}
