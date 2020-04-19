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
