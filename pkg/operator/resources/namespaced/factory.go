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

package namespaced

import (
	"fmt"

	"k8s.io/apimachinery/pkg/runtime"

	cdiv1 "kubevirt.io/containerized-data-importer/pkg/apis/core/v1beta1"
	utils "kubevirt.io/containerized-data-importer/pkg/operator/resources/utils"
)

// FactoryArgs contains the required parameters to generate all namespaced resources
type FactoryArgs struct {
	OperatorVersion        string `required:"true" split_words:"true"`
	ControllerImage        string `required:"true" split_words:"true"`
	DeployClusterResources string `required:"true" split_words:"true"`
	ImporterImage          string `required:"true" split_words:"true"`
	ClonerImage            string `required:"true" split_words:"true"`
	APIServerImage         string `required:"true" envconfig:"apiserver_image"`
	UploadProxyImage       string `required:"true" split_words:"true"`
	UploadServerImage      string `required:"true" split_words:"true"`
	Verbosity              string `required:"true"`
	PullPolicy             string `required:"true" split_words:"true"`
	Namespace              string
	InfraNodePlacement     cdiv1.NodePlacement
}

type factoryFunc func(*FactoryArgs) []runtime.Object

type namespaceHaver interface {
	SetNamespace(string)
	GetNamespace() string
}

var factoryFunctions = map[string]factoryFunc{
	"apiserver":   createAPIServerResources,
	"controller":  createControllerResources,
	"uploadproxy": createUploadProxyResources,
}

// CreateAllResources creates all namespaced resources
func CreateAllResources(args *FactoryArgs) ([]runtime.Object, error) {
	var resources []runtime.Object
	for group := range factoryFunctions {
		rs, err := CreateResourceGroup(group, args)
		if err != nil {
			return nil, err
		}
		resources = append(resources, rs...)
	}
	return resources, nil
}

// CreateResourceGroup creates namespaced resources for a specific group/component
func CreateResourceGroup(group string, args *FactoryArgs) ([]runtime.Object, error) {
	f, ok := factoryFunctions[group]
	if !ok {
		return nil, fmt.Errorf("group %s does not exist", group)
	}
	resources := f(args)
	utils.ValidateGVKs(resources)
	for _, resource := range resources {
		assignNamspaceIfMissing(resource, args.Namespace)
	}
	return resources, nil
}

func assignNamspaceIfMissing(resource runtime.Object, namespace string) {
	obj, ok := resource.(namespaceHaver)
	if !ok {
		return
	}

	if obj.GetNamespace() == "" {
		obj.SetNamespace(namespace)
	}
}
