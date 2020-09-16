// Copyright The SOPS Operator Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     https://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package sops

import (
	"bytes"
	"os/exec"
	"path/filepath"

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
	args := []string{"--decrypt", "--input-type", format, "--output-type", format, "/dev/stdin"}
	log.V(1).Info("running sops", "args", args)

	// We shell out to SOPS because that way we get better error messages
	command := exec.Command("sops", args...)

	buffer := bytes.Buffer{}
	buffer.WriteString(encrypted)
	command.Stdin = &buffer

	output, err := command.Output()
	if err != nil {
		if e, ok := err.(*exec.ExitError); ok {
			log.Error(e, "failed to decrypt file", "file", fileName, "stderr", string(e.Stderr))
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
