package helper

import (
	"context"
	"fmt"

	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/util/retry"
)

// GVR returns GroupVersionResource for an Istio CRD.
func GVR(group, version, resource string) schema.GroupVersionResource {
	return schema.GroupVersionResource{Group: group, Version: version, Resource: resource}
}

// GetCRD fetches a single custom resource by name and namespace.
func GetCRD(ctx context.Context, dc dynamic.Interface, gvr schema.GroupVersionResource, namespace, name string) (*unstructured.Unstructured, error) {
	var obj *unstructured.Unstructured
	var err error
	if namespace != "" {
		obj, err = dc.Resource(gvr).Namespace(namespace).Get(ctx, name, metav1.GetOptions{})
	} else {
		obj, err = dc.Resource(gvr).Get(ctx, name, metav1.GetOptions{})
	}
	if err != nil {
		return nil, err
	}
	return obj, nil
}

// CreateCRD creates a custom resource. When dryRun is true, uses server-side dry-run only.
func CreateCRD(ctx context.Context, dc dynamic.Interface, gvr schema.GroupVersionResource, obj *unstructured.Unstructured, dryRun bool) error {
	opts := metav1.CreateOptions{}
	if dryRun {
		opts.DryRun = []string{metav1.DryRunAll}
	}
	ns := obj.GetNamespace()
	if ns != "" {
		_, err := dc.Resource(gvr).Namespace(ns).Create(ctx, obj, opts)
		return err
	}
	_, err := dc.Resource(gvr).Create(ctx, obj, opts)
	return err
}

// UpdateCRD updates a custom resource. When dryRun is true, uses server-side dry-run only.
func UpdateCRD(ctx context.Context, dc dynamic.Interface, gvr schema.GroupVersionResource, obj *unstructured.Unstructured, dryRun bool) error {
	opts := metav1.UpdateOptions{}
	if dryRun {
		opts.DryRun = []string{metav1.DryRunAll}
	}
	ns := obj.GetNamespace()
	return retry.RetryOnConflict(retry.DefaultRetry, func() error {
		var current *unstructured.Unstructured
		var err error
		if ns != "" {
			current, err = dc.Resource(gvr).Namespace(ns).Get(ctx, obj.GetName(), metav1.GetOptions{})
		} else {
			current, err = dc.Resource(gvr).Get(ctx, obj.GetName(), metav1.GetOptions{})
		}
		if err != nil {
			return err
		}
		obj.SetResourceVersion(current.GetResourceVersion())
		if ns != "" {
			_, err = dc.Resource(gvr).Namespace(ns).Update(ctx, obj, opts)
		} else {
			_, err = dc.Resource(gvr).Update(ctx, obj, opts)
		}
		return err
	})
}

// DeleteCRD deletes a custom resource.
func DeleteCRD(ctx context.Context, dc dynamic.Interface, gvr schema.GroupVersionResource, namespace, name string) error {
	if namespace != "" {
		return dc.Resource(gvr).Namespace(namespace).Delete(ctx, name, metav1.DeleteOptions{})
	}
	return dc.Resource(gvr).Delete(ctx, name, metav1.DeleteOptions{})
}

// ManifestToUnstructured converts a map (from JSON) to Unstructured. Strips server-side fields.
func ManifestToUnstructured(manifest map[string]interface{}) (*unstructured.Unstructured, error) {
	StripServerFields(manifest)
	u := &unstructured.Unstructured{Object: manifest}
	return u, nil
}

// StripServerFields removes Kubernetes server-side metadata from an object map
// so that Terraform state does not drift on fields the user never set.
func StripServerFields(obj map[string]interface{}) {
	delete(obj, "status")
	if meta, ok := obj["metadata"].(map[string]interface{}); ok {
		delete(meta, "resourceVersion")
		delete(meta, "uid")
		delete(meta, "creationTimestamp")
		delete(meta, "generation")
		delete(meta, "managedFields")
		// Remove annotations added by kubectl / server-side apply
		if ann, ok := meta["annotations"].(map[string]interface{}); ok {
			delete(ann, "kubectl.kubernetes.io/last-applied-configuration")
			if len(ann) == 0 {
				delete(meta, "annotations")
			}
		}
	}
}

// IdFromObject returns a Terraform ID: namespace/name or name for cluster-scoped.
func IdFromObject(obj *unstructured.Unstructured) string {
	ns := obj.GetNamespace()
	name := obj.GetName()
	if ns != "" {
		return fmt.Sprintf("%s/%s", ns, name)
	}
	return name
}

// ParseId returns namespace and name from id (namespace/name or name).
func ParseId(id string) (namespace, name string) {
	for i := len(id) - 1; i >= 0; i-- {
		if id[i] == '/' {
			return id[:i], id[i+1:]
		}
	}
	return "", id
}

// IsNotFound returns true if the error is a Kubernetes 404.
func IsNotFound(err error) bool {
	return errors.IsNotFound(err)
}
