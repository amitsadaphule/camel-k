/*
Licensed to the Apache Software Foundation (ASF) under one or more
contributor license agreements.  See the NOTICE file distributed with
this work for additional information regarding copyright ownership.
The ASF licenses this file to You under the Apache License, Version 2.0
(the "License"); you may not use this file except in compliance with
the License.  You may obtain a copy of the License at

   http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package cmd

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"strings"

	"github.com/apache/camel-k/pkg/util/docker"
	"github.com/pkg/errors"
)

/// Local Docker file system management functions.

func createDockerBaseWorkingDirectory() error {
	// Create local docker base directory.
	temporaryDirectory, err := ioutil.TempDir(os.TempDir(), "docker-base-")
	if err != nil {
		return err
	}

	// Set the Docker base directory to the default value.
	docker.BaseWorkingDirectory = temporaryDirectory

	return nil
}

func deleteDockerBaseWorkingDirectory() error {
	// Remove directory used for computing the dependencies.
	defer os.RemoveAll(docker.BaseWorkingDirectory)

	return nil
}

func createDockerWorkingDirectory() error {
	// Create local docker base directory.
	temporaryDirectory, err := ioutil.TempDir(os.TempDir(), "docker-")
	if err != nil {
		return err
	}

	// Set the Docker base directory to the default value.
	docker.IntegrationWorkingDirectory = temporaryDirectory

	return nil
}

func deleteDockerWorkingDirectory() error {
	// Remove directory used for computing the dependencies.
	defer os.RemoveAll(docker.IntegrationWorkingDirectory)

	return nil
}

func createAndBuildBaseImage(dockerRegistry string) error {
	// Set docker registry.
	docker.RegistryName = dockerRegistry

	// Create the base image Docker file.
	err := docker.CreateBaseImageDockerFile()
	if err != nil {
		return err
	}

	// Get the Docker command arguments for building the base image and create the command.
	args := docker.BuildBaseImageArgs()
	cmd := exec.CommandContext(ctx, "docker", args...)

	// Output executed command.
	fmt.Printf("Executing: " + strings.Join(cmd.Args, " ") + "\n")

	// Run the command.
	if err := cmd.Run(); err != nil {
		errors.Errorf("base image containerization did not run successfully: %v", err)
	}

	return nil
}

func createAndBuildIntegrationImage(dockerRegistry string, justBaseImage bool, imageName string,
	propertyFiles []string, dependencies []string, routes []string) error {
	docker.RegistryName = dockerRegistry

	// Create the Dockerfile and build the base image.
	err := createAndBuildBaseImage(dockerRegistry)
	if err != nil {
		return err
	}

	// Create integration image if integration files were provided.
	if !justBaseImage {
		// Copy dependencies to a dependencies folder under a local directory.
		err = updateIntegrationDependencies(dependencies)
		if err != nil {
			return err
		}

		// Copy routes to a routes folder under a local directory.
		err = updateIntegrationRoutes(routes)
		if err != nil {
			return err
		}

		// Get integration run command to be run inside the container. This means the command
		// has to be created with the paths which will be valid inside the container.
		containerCmd := GetContainerIntegrationRunCommand(propertyFiles, dependencies, routes)

		// Create the integration image Docker file.
		err = docker.CreateIntegrationImageDockerFile(containerCmd)
		if err != nil {
			return err
		}

		// Get the Docker command arguments for building the base image and create the command.
		args := docker.BuildIntegrationImageArgs(imageName)
		cmd := exec.CommandContext(ctx, "docker", args...)

		// Output executed command.
		fmt.Printf("Executing: " + strings.Join(cmd.Args, " ") + "\n")

		// Run the command.
		if err := cmd.Run(); err != nil {
			errors.Errorf("integration image containerization did not run successfully: %v", err)
		}
	}

	return nil
}

func runIntegrationImage(dockerRegistry string, imageName string) error {
	docker.RegistryName = dockerRegistry

	// Get the docker command line argument for running an image.
	args := docker.RunIntegrationImageArgs(imageName)
	cmd := exec.CommandContext(ctx, "docker", args...)

	// Output executed command.
	fmt.Printf("Executing: " + strings.Join(cmd.Args, " ") + "\n")

	// Run the command.
	if err := cmd.Run(); err != nil {
		errors.Errorf("integration image did not run successfully: %v", err)
	}

	return nil
}