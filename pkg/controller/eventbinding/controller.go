/*
Copyright 2017 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package eventbinding

import (
	"context"
	"fmt"
	"os"
	"time"

	samplev1alpha1 "github.com/vincent-pli/tektonpipeline-listener/pkg/apis/samplecontroller/v1alpha1"
	"github.com/vincent-pli/tektonpipeline-listener/pkg/controller/resources"
	clientset "github.com/vincent-pli/tektonpipeline-listener/pkg/generated/clientset/versioned"
	samplescheme "github.com/vincent-pli/tektonpipeline-listener/pkg/generated/clientset/versioned/scheme"
	informers "github.com/vincent-pli/tektonpipeline-listener/pkg/generated/informers/externalversions/samplecontroller/v1alpha1"
	listers "github.com/vincent-pli/tektonpipeline-listener/pkg/generated/listers/samplecontroller/v1alpha1"
	ksvclister "github.com/knative/serving/pkg/client/listers/serving/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/labels"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/apimachinery/pkg/util/wait"
	appsinformers "k8s.io/client-go/informers/apps/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/scheme"
	typedcorev1 "k8s.io/client-go/kubernetes/typed/core/v1"
	appslisters "k8s.io/client-go/listers/apps/v1"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/record"
	"k8s.io/client-go/util/workqueue"
	"k8s.io/klog"

	servingv1alpha1 "github.com/knative/serving/pkg/apis/serving/v1alpha1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	ksvcClientset "github.com/knative/serving/pkg/client/clientset/versioned"
)

const controllerAgentName = "tektonpipeline-listener"

const (
	// SuccessSynced is used as part of the Event 'reason' when a ListenerTemplate is synced
	SuccessSynced = "Synced"
	// FailedDeleted is used as part of the Event 'reason' when a ListenerTemplate is delete failed
	FailedDeleted = "FailedDeleted"
	// ErrResourceExists is used as part of the Event 'reason' when a Foo fails
	// to sync due to a Deployment of the same name already existing.
	ErrResourceExists = "ErrResourceExists"

	// MessageResourceExists is the message used for Events when a resource
	// fails to sync due to a Deployment already existing
	MessageResourceExists = "Resource %q already exists and is not managed by Foo"
	// MessageResourceSynced is the message used for an Event fired when a Foo
	// is synced successfully
	MessageResourceSynced = "EventBinding synced successfully"

	// MessageResourceDeleteFailed is the message used for an Event fired when a ListenerTemplate
	// is failed to delete
	MessageResourceDeleteFailed = "EventBinding delete failed, since refrence from EventBinding"

	lsImageEnvVar = "LISTENER_IMAGE"
)

// Controller is the controller implementation for Foo resources
type Controller struct {
	// kubeclientset is a standard kubernetes clientset
	kubeclientset kubernetes.Interface
	// sampleclientset is a clientset for our own API group
	sampleclientset clientset.Interface
	ksvcclientset ksvcClientset.Interface,

	deploymentsLister   appslisters.DeploymentLister
	deploymentsSynced   cache.InformerSynced
	eventBindingsLister listers.EventBindingLister
	eventBindingsSynced cache.InformerSynced

	ksvcLister ksvclister.EventBindingLister
	// workqueue is a rate limited work queue. This is used to queue work to be
	// processed instead of performing it as soon as a change happens. This
	// means we can ensure we only process a fixed amount of resources at a
	// time, and makes it easy to ensure we are never processing the same item
	// simultaneously in two different workers.
	workqueue workqueue.RateLimitingInterface
	// recorder is an event recorder for recording Event resources to the
	// Kubernetes API.
	recorder record.EventRecorder
	client   client.Client
}

// NewController returns a new sample controller
func NewController(
	kubeclientset kubernetes.Interface,
	sampleclientset clientset.Interface,
	ksvcclientset ksvcClientset.Interface,
	deploymentInformer appsinformers.DeploymentInformer,
	eventBindingInformer informers.EventBindingInformer) *Controller {

	// Create event broadcaster
	// Add tektonpipeline-listener types to the default Kubernetes Scheme so Events can be
	// logged for tektonpipeline-listener types.
	utilruntime.Must(samplescheme.AddToScheme(scheme.Scheme))
	klog.V(4).Info("Creating event broadcaster")
	eventBroadcaster := record.NewBroadcaster()
	eventBroadcaster.StartLogging(klog.Infof)
	eventBroadcaster.StartRecordingToSink(&typedcorev1.EventSinkImpl{Interface: kubeclientset.CoreV1().Events("")})
	recorder := eventBroadcaster.NewRecorder(scheme.Scheme, corev1.EventSource{Component: controllerAgentName})

	listenerImage, defined := os.LookupEnv(lsImageEnvVar)
	if !defined {
		return fmt.Errorf("required environment variable %q not defined", lsImageEnvVar)
	}

	controller := &Controller{
		kubeclientset:       kubeclientset,
		sampleclientset:     sampleclientset,
		deploymentsLister:   deploymentInformer.Lister(),
		deploymentsSynced:   deploymentInformer.Informer().HasSynced,
		eventBindingsLister: eventBindingInformer.Lister(),
		eventBindingsSynced: eventBindingInformer.Informer().HasSynced,
		workqueue:           workqueue.NewNamedRateLimitingQueue(workqueue.DefaultControllerRateLimiter(), "listenerTemplates"),
		recorder:            recorder,
		listenerImage:       listenerImage,
	}

	klog.Info("Setting up event handlers")
	// Set up an event handler for when Foo resources change
	eventBindingInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: controller.enqueueEventBinding,
		UpdateFunc: func(old, new interface{}) {
			klog.Info("Update event received......")
			oldBinding := old.(*samplev1alpha1.EventBinding)
			newBinding := new.(*samplev1alpha1.EventBinding)
			if oldBinding.ResourceVersion == newBinding.ResourceVersion {
				// Periodic resync will send update events for all known Networks.
				// Two different versions of the same Network will always have different RVs.
				return
			}
			controller.enqueueEventBinding(new)
		},
		DeleteFunc: controller.enqueueEventBindingforDelete,
	})

	return controller
}

// Run will set up the event handlers for types we are interested in, as well
// as syncing informer caches and starting workers. It will block until stopCh
// is closed, at which point it will shutdown the workqueue and wait for
// workers to finish processing their current work items.
func (c *Controller) Run(threadiness int, stopCh <-chan struct{}) error {
	defer utilruntime.HandleCrash()
	defer c.workqueue.ShutDown()

	// Start the informer factories to begin populating the informer caches
	klog.Info("Starting eventBinding controller")

	// Wait for the caches to be synced before starting workers
	klog.Info("Waiting for informer caches to sync")
	if ok := cache.WaitForCacheSync(stopCh, c.deploymentsSynced, c.eventBindingsSynced); !ok {
		return fmt.Errorf("failed to wait for caches to sync")
	}

	klog.Info("Starting workers")
	// Launch two workers to process Foo resources
	for i := 0; i < threadiness; i++ {
		go wait.Until(c.runWorker, time.Second, stopCh)
	}

	klog.Info("Started workers")
	<-stopCh
	klog.Info("Shutting down workers")

	return nil
}

// runWorker is a long-running function that will continually call the
// processNextWorkItem function in order to read and process a message on the
// workqueue.
func (c *Controller) runWorker() {
	for c.processNextWorkItem() {
	}
}

// processNextWorkItem will read a single work item off the workqueue and
// attempt to process it, by calling the syncHandler.
func (c *Controller) processNextWorkItem() bool {
	obj, shutdown := c.workqueue.Get()

	if shutdown {
		return false
	}

	// We wrap this block in a func so we can defer c.workqueue.Done.
	err := func(obj interface{}) error {
		// We call Done here so the workqueue knows we have finished
		// processing this item. We also must remember to call Forget if we
		// do not want this work item being re-queued. For example, we do
		// not call Forget if a transient error occurs, instead the item is
		// put back on the workqueue and attempted again after a back-off
		// period.
		defer c.workqueue.Done(obj)
		var key string
		var ok bool
		// We expect strings to come off the workqueue. These are of the
		// form namespace/name. We do this as the delayed nature of the
		// workqueue means the items in the informer cache may actually be
		// more up to date that when the item was initially put onto the
		// workqueue.
		if key, ok = obj.(string); !ok {
			// As the item in the workqueue is actually invalid, we call
			// Forget here else we'd go into a loop of attempting to
			// process a work item that is invalid.
			c.workqueue.Forget(obj)
			utilruntime.HandleError(fmt.Errorf("expected string in workqueue but got %#v", obj))
			return nil
		}
		// Run the syncHandler, passing it the namespace/name string of the
		// Foo resource to be synced.
		if err := c.syncHandler(key); err != nil {
			// Put the item back on the workqueue to handle any transient errors.
			c.workqueue.AddRateLimited(key)
			return fmt.Errorf("error syncing '%s': %s, requeuing", key, err.Error())
		}
		// Finally, if no error occurs we Forget this item so it does not
		// get queued again until another change happens.
		c.workqueue.Forget(obj)
		klog.Infof("Successfully synced '%s'", key)
		return nil
	}(obj)

	if err != nil {
		utilruntime.HandleError(err)
		return true
	}

	return true
}

// syncHandler compares the actual state with the desired, and attempts to
// converge the two. It then updates the Status block of the Foo resource
// with the current status of the resource.
func (c *Controller) syncHandler(key string) error {
	// Convert the namespace/name string into a distinct namespace and name
	ctx := context.TODO()
	namespace, name, err := cache.SplitMetaNamespaceKey(key)
	if err != nil {
		utilruntime.HandleError(fmt.Errorf("invalid resource key: %s", key))
		return nil
	}

	// Get the listenerTemplate resource with this namespace/name
	source, err := c.eventBindingsLister.EventBindings(namespace).Get(name)
	if err != nil {
		// The Foo resource may no longer exist, in which case we stop
		// processing.
		if errors.IsNotFound(err) {
			utilruntime.HandleError(fmt.Errorf("eventBinding '%s' in work queue no longer exists", key))
			return nil
		}

		return err
	}

	if source.Finalizers == nil {
		klog.Info("Add Finalizers to resource.")
		c.addFinalizer(source)
		err = c.updateEventBinding(source)
		if err != nil {
			return err
		}
	}

	if source.DeletionTimestamp != nil {
		err = c.finalize(source)
		if err != nil {
			//c.recorder.Event(listenerTemplate, corev1.EventTypeWarning, FailedDeleted, MessageResourceDeleteFailed)
			return err
		}
	}

	if err != nil {
		//c.recorder.Event(listenerTemplate, corev1.EventTypeWarning, FailedDeleted, MessageResourceDeleteFailed)
		return err
	}

	// ksvc, err := c.getOwnedService(ctx, source)
	ksvcName := name + "-ksvc"
	ksvc, err := c.ksvcLister.Services(namespace).Get(ksvcName)
	if err != nil {
		// The Foo resource may no longer exist, in which case we stop
		// processing.
		if errors.IsNotFound(err) {
			utilruntime.HandleError(fmt.Errorf("ksvc '%s' in work queue no longer exists", key))
			return nil
		}

		return err
	}

	c.recorder.Event(eventBinding, corev1.EventTypeNormal, SuccessSynced, MessageResourceSynced)
	return nil
}

func (c *Controller) updateEventBinding(eventBinding *samplev1alpha1.EventBinding) error {
	klog.Info("Update binding...........................")
	// NEVER modify objects from the store. It's a read-only, local cache.
	// You can use DeepCopy() to make a deep copy of original object and modify this copy
	// Or create a copy manually for better performance
	eventBindingCopy := eventBinding.DeepCopy()
	// fooCopy.Status.AvailableReplicas = deployment.Status.AvailableReplicas
	// If the CustomResourceSubresources feature gate is not enabled,
	// we must use Update instead of UpdateStatus to update the Status block of the Foo resource.
	// UpdateStatus will not allow changes to the Spec of the resource,
	// which is ideal for ensuring nothing other than resource status has been updated.
	_, err := c.sampleclientset.SamplecontrollerV1alpha1().EventBindings(eventBindingCopy.Namespace).Update(eventBindingCopy)
	return err
}

func (c *Controller) updateEventBindingStatus(eventBinding *samplev1alpha1.EventBinding) error {
	klog.Info("updateStatus............................")
	// NEVER modify objects from the store. It's a read-only, local cache.
	// You can use DeepCopy() to make a deep copy of original object and modify this copy
	// Or create a copy manually for better performance
	eventBindingCopy := eventBinding.DeepCopy()
	// fooCopy.Status.AvailableReplicas = deployment.Status.AvailableReplicas
	// If the CustomResourceSubresources feature gate is not enabled,
	// we must use Update instead of UpdateStatus to update the Status block of the Foo resource.
	// UpdateStatus will not allow changes to the Spec of the resource,
	// which is ideal for ensuring nothing other than resource status has been updated.
	_, err := c.sampleclientset.SamplecontrollerV1alpha1().EventBindings(eventBindingCopy.Namespace).UpdateStatus(eventBindingCopy)
	return err
}

// enqueueFoo takes a Foo resource and converts it into a namespace/name
// string which is then put onto the work queue. This method should *not* be
// passed resources of any type other than Foo.
func (c *Controller) enqueueEventBinding(obj interface{}) {
	var key string
	var err error
	if key, err = cache.MetaNamespaceKeyFunc(obj); err != nil {
		utilruntime.HandleError(err)
		return
	}
	c.workqueue.Add(key)
}

func (c *Controller) enqueueEventBindingforDelete(obj interface{}) {
	klog.Info("delete event received.............")
	var key string
	var err error
	if key, err = cache.MetaNamespaceKeyFunc(obj); err != nil {
		utilruntime.HandleError(err)
		return
	}

	c.workqueue.Add(key)
}

// newDeployment creates a new Deployment for a Foo resource. It also sets
// the appropriate OwnerReferences on the resource so handleObject can discover
// the Foo resource that 'owns' it.
func newDeployment(foo *samplev1alpha1.Foo) *appsv1.Deployment {
	labels := map[string]string{
		"app":        "nginx",
		"controller": foo.Name,
	}
	return &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      foo.Spec.DeploymentName,
			Namespace: foo.Namespace,
			OwnerReferences: []metav1.OwnerReference{
				*metav1.NewControllerRef(foo, samplev1alpha1.SchemeGroupVersion.WithKind("Foo")),
			},
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: foo.Spec.Replicas,
			Selector: &metav1.LabelSelector{
				MatchLabels: labels,
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: labels,
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name:  "nginx",
							Image: "nginx:latest",
						},
					},
				},
			},
		},
	}
}

func (c *Controller) addFinalizer(s *samplev1alpha1.EventBinding) {
	finalizers := sets.NewString(s.Finalizers...)
	finalizers.Insert(eventBindingFinalizerName)
	s.Finalizers = finalizers.List()
}

func (c *Controller) removeFinalizer(s *samplev1alpha1.EventBinding) {
	finalizers := sets.NewString(s.Finalizers...)
	finalizers.Delete(eventBindingFinalizerName)
	s.Finalizers = finalizers.List()
}

func (c *Controller) finalize(source *samplev1alpha1.EventBinding) error {
	// Always remove the finalizer. If there's a failure cleaning up, an event
	// will be recorded allowing the webhook to be removed manually by the
	// operator.
	klog.Info("Delete...............")
	c.removeFinalizer(source)
	c.updateEventBinding(source)
	return nil
}
