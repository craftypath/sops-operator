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

package sops

import (
	"bytes"
	"fmt"
	"os/exec"
	"path/filepath"
	"regexp"

	logf "sigs.k8s.io/controller-runtime/pkg/log"
)

type Decryptor struct{}

var (
	log = logf.Log.WithName("sops")

	fileFormats = map[string]string{
		".yaml": "yaml",
		".yml":  "yaml",
		".json": "json",
		".ini":  "ini",
		".env":  "dotenv",
	}
)

// Decrypt decrypts the given encrypted string. The format (yaml, json, dotenv, init, binary)
// is determined by the given fileName.
func (d *Decryptor) Decrypt(fileName string, encrypted string) ([]byte, error) {
	format := determineFileFormat(fileName)
	regionMatcher := regexp.MustCompile(`arn:aws:kms:([\w\d-]+):.*`)
	matches := regionMatcher.FindStringSubmatch(encrypted)
	if len(matches) < 2 {
		return nil, fmt.Errorf("failed to detect region from encrypted string, got matches %v", matches)
	}
	args := []string{"--decrypt", "--aws-endpoint", fmt.Sprintf("https://kms-fips.%s.amazonaws.com", matches[1]), format, "--output-type", format, "/dev/stdin"}
	log.V(1).Info("running sops", "args", args)

	// We shell out to SOPS because that way we get better error messages
	command := exec.Command("sops", args...)
	command.Stdin = bytes.NewBufferString(encrypted)

	output, err := command.Output()
	if err != nil {
		if e, ok := err.(*exec.ExitError); ok {
			return nil, fmt.Errorf("failed to decrypt file: %s", string(e.Stderr))
		}
		return nil, err
	}
	return output, err
}

func determineFileFormat(fileName string) string {
	ext := filepath.Ext(fileName)
	if format, exists := fileFormats[ext]; exists {
		return format
	}
	return "binary"
}
