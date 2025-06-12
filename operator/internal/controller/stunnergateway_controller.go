package controller

import (
	"context"
	"fmt"
	"time"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	cachev1alpha1 "github.com/Razano26/Nebula/operator/api/v1alpha1"
	"k8s.io/utils/pointer"
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

	// Create or update the Stunner Deployment
	deployment := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      fmt.Sprintf("stunner-%s", stunnerIngress.Name),
			Namespace: namespace,
		},
		Spec: appsv1.DeploymentSpec{
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"app": fmt.Sprintf("stunner-%s", stunnerIngress.Name),
				},
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"app": fmt.Sprintf("stunner-%s", stunnerIngress.Name),
					},
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name:  "stunner",
							Image: "l7mp/stunner:latest", // Use appropriate Stunner image
							SecurityContext: &corev1.SecurityContext{
								AllowPrivilegeEscalation: pointer.Bool(false),
								Capabilities: &corev1.Capabilities{
									Drop: []corev1.Capability{"ALL"},
								},
								RunAsNonRoot: pointer.Bool(true),
								SeccompProfile: &corev1.SeccompProfile{
									Type: corev1.SeccompProfileTypeRuntimeDefault,
								},
							},
							Ports: []corev1.ContainerPort{
								{
									ContainerPort: stunnerIngress.Spec.Port,
									Protocol:      getProtocol(stunnerIngress.Spec.Protocol),
								},
							},
							Env: []corev1.EnvVar{
								{
									Name:  "STUNNER_PROTOCOL",
									Value: stunnerIngress.Spec.Protocol,
								},
								{
									Name:  "STUNNER_PORT",
									Value: fmt.Sprintf("%d", stunnerIngress.Spec.Port),
								},
							},
						},
					},
				},
			},
		},
	}

	// Create or update the deployment
	if err := r.createOrUpdate(ctx, deployment); err != nil {
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
