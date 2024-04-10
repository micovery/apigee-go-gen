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

package main

import (
	"flag"
	"github.com/micovery/apigee-yaml-toolkit/cmd/render-graphql/resources"
	"github.com/micovery/apigee-yaml-toolkit/pkg/render"
	"github.com/micovery/apigee-yaml-toolkit/pkg/utils"
	"strings"
)

func main() {

	var schemaFile string

	flag.StringVar(&schemaFile, "schema", "", `(required) path to proto file e.g "./greeter.proto"`)
	flags, err := render.GetRenderFlags(utils.PrintVersion, resources.PrintUsage)
	if err != nil {
		utils.PrintErrorWithStackAndExit(err)
		return
	}

	if strings.TrimSpace(schemaFile) == "" {
		utils.RequireParamAndExit("schema")
		return
	}

	err = render.RenderGraphQL(schemaFile, flags)
	if err != nil {
		utils.PrintErrorWithStackAndExit(err)
		return
	}
}
