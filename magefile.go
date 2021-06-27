// +build mage

/*
Copyright The SOPS Operator Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    https://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package main

import (
	"github.com/magefile/mage/mg"
	"github.com/magefile/mage/sh"
)

func Lint() error {
	if err := sh.RunV("bash", "-c", "shopt -s globstar; shellcheck **/*.sh"); err != nil {
		return err
	}
	if err := sh.RunV("golangci-lint", "run", "--timeout", "2m"); err != nil {
		return err
	}
	if err := sh.RunV("go", "vet", "-v", "./..."); err != nil {
		return err
	}
	if err := sh.RunV("goimports", "-w", "-l", "."); err != nil {
		return err
	}
	if err := sh.RunV("go", "mod", "tidy"); err != nil {
		return err
	}
	return sh.RunV("git", "diff", "--exit-code")
}

func CheckLicenseHeaders() error {
	return sh.RunV("./check_license_headers.sh")
}

func ControllerGen() error {
	const crdOpts = "crd:trivialVersions=true,preserveUnknownFields=false"
	if err := sh.RunV("controller-gen", crdOpts, "paths=./...", "output:crd:artifacts:config=config/crd"); err != nil {
		return err
	}
	if err := sh.RunV("controller-gen", crdOpts, "object:headerFile=hack/boilerplate.go.txt"); err != nil {
		return err
	}
	return nil
}

func Test() error {
	return sh.RunV("go", "test", "./...", "-race")
}

func Build() error {
	return sh.RunV("goreleaser", "release", "--rm-dist", "--skip-publish", "--skip-validate")
}

func Release() error {
	mg.Deps(Test)
	return sh.RunV("goreleaser", "release", "--rm-dist")
}
