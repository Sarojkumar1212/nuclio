/*
Copyright 2017 The Nuclio Authors.

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

package command

import (
	"encoding/json"
	"os"

	"github.com/nuclio/nuclio/pkg/functionconfig"
	"github.com/nuclio/nuclio/pkg/platform"

	"github.com/nuclio/errors"
	"github.com/spf13/cobra"
)

type buildCommandeer struct {
	cmd                        *cobra.Command
	rootCommandeer             *RootCommandeer
	commands                   stringSliceFlag
	functionConfig             functionconfig.Config
	encodedRuntimeAttributes   string
	encodedCodeEntryAttributes string
	outputImageFile            string
}

func newBuildCommandeer(rootCommandeer *RootCommandeer) *buildCommandeer {
	commandeer := &buildCommandeer{
		rootCommandeer: rootCommandeer,
		functionConfig: *functionconfig.NewConfig(),
	}

	cmd := &cobra.Command{
		Use:     "build function-name [options]",
		Aliases: []string{"bu"},
		Short:   "Build a function",
		RunE: func(cmd *cobra.Command, args []string) error {

			// update build stuff
			if len(args) == 1 {
				commandeer.functionConfig.Meta.Name = args[0]
			}

			// initialize root
			if err := rootCommandeer.initialize(); err != nil {
				return errors.Wrap(err, "Failed to initialize root")
			}

			commandeer.functionConfig.Meta.Namespace = rootCommandeer.namespace
			commandeer.functionConfig.Spec.Build.Commands = commandeer.commands

			// decode the JSON build runtime attributes
			if err := json.Unmarshal([]byte(commandeer.encodedRuntimeAttributes),
				&commandeer.functionConfig.Spec.Build.RuntimeAttributes); err != nil {
				return errors.Wrap(err, "Failed to decode build runtime attributes")
			}

			if commandeer.functionConfig.Spec.Build.Offline {
				rootCommandeer.loggerInstance.Debug("Offline flag is passed, setting no-pull as well")
				commandeer.functionConfig.Spec.Build.NoBaseImagesPull = true
			}

			// decode the JSON build code entry attributes
			if err := json.Unmarshal([]byte(commandeer.encodedCodeEntryAttributes),
				&commandeer.functionConfig.Spec.Build.CodeEntryAttributes); err != nil {
				return errors.Wrap(err, "Failed to decode code entry attributes")
			}

			_, err := rootCommandeer.platform.CreateFunctionBuild(&platform.CreateFunctionBuildOptions{
				Logger:          rootCommandeer.loggerInstance,
				FunctionConfig:  commandeer.functionConfig,
				PlatformName:    rootCommandeer.platform.GetName(),
				OutputImageFile: commandeer.outputImageFile,
			})
			return err
		},
	}

	addBuildFlags(cmd, &commandeer.functionConfig, &commandeer.commands, &commandeer.encodedRuntimeAttributes, &commandeer.encodedCodeEntryAttributes)
	cmd.Flags().StringVarP(&commandeer.outputImageFile, "output-image-file", "", "", "Path to output container image of the build")

	commandeer.cmd = cmd

	return commandeer
}

func addBuildFlags(cmd *cobra.Command, config *functionconfig.Config, commands *stringSliceFlag, encodedRuntimeAttributes *string, encodedCodeEntryAttributes *string) { // nolint
	cmd.Flags().StringVarP(&config.Spec.Build.Path, "path", "p", "", "Path to the function's source code")
	cmd.Flags().StringVarP(&config.Spec.Build.FunctionSourceCode, "source", "", "", "The function's source code (overrides \"path\")")
	cmd.Flags().StringVarP(&config.Spec.Build.FunctionConfigPath, "file", "f", "", "Path to a function-configuration file")
	cmd.Flags().StringVarP(&config.Spec.Build.Image, "image", "i", "", "Name of a container image (default - the function name)")
	cmd.Flags().StringVarP(&config.Spec.Build.Registry, "registry", "r", os.Getenv("NUCTL_REGISTRY"), "URL of a container registry (env: NUCTL_REGISTRY)")
	cmd.Flags().StringVarP(&config.Spec.Runtime, "runtime", "", "", "Runtime (for example, \"golang\", \"python:3.6\")")
	cmd.Flags().StringVarP(&config.Spec.Handler, "handler", "", "", "Name of a function handler")
	cmd.Flags().BoolVarP(&config.Spec.Build.NoBaseImagesPull, "no-pull", "", false, "Don't pull base images - use local versions")
	cmd.Flags().BoolVarP(&config.Spec.Build.NoCleanup, "no-cleanup", "", false, "Don't clean up temporary directories")
	cmd.Flags().StringVarP(&config.Spec.Build.BaseImage, "base-image", "", "", "Name of the base image (default - per-runtime default)")
	cmd.Flags().Var(commands, "build-command", "Commands to run when building the processor image")
	cmd.Flags().StringVarP(&config.Spec.Build.OnbuildImage, "onbuild-image", "", "", "The runtime onbuild image used to build the processor image")
	cmd.Flags().BoolVarP(&config.Spec.Build.Offline, "offline", "", false, "Don't assume internet connectivity exists")
	cmd.Flags().StringVar(encodedRuntimeAttributes, "build-runtime-attrs", "{}", "JSON-encoded build runtime attributes for the function")
	cmd.Flags().StringVar(encodedCodeEntryAttributes, "build-code-entry-attrs", "{}", "JSON-encoded build code entry attributes for the function")
	cmd.Flags().StringVar(&config.Spec.Build.CodeEntryType, "code-entry-type", "", "Type of code entry (for example, \"url\", \"github\", \"image\")")
}
