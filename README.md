# SOPS Operator

[![License](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](https://opensource.org/licenses/Apache-2.0)

A Kubernetes operator for [Mozilla SOPS](https://github.com/mozilla/sops).

## Overview

Put SOPS-encrypted data into a `SopsSecret` which can then be committed to a Git repository.
Once deployed on a Kubernetes cluster, the SOPS Operator will decrypt the data and create a standard Kubernetes `Secret` from it.

*Example for a SopsSecret:*

```yaml
apiVersion: craftypath.github.io/v1alpha1
kind: SopsSecret
metadata:
  name: test-secret
spec:
  stringData:
    test.yaml: |
      test: ENC[AES256_GCM,data:xo8jZTsQ,iv:DTouw1kgBLok6BbR5vx8366fFavV70QeCWGNQPhNb9s=,tag:RAjeoNhvGUezdOS4YOorfA==,type:str]
      sops:
          kms: []
          gcp_kms: []
          azure_kv:
          -   vault_url: https://myakskeyvault12345567.vault.azure.net
              name: sops
              version: 08faa451b1d04b8bacec0395fc8539f1
              created_at: '2020-05-01T19:42:49Z'
              enc: DvZNm3tfyoyWibQcVPts9ODRPs3aaHbRaXOPIx1Ukypa2nPmU4RCTchBPUoqscIxDjKpSy9k6A_dfE8XAu8-XrEyuOGCEy-i6Q1OtZSGW1XnWfWXPic5TF7XCVz_08h1My1RzVUr51PPNX9uazCqQeUTfBx05KC1bT3entgfttHp-98uZkZNaI8IUUnPGCH8bZzthsXRSvRQpbZcNoOW3y04pLAVYN3xVSOdDWQSElmntg_t7eVdCsmj4iXrC-J80VPU6BoZetcsQhOLjAhXHEYMOP7fqjd2bXob59Ad8rblUDwwtcZrku5lF_LVvAKGBURxockQXmEuVAjqha1SyA
          lastmodified: '2020-05-01T19:42:50Z'
          mac: ENC[AES256_GCM,data:L4YfHJ59L+/YFMTizeSmEz3QiFbNYoRBVeAJNbHOCUU0W7Iv/WfGnZuNnG5c3gOELYafc812CxCFHYwoLK0bLxOd+KHwGp5IBZ7zqrg91e04V/7Tc3iEYCE3YuTQZ56XMeSSKsct7HT7jxzmVMjW0ozJ06vzQCEC/Ljsl2NfFNs=,iv:RiBXtk6Gpc/MZvDRaGKlvA8A0K7E7bGdhs8tVa6LL5w=,tag:hwnh954tiRC/VBp6LQ6nPg==,type:str]
          pgp: []
          unencrypted_suffix: _unencrypted
          version: 3.5.0
```

*Here's the Secret that's created from it:*

```yaml
apiVersion: v1
kind: Secret
metadata:
  name: test-secret
data:
  test.yaml: dGVzdDogdGVzdHZhbHVlCg==
```

## Installation

A Helm chart is available in our charts repo at https://github.com/craftypath/helm-charts.

```console
helm repo add craftypath https://craftypath.github.io/helm-charts
helm install craftypath/sops-operator
```

Check out the chart's documentation for configuration options.
