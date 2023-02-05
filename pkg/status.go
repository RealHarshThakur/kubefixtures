package pkg

import (
	"context"
	"fmt"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

// UpdateResourceStatusDynamically updates a resource in the cluster.
// Note: it doesn't retry on conflict errors.
func (f *FixtureLoader) UpdateResourceStatusDynamically(ctx context.Context,
	r *ResourceInfo, obj unstructured.Unstructured) error {

	f.Log.Infof("Updating status of resource %s %s/%s", r.GVR.Resource, r.NamespacedName.Namespace, r.NamespacedName.Name)

	resourceID := schema.GroupVersionResource{
		Group:    r.GVR.Group,
		Version:  r.GVR.Version,
		Resource: r.GVR.Resource,
	}
	_, err := f.Dynamic.Resource(resourceID).Namespace(r.NamespacedName.Namespace).UpdateStatus(ctx, &obj, metav1.UpdateOptions{})
	if err != nil {
		f.Log.Errorf("Failed to update status of resource %s %s/%s: %v", r.GVR.Resource, r.NamespacedName.Namespace, r.NamespacedName.Name, err)
		return err
	}

	f.Log.Infof("Updated status of resource %s %s/%s", r.GVR.Resource, r.NamespacedName.Namespace, r.NamespacedName.Name)

	return err
}

// CopyStatus copies the status from one object to another.
func (f *FixtureLoader) CopyStatus(src, dst *unstructured.Unstructured) error {
	f.Log.Infof("Copying status from %s/%s to %s/%s", src.GetNamespace(), src.GetName(), dst.GetNamespace(), dst.GetName())

	srcStatus, found, err := unstructured.NestedMap(src.Object, "status")
	if err != nil || !found {
		f.Log.Errorf("Failed to get status from source object: %v", err)
		return fmt.Errorf("Failed to get status from source object: %v", err)
	}

	if err := unstructured.SetNestedMap(dst.Object, srcStatus, "status"); err != nil {
		f.Log.Errorf("Failed to set status to destination object: %v", err)
		return fmt.Errorf("Failed to set status to destination object: %v", err)
	}
	return nil
}

// UpdateStatusField updates a field in the status of an object.
func UpdateStatusField(obj *unstructured.Unstructured, field string, value interface{}) error {
	statusMap, found, err := unstructured.NestedMap(obj.Object, "status")
	if err != nil || !found {
		return fmt.Errorf("Failed to retrieve status map: %v", err)
	}

	statusMap[field] = value

	if err := unstructured.SetNestedMap(obj.Object, statusMap, "status"); err != nil {
		return fmt.Errorf("Failed to update status field: %v", err)
	}

	return nil
}
