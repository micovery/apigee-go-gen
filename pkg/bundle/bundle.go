// Copyright 2024 Google LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package bundle

import (
	"bytes"
	"fmt"
	"github.com/go-errors/errors"
	"github.com/micovery/apigee-yaml-toolkit/pkg/utils"
	"github.com/micovery/apigee-yaml-toolkit/pkg/zip"
	"gopkg.in/yaml.v3"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"strings"
)

func ProxyBundle2YAML(proxyBundle string, outputFile string) error {
	extension := filepath.Ext(proxyBundle)
	if extension == ".zip" {
		err := ProxyBundleZip2YAML(proxyBundle, outputFile)
		if err != nil {
			return err
		}
	} else if extension != "" {
		return errors.Errorf("input extension %s is not supported", extension)
	} else {
		err := ProxyBundleDir2YAML(proxyBundle, outputFile)
		if err != nil {
			return err
		}
	}

	return nil
}

func ProxyBundleZip2YAML(inputZip string, outputFile string) error {
	tmpDir, err := os.MkdirTemp("", "unzipped-bundle-*")
	if err != nil {
		return errors.New(err)
	}

	err = zip.Unzip(tmpDir, inputZip)
	if err != nil {
		return errors.New(err)
	}

	return ProxyBundleDir2YAML(tmpDir, outputFile)

}

func ProxyBundleDir2YAML(inputDir string, outputFile string) error {
	policyFiles := []string{}
	proxyEndpointsFiles := []string{}
	targetEndpointsFiles := []string{}
	resourcesFiles := []string{}
	manifestFiles := []string{}

	apiProxyDir := filepath.Join(inputDir, "apiproxy")
	stat, err := os.Stat(apiProxyDir)
	if err != nil {
		return errors.Errorf("%s not found. %s", err.Error())
	} else if !stat.IsDir() {
		return errors.Errorf("%s is not a directory")
	}

	fSys := os.DirFS(apiProxyDir)

	manifestFiles, _ = fs.Glob(fSys, "*.xml")
	policyFiles, _ = fs.Glob(fSys, "policies/*.xml")
	proxyEndpointsFiles, _ = fs.Glob(fSys, "proxies/*.xml")
	targetEndpointsFiles, _ = fs.Glob(fSys, "targets/*.xml")
	resourcesFiles, _ = fs.Glob(fSys, "resources/*/*")

	allFiles := []string{}
	if len(manifestFiles) == 0 {
		return errors.Errorf("no proxy XML file found in %s", apiProxyDir)
	}

	allFiles = append(allFiles, manifestFiles[0])
	allFiles = append(allFiles, policyFiles...)
	allFiles = append(allFiles, proxyEndpointsFiles...)
	allFiles = append(allFiles, targetEndpointsFiles...)

	createMapEntry := func(parent *yaml.Node, key string, value *yaml.Node) *yaml.Node {
		parent.Content = append(parent.Content, &yaml.Node{Kind: yaml.ScalarNode, Value: key}, value)
		return value
	}

	fileToYAML := func(filePath string) (*yaml.Node, error) {
		fullPath := filepath.Join(apiProxyDir, filePath)
		fileContents, err := os.ReadFile(fullPath)
		if err != nil {
			return nil, errors.New(err)
		}
		yamlNode, err := utils.XMLText2YAML(bytes.NewReader(fileContents))
		if err != nil {
			return nil, err
		}

		return yamlNode, nil
	}

	addSequence := func(parentNode *yaml.Node, key string, files []string) error {
		sequence := createMapEntry(parentNode, key, &yaml.Node{Kind: yaml.SequenceNode})
		for _, filePath := range files {
			yamlNode, err := fileToYAML(filePath)
			if err != nil {
				return err
			}
			if len(yamlNode.Content) > 0 {
				sequence.Content = append(sequence.Content, yamlNode)
			}
		}
		return nil
	}

	docNode := &yaml.Node{Kind: yaml.DocumentNode}
	mainNode := &yaml.Node{Kind: yaml.MappingNode}
	docNode.Content = append(docNode.Content, mainNode)

	manifestNode, err := fileToYAML(manifestFiles[0])
	if err != nil {
		return err
	}

	mainNode.Content = append(mainNode.Content, manifestNode.Content...)

	err = addSequence(mainNode, "Policies", policyFiles)
	if err != nil {
		return err
	}
	err = addSequence(mainNode, "ProxyEndpoints", proxyEndpointsFiles)
	if err != nil {
		return err
	}
	err = addSequence(mainNode, "TargetEndpoints", targetEndpointsFiles)
	if err != nil {
		return err
	}

	//copy resource files
	resourcesNode := createMapEntry(mainNode, "Resources", &yaml.Node{Kind: yaml.MappingNode})
	for _, resourceFile := range resourcesFiles {
		dirName, fileName := filepath.Split(resourceFile)
		fileType := filepath.Base(dirName)

		location := path.Join(".", fileName)
		resourceDataNode := createMapEntry(resourcesNode, "Resource", &yaml.Node{Kind: yaml.MappingNode})
		createMapEntry(resourceDataNode, "Type", &yaml.Node{Kind: yaml.ScalarNode, Value: fileType})
		createMapEntry(resourceDataNode, "Path", &yaml.Node{Kind: yaml.ScalarNode, Value: fmt.Sprintf("./%s", location)})

		outputDir := filepath.Dir(outputFile)
		err := utils.CopyFile(filepath.Join(outputDir, fileName), filepath.Join(apiProxyDir, resourceFile))
		if err != nil {
			return err
		}
	}

	err = WriteYAMLDocToDisk(docNode, outputFile, manifestFiles[0])
	if err != nil {
		return err
	}

	return nil
}

func WriteYAMLDocToDisk(docNode *yaml.Node, outputFile string, fileName string) error {
	var err error
	var docBytes []byte
	if docBytes, err = utils.YAML2Text(docNode, 2); err != nil {
		return err
	}

	//generate output directory
	outputDir := filepath.Dir(outputFile)
	if err = os.MkdirAll(outputDir, os.ModePerm); err != nil {
		return errors.New(err)
	}

	//generate the main YAML file
	fileName = filepath.Base(fileName)
	fileName = fmt.Sprintf("%s.yaml", strings.TrimSuffix(fileName, filepath.Ext(fileName)))
	if err = os.WriteFile(outputFile, docBytes, os.ModePerm); err != nil {
		return err
	}
	return nil
}