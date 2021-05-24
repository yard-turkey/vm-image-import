/*
Copyright 2018 The CDI Authors.

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

// Code generated by informer-gen. DO NOT EDIT.

package externalversions

import (
	"fmt"

	schema "k8s.io/apimachinery/pkg/runtime/schema"
	cache "k8s.io/client-go/tools/cache"
	v1alpha1 "kubevirt.io/containerized-data-importer/pkg/apis/core/v1alpha1"
	v1beta1 "kubevirt.io/containerized-data-importer/pkg/apis/core/v1beta1"
	uploadv1alpha1 "kubevirt.io/containerized-data-importer/pkg/apis/upload/v1alpha1"
	uploadv1beta1 "kubevirt.io/containerized-data-importer/pkg/apis/upload/v1beta1"
)

// GenericInformer is type of SharedIndexInformer which will locate and delegate to other
// sharedInformers based on type
type GenericInformer interface {
	Informer() cache.SharedIndexInformer
	Lister() cache.GenericLister
}

type genericInformer struct {
	informer cache.SharedIndexInformer
	resource schema.GroupResource
}

// Informer returns the SharedIndexInformer.
func (f *genericInformer) Informer() cache.SharedIndexInformer {
	return f.informer
}

// Lister returns the GenericLister.
func (f *genericInformer) Lister() cache.GenericLister {
	return cache.NewGenericLister(f.Informer().GetIndexer(), f.resource)
}

// ForResource gives generic access to a shared informer of the matching type
// TODO extend this to unknown resources with a client pool
func (f *sharedInformerFactory) ForResource(resource schema.GroupVersionResource) (GenericInformer, error) {
	switch resource {
	// Group=cdi.kubevirt.io, Version=v1alpha1
	case v1alpha1.SchemeGroupVersion.WithResource("cdis"):
		return &genericInformer{resource: resource.GroupResource(), informer: f.Cdi().V1alpha1().CDIs().Informer()}, nil
	case v1alpha1.SchemeGroupVersion.WithResource("cdiconfigs"):
		return &genericInformer{resource: resource.GroupResource(), informer: f.Cdi().V1alpha1().CDIConfigs().Informer()}, nil
	case v1alpha1.SchemeGroupVersion.WithResource("datavolumes"):
		return &genericInformer{resource: resource.GroupResource(), informer: f.Cdi().V1alpha1().DataVolumes().Informer()}, nil

		// Group=cdi.kubevirt.io, Version=v1beta1
	case v1beta1.SchemeGroupVersion.WithResource("cdis"):
		return &genericInformer{resource: resource.GroupResource(), informer: f.Cdi().V1beta1().CDIs().Informer()}, nil
	case v1beta1.SchemeGroupVersion.WithResource("cdiconfigs"):
		return &genericInformer{resource: resource.GroupResource(), informer: f.Cdi().V1beta1().CDIConfigs().Informer()}, nil
	case v1beta1.SchemeGroupVersion.WithResource("datasources"):
		return &genericInformer{resource: resource.GroupResource(), informer: f.Cdi().V1beta1().DataSources().Informer()}, nil
	case v1beta1.SchemeGroupVersion.WithResource("datavolumes"):
		return &genericInformer{resource: resource.GroupResource(), informer: f.Cdi().V1beta1().DataVolumes().Informer()}, nil
	case v1beta1.SchemeGroupVersion.WithResource("objecttransfers"):
		return &genericInformer{resource: resource.GroupResource(), informer: f.Cdi().V1beta1().ObjectTransfers().Informer()}, nil
	case v1beta1.SchemeGroupVersion.WithResource("storageprofiles"):
		return &genericInformer{resource: resource.GroupResource(), informer: f.Cdi().V1beta1().StorageProfiles().Informer()}, nil

		// Group=upload.cdi.kubevirt.io, Version=v1alpha1
	case uploadv1alpha1.SchemeGroupVersion.WithResource("uploadtokenrequests"):
		return &genericInformer{resource: resource.GroupResource(), informer: f.Upload().V1alpha1().UploadTokenRequests().Informer()}, nil

		// Group=upload.cdi.kubevirt.io, Version=v1beta1
	case uploadv1beta1.SchemeGroupVersion.WithResource("uploadtokenrequests"):
		return &genericInformer{resource: resource.GroupResource(), informer: f.Upload().V1beta1().UploadTokenRequests().Informer()}, nil

	}

	return nil, fmt.Errorf("no informer found for %v", resource)
}
