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
	"go.mozilla.org/sops/v3/cmd/sops/formats"
	"go.mozilla.org/sops/v3/decrypt"
)

type Decryptor struct{}

// Decrypt decrypts the given encrypted string. The format (yaml, json, env, init, binary)
// is determined by the given fileName.
func (d *Decryptor) Decrypt(fileName string, encrypted string) ([]byte, error) {
	format := formats.FormatForPath(fileName)
	return decrypt.DataWithFormat([]byte(encrypted), format)
}
