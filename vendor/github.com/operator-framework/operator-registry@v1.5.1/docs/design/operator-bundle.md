# Operator Bundle

An `Operator Bundle` is a container image that stores Kubernetes manifests and metadata associated with an operator. A bundle is meant to present a *specific* version of an operator.

## Operator Bundle Overview

The operator manifests refers to a set of Kubernetes manifest(s) the defines the deployment and RBAC model of the operator. The operator metadata on the other hand are, but not limited to:
* Information that identifies the operator, its name, version etc.
* Additional information that drives the UI:
    * Icon
    * Example CR(s)
* Channel(s)
* API(s) provided and required.
* Related images.

An `Operator Bundle` is built as a scratch (non-runnable) container image that contains operator manifests and specific metadata in designated directories inside the image. Then, it can be pushed and pulled from an OCI-compliant container registry. Ultimately, an operator bundle will be used by [Operator Registry](https://github.com/operator-framework/operator-registry) and [Operator-Lifecycle-Manager (OLM)](https://github.com/operator-framework/operator-lifecycle-manager) to install an operator in OLM-enabled clusters.

### Bundle Annotations

We use the following labels to annotate the operator bundle image.
* The label `operators.operatorframework.io.bundle.mediatype.v1` reflects the media type or format of the operator bundle. It could be helm charts, plain Kubernetes manifests etc.
* The label `operators.operatorframework.io.bundle.manifests.v1 `reflects the path in the image to the directory that contains the operator manifests.
* The label `operators.operatorframework.io.bundle.metadata.v1` reflects the path in the image to the directory that contains metadata files about the bundle.
* The `manifests.v1` and `metadata.v1` labels imply the bundle type:
    * The value `manifests.v1` implies that this bundle contains operator manifests.
    * The value `metadata.v1` implies that this bundle has operator metadata.
* The label `operators.operatorframework.io.bundle.package.v1` reflects the package name of the bundle.
* The label `operators.operatorframework.io.bundle.channels.v1` reflects the list of channels the bundle is subscribing to when added into an operator registry
* The label `operators.operatorframework.io.bundle.channel.default.v1` reflects the default channel an operator should be subscribed to when installed from a registry

The labels will also be put inside a YAML file, as shown below.

*annotations.yaml*
```yaml
annotations:
  operators.operatorframework.io.bundle.mediatype.v1: "registry+v1"
  operators.operatorframework.io.bundle.manifests.v1: "manifests/"
  operators.operatorframework.io.bundle.metadata.v1: "metadata/"
  operators.operatorframework.io.bundle.package.v1: "test-operator"
  operators.operatorframework.io.bundle.channels.v1: "beta,stable"
  operators.operatorframework.io.bundle.channel.default.v1: "stable"
```

*Notes:*
* In case of a mismatch, the `annotations.yaml` file is authoritative because the on-cluster operator-registry that relies on these annotations has access to the yaml file only.
* The potential use case for the `LABELS` is - an external off-cluster tool can inspect the image to check the type of a given bundle image without downloading the content.

This example uses [Operator Registry Manifests](https://github.com/operator-framework/operator-registry#manifest-format) format to build an operator bundle image. The source directory of an operator registry bundle has the following layout.
```
$ tree test
test
├── testbackup.crd.yaml
├── testcluster.crd.yaml
├── testoperator.v0.1.0.clusterserviceversion.yaml
├── testrestore.crd.yaml
└── metadata
    └── annotations.yaml
```

### Bundle Dockerfile

This is an example of a `Dockerfile` for operator bundle:
```
FROM scratch

# We are pushing an operator-registry bundle
# that has both metadata and manifests.
LABEL operators.operatorframework.io.bundle.mediatype.v1=registry+v1
LABEL operators.operatorframework.io.bundle.manifests.v1=manifests/
LABEL operators.operatorframework.io.bundle.metadata.v1=metadata/
LABEL operators.operatorframework.io.bundle.package.v1=test-operator
LABEL operators.operatorframework.io.bundle.channels.v1=beta,stable
LABEL operators.operatorframework.io.bundle.channel.default.v1=stable

ADD test/*.yaml /manifests
ADD test/metadata/annotations.yaml /metadata/annotations.yaml
```

Below is the directory layout of the operator bundle inside the image:
```bash
$ tree
/
├── manifests
│   ├── testbackup.crd.yaml
│   ├── testcluster.crd.yaml
│   ├── testoperator.v0.1.0.clusterserviceversion.yaml
│   └── testrestore.crd.yaml
└── metadata
    └── annotations.yaml
```

## Operator Bundle Commands

Operator SDK CLI is available to generate Bundle annotations and Dockerfile based on provided operator manifests.

### Operator SDK CLI

In order to use Operator SDK CLI, follow the operator-SDK installation instruction:

1. Install the [Operator SDK CLI](https://github.com/operator-framework/operator-sdk/blob/master/doc/user/install-operator-sdk.md)

Now, a binary named `operator-sdk` is available in OLM's directory to use.
```bash
$ ./operator-sdk
An SDK for building operators with ease

Usage:
  operator-sdk [command]

Available Commands:
    bundle      Operator bundle commands

Flags:
  -h, --help      help for operator-sdk
      --verbose   Enable verbose logging

Use "operator-sdk [command] --help" for more information about a command.
```

### Generate Bundle Annotations and DockerFile
*Notes:*
* If there are `annotations.yaml` and `Dockerfile` existing in the directory, they will be overwritten.

Using `operator-sdk` CLI, bundle annotations can be generated from provided operator manifests. The overall `bundle generate` command usage is:
```bash
Usage:
  operator-sdk bundle generate [flags]

Flags:
  -c, --channels string    The list of channels that bundle image belongs to
  -e, --default string     The default channel for the bundle image
  -d, --directory string   The directory where bundle manifests are located.
  -h, --help               help for generate
  -p, --package string     The name of the package that bundle image belongs to
```

The `--directory/-d`, `--channels/-c`, `--package/-p` are required flags while `--default/-e` is optional.

The command for `generate` task is:
```bash
$ ./operator-sdk bundle generate --directory /test --package test-operator \
--channels stable,beta --default stable
```

The `--directory` or `-d` specifies the directory where the operator manifests are located. The `Dockerfile` is generated in the same directory where the YAML manifests are located while the `annotations.yaml` file is located in a folder named `metadata`. For example:
```bash
$ tree test
test
├── testbackup.crd.yaml
├── testcluster.crd.yaml
├── testoperator.v0.1.0.clusterserviceversion.yaml
├── testrestore.crd.yaml
├── metadata
│   └── annotations.yaml
└── Dockerfile
```

The `--package` or `-p` is the name of package fo the operator such as `etcd` which which map `channels` to a particular application definition. `channels` allow package authors to write different upgrade paths for different users (e.g. `beta` vs. `stable`). The `channels` list is provided via `--channels` or `-c` flag. Multiple `channels` are separated by a comma (`,`). The default channel is provided optionally via `--default` or `-e` flag. If the default channel is not provided, the first channel in channel list is selected as default.

All information in `annotations.yaml` is also existed in `LABEL` section of `Dockerfile`.

### Build Bundle Image

Operator bundle image can be built from provided operator manifests using `build` command (see *Notes* below). The overall `bundle build` command usage is:
```bash
Usage:
  operator-SDK bundle build [flags]

Flags:
  -c, --channels string        The list of channels that bundle image belongs to
  -e, --default string         The default channel for the bundle image
  -d, --directory string       The directory where bundle manifests are located
  -h, --help                   help for build
  -b, --image-builder string   Tool to build container images. One of: [docker, podman, buildah] (default "docker")
  -0, --overwrite               To overwrite annotations.yaml if existing
  -p, --package string         The name of the package that bundle image belongs to
  -t, --tag string             The name of the bundle image will be built
```

The command for `build` task is:
```bash
$ ./operator-sdk bundle build --directory /test --tag quay.io/coreos/test-operator.v0.1.0:latest \
--package test-operator --channels stable,beta --default stable
```

The `--directory` or `-d` specifies the directory where the operator manifests are located. The `--tag` or `-t` specifies the image tag that you want the operator bundle image to have. By using `build` command, the `annotations.yaml` and `Dockerfile` are automatically generated in the background.

The default image builder is `Docker`. However, ` Buildah` and `Podman` are also supported. An image builder can specified via `--image-builder` or `-b` optional tag in `build` command. For example:
```bash
$ ./operator-sdk bundle build --directory /test/0.1.0/ --tag quay.io/coreos/test-operator.v0.1.0:latest \
--image-builder podman --package test-operator --channels stable,beta --default stable
```

The `--package` or `-p` is the name of package fo the operator such as `etcd` which which map `channels` to a particular application definition. `channels` allow package authors to write different upgrade paths for different users (e.g. `beta` vs. `stable`). The `channels` list is provided via `--channels` or `-c` flag. Multiple `channels` are separated by a comma (`,`). The default channel is provided optionally via `--default` or `-e` flag. If the default channel is not provided, the first channel in channel list is selected as default.

*Notes:*
* If there is `Dockerfile` existing in the directory, it will be overwritten.
* If there is an existing `annotations.yaml` in `/metadata` directory, the cli will attempt to validate it and returns any found errors. If the ``annotations.yaml`` is valid, it will be used as a part of build process. The optional boolean `--overwrite/-o` flag can be enabled (false by default) to allow cli to overwrite the `annotations.yaml` if existed.
* The directory where the operator manifests are located must be inside the context of the build which in this case is inside the directory where you run the command.
