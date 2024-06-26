//  Copyright 2024 Google LLC
//
//  Licensed under the Apache License, Version 2.0 (the "License");
//  you may not use this file except in compliance with the License.
//  You may obtain a copy of the License at
//
//       http://www.apache.org/licenses/LICENSE-2.0
//
//  Unless required by applicable law or agreed to in writing, software
//  distributed under the License is distributed on an "AS IS" BASIS,
//  WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
//  See the License for the specific language governing permissions and
//  limitations under the License.

package v1

import (
	"encoding/xml"
	"github.com/apigee/apigee-go-gen/pkg/utils"
	"github.com/go-errors/errors"
	"gopkg.in/yaml.v3"
	"path/filepath"
)

func NewSharedFlowBundleModel(input string) (*SharedFlowBundleModel, error) {
	sharedFlowModel := &SharedFlowBundleModel{}
	err := sharedFlowModel.Hydrate(input)
	if err != nil {
		return nil, err
	}
	return sharedFlowModel, nil
}

type SharedFlowBundleModel struct {
	SharedFlowBundle SharedFlowBundle `xml:"SharedFlowBundle"`
	Policies         Policies         `xml:"Policies"`
	Resources        Resources        `xml:"Resources"`
	SharedFlows      SharedFlows      `xml:"SharedFlows"`

	UnknownNode AnyList `xml:",any"`

	YAMLDoc *yaml.Node `xml:"-"`
}

func (a *SharedFlowBundleModel) Name() string {
	return a.SharedFlowBundle.Name
}

func (a *SharedFlowBundleModel) DisplayName() string {
	return a.SharedFlowBundle.DisplayName
}

func (a *SharedFlowBundleModel) Revision() int {
	return a.SharedFlowBundle.Revision
}

func (a *SharedFlowBundleModel) GetResources() *Resources {
	return &a.Resources
}

func (a *SharedFlowBundleModel) Hydrate(filePath string) error {
	var err error

	a.YAMLDoc, err = utils.YAMLFile2YAML(filePath)

	if err != nil {
		return err
	}

	wrapper := &yaml.Node{Kind: yaml.MappingNode}
	wrapper.Content = append(wrapper.Content, &yaml.Node{Kind: yaml.ScalarNode, Value: "SharedFlowBundleModel"}, a.YAMLDoc)

	xmlText, err := utils.YAML2XMLText(wrapper)
	if err != nil {
		return err
	}

	if err = xml.Unmarshal(xmlText, a); err != nil {
		return errors.New(err)
	}

	err = HydrateResources(a, filepath.Dir(filePath))
	if err != nil {
		return errors.New(err)
	}

	return nil
}

func (a *SharedFlowBundleModel) BundleRoot() string {
	return "sharedflowbundle"
}

func (a *SharedFlowBundleModel) BundleFiles() []BundleFile {
	bundleFiles := []BundleFile{
		BundleFile(&a.SharedFlowBundle),
	}

	for _, item := range a.SharedFlows.List {
		bundleFiles = append(bundleFiles, BundleFile(item))
	}

	for _, item := range a.Policies.List {
		bundleFiles = append(bundleFiles, BundleFile(item))
	}

	for _, item := range a.Resources.List {
		bundleFiles = append(bundleFiles, BundleFile(item))
	}

	return bundleFiles
}

func (a *SharedFlowBundleModel) XML() ([]byte, error) {
	return utils.Struct2XMLDocText(a)
}

func (a *SharedFlowBundleModel) YAML() ([]byte, error) {
	return utils.YAML2Text(a.YAMLDoc, 2)
}

func (a *SharedFlowBundleModel) Validate() error {
	if a == nil {
		return nil
	}

	err := utils.MultiError{Errors: []error{}}
	path := "Root"
	if len(a.UnknownNode) > 0 {
		err.Errors = append(err.Errors, &UnknownNodeError{path, a.UnknownNode[0]})
		return err
	}

	var subErrors []error
	subErrors = append(subErrors, ValidateSharedFlowBundle(&a.SharedFlowBundle, path)...)
	subErrors = append(subErrors, ValidateSharedFlows(&a.SharedFlows, path)...)
	subErrors = append(subErrors, ValidateResources(&a.Resources, path)...)

	if len(subErrors) > 0 {
		err.Errors = append(err.Errors, subErrors...)
		return err
	}

	return nil
}
