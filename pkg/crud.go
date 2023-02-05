package pkg

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/sirupsen/logrus"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/tools/clientcmd"
)

// FixtureLoader loads fixtures
type FixtureLoader struct {
	Dynamic dynamic.Interface
	Log     *logrus.Logger
}

// ResourceInfo holds the information about a resource
type ResourceInfo struct {
	GVR            schema.GroupVersionResource
	NamespacedName types.NamespacedName
}

// SetupFixtureLoader sets up a FixtureLoader
func SetupFixtureLoader(dynamic dynamic.Interface, log *logrus.Logger) *FixtureLoader {
	return &FixtureLoader{
		Dynamic: dynamic,
		Log:     log,
	}
}

// SetupDynamicClient sets up a dynamic client
func SetupDynamicClient(kubeconfig *string) (dynamic.Interface, error) {
	envKubeconfig := PointerTo(os.Getenv("KUBECONFIG"))
	if kubeconfig == nil {
		kubeconfig = envKubeconfig
	}

	if envKubeconfig != nil {
		kubeconfig = envKubeconfig
	}

	config, err := clientcmd.BuildConfigFromFlags("", *kubeconfig)
	if err != nil {
		panic(err.Error())
	}

	dynamic := dynamic.NewForConfigOrDie(config)
	return dynamic, nil
}

// GetResourceDynamically returns a resource from the cluster
func (f *FixtureLoader) GetResourceDynamically(ctx context.Context, r *ResourceInfo) (*unstructured.Unstructured,
	error) {
	if r == nil {
		return nil, fmt.Errorf("ResourceInfo is nil")
	}

	f.Log.Infof("Getting resource %s/%s", r.NamespacedName.Namespace, r.NamespacedName.Name)

	resourceID := schema.GroupVersionResource{
		Group:    r.GVR.Group,
		Version:  r.GVR.Version,
		Resource: r.GVR.Resource,
	}
	obj, err := f.Dynamic.Resource(resourceID).Namespace(r.NamespacedName.Namespace).
		Get(ctx, r.NamespacedName.Name, metav1.GetOptions{})
	if err != nil {
		f.Log.Errorf("Failed to get resource %s %s/%s: %v", r.GVR.Resource,
			r.NamespacedName.Namespace,
			r.NamespacedName.Name,
			err)
		return nil, err
	}

	f.Log.Infof("Got resource %s %s/%s", r.GVR.Resource, r.NamespacedName.Namespace, r.NamespacedName.Name)

	return obj, nil
}

// CreateResourceDynamically creates a resource in the cluster
func (f *FixtureLoader) CreateResourceDynamically(ctx context.Context,
	r *ResourceInfo, obj unstructured.Unstructured) error {
	if r == nil {
		return fmt.Errorf("ResourceInfo is nil")
	}

	f.Log.Infof("Creating resource %s %s/%s", r.GVR.Resource, r.NamespacedName.Namespace, r.NamespacedName.Name)

	resourceID := schema.GroupVersionResource{
		Group:    r.GVR.Group,
		Version:  r.GVR.Version,
		Resource: r.GVR.Resource,
	}
	_, err := f.Dynamic.Resource(resourceID).Namespace(r.NamespacedName.Namespace).
		Create(ctx, &obj, metav1.CreateOptions{})
	if err != nil {
		f.Log.Errorf("Failed to create resource %s %s/%s: %v", r.GVR.Resource,
			r.NamespacedName.Namespace,
			r.NamespacedName.Name,
			err)
		return err
	}

	f.Log.Infof("Created resource %s %s/%s", r.GVR.Resource, r.NamespacedName.Namespace, r.NamespacedName.Name)
	return nil
}

// StatusLoad gets the resouce from the cluster, copies the status from the given object to the object from the cluster and updates the resource in the cluster
func (f *FixtureLoader) StatusLoad(ctx context.Context, ri *ResourceInfo, obj unstructured.Unstructured) error {
	newObj, err := f.GetResourceDynamically(ctx, ri)
	if err != nil {
		f.Log.Errorf("Failed to get resource %s %s/%s: %v", ri.GVR.Resource, ri.NamespacedName.Namespace, ri.NamespacedName.Name, err)
		return err
	}

	err = f.CopyStatus(&obj, newObj)
	if err != nil {
		f.Log.Errorf("Failed to copy status from %s %s/%s to %s %s/%s: %v", ri.GVR.Resource, ri.NamespacedName.Namespace, ri.NamespacedName.Name, ri.GVR.Resource, ri.NamespacedName.Namespace, ri.NamespacedName.Name, err)
		return err
	}

	err = f.UpdateResourceStatusDynamically(ctx, ri, *newObj)
	if err != nil {
		f.Log.Errorf("Failed to update resource %s %s/%s: %v", ri.GVR.Resource, ri.NamespacedName.Namespace, ri.NamespacedName.Name, err)
		return err
	}

	return nil
}

// StatusFieldLoad gets the resouce from the cluster, copies the status from the given object to the object from the cluster and updates the resource in the cluster
func (f *FixtureLoader) StatusFieldLoad(ctx context.Context, ri *ResourceInfo, obj unstructured.Unstructured, field string) error {
	newObj, err := f.GetResourceDynamically(ctx, ri)
	if err != nil {
		f.Log.Errorf("Failed to get resource %s %s/%s: %v", ri.GVR.Resource, ri.NamespacedName.Namespace, ri.NamespacedName.Name, err)
		return err
	}

	// err = f.CopyStatus(&obj, newObj, field)
	if err != nil {
		f.Log.Errorf("Failed to copy status from %s %s/%s to %s %s/%s: %v", ri.GVR.Resource, ri.NamespacedName.Namespace, ri.NamespacedName.Name, ri.GVR.Resource, ri.NamespacedName.Namespace, ri.NamespacedName.Name, err)
		return err
	}

	err = f.UpdateResourceStatusDynamically(ctx, ri, *newObj)
	if err != nil {
		f.Log.Errorf("Failed to update resource %s %s/%s: %v", ri.GVR.Resource, ri.NamespacedName.Namespace, ri.NamespacedName.Name, err)
		return err
	}

	return nil
}

// GetResourceInfo returns a ResourceInfo object for the given object
func GetResourceInfo(obj unstructured.Unstructured) *ResourceInfo {
	apiVersion := obj.GetAPIVersion()
	kind := fmt.Sprintf("%ss", strings.ToLower(obj.GetKind()))

	group := strings.Split(apiVersion, "/")[0]
	version := strings.Split(apiVersion, "/")[1]

	ns := obj.GetNamespace()
	if ns == "" {
		ns = "default"
	}

	ri := &ResourceInfo{
		GVR: schema.GroupVersionResource{
			Group:    group,
			Version:  version,
			Resource: kind,
		},
		NamespacedName: types.NamespacedName{
			Namespace: ns,
			Name:      obj.GetName(),
		},
	}

	return ri
}
