# Capabilities Isolators Guide

This document is a walk-through guide describing how to use rkt isolators for
[Linux Capabilities][capabilities].

* [About Linux Capabilities](#about-linux-capabilities)
* [Default Capabilities](#default-capabilities)
* [Capability Isolators](#capability-isolators)
* [Configure capabilities via the command line](#configure-capabilities-via-the-command-line)
* [Configure capabilities in ACI images](#configure-capabilities-in-aci-images)
* [Capabilities when running as non-root](#capabilities-when-running-as-non-root)
* [Recommendations](#recommendations)

## About Linux Capabilities

Linux capabilities are meant to be a modern evolution of traditional UNIX
permissions checks.
The goal is to split the permissions granted to privileged processes into a set
of capabilities (eg. `CAP_NET_RAW` to open a raw socket), which can be
separately handled and assigned to single threads.

Processes can gain specific capabilities by either being run by superuser, or by
having the setuid/setgid bits or specific file-capabilities set on their
executable file.
Once running, each process has a bounding set of capabilities which it can
enable and use; such process cannot get further capabilities outside of this set.

In the context of containers, capabilities are useful for:

* Restricting the effective privileges of applications running as root
* Allowing applications to perform specific privileged operations, without
   having to run them as root

For the complete list of existing Linux capabilities and a detailed description
of this security mechanism, see the [capabilities(7) man page][man-capabilities].

## Default capabilities

By default, rkt enforces [a default set of capabilities][default-caps] onto applications.
This default set is tailored to stop applications from performing a large
variety of privileged actions, while not impacting their normal behavior.
Operations which are typically not needed in containers and which may
impact host state, eg. invoking `reboot(2)`, are denied in this way.

However, this default set is mostly meant as a safety precaution against erratic
and misbehaving applications, and will not suffice against tailored attacks.
As such, it is recommended to fine-tune the capabilities bounding set using one
of the customizable isolators available in rkt.

## Capability Isolators

When running Linux containers, rkt provides two mutually exclusive isolators
to define the bounding set under which an application will be run:

* `os/linux/capabilities-retain-set`
* `os/linux/capabilities-remove-set`

Those isolators cover different use-cases and employ different techniques to
achieve the same goal of limiting available capabilities. As such, they cannot
be used together at the same time, and recommended usage varies on a
case-by-case basis.

As the granularity of capabilities varies for specific permission cases, a word
of warning is needed in order to avoid a false sense of security.
In many cases it is possible to abuse granted capabilities in order to
completely subvert the sandbox: for example, `CAP_SYS_PTRACE` allows to access
stage1 environment and `CAP_SYS_ADMIN` grants a broad range of privileges,
effectively equivalent to root.
Many other ways to maliciously transition across capabilities have already been
[reported][grsec-forums].

### Retain-set

`os/linux/capabilities-retain-set` allows for an additive approach to
capabilities: applications will be stripped of all capabilities, except the ones
listed in this isolator.

This whitelisting approach is useful for completely locking down environments
and whenever application requirements (in terms of capabilities) are
well-defined in advance. It allows one to ensure that exactly and only the
specified capabilities could ever be used.

For example, an application that will only need to bind to port 80 as
a privileged operation, will have `CAP_NET_BIND_SERVICE` as the only entry in
its "retain-set".

### Remove-set

`os/linux/capabilities-remove-set` tackles capabilities in a subtractive way:
starting from the default set of capabilities, single entries can be further
forbidden in order to prevent specific actions.

This blacklisting approach is useful to somehow limit applications which have
broad requirements in terms of privileged operations, in order to deny some
potentially malicious operations.

For example, an application that will need to perform multiple privileged
operations but is known to never open a raw socket, will have
`CAP_NET_RAW` specified in its "remove-set".

## Configure capabilities via the command line

Capabilities can be directly overridden at run time from the command-line,
without changing the executed images.
The `--caps-retain` option to `rkt run` manipulates the `retain` capabilities set.
The `--caps-remove` option manipulates the `remove` set.

Capabilities specified from the command-line will replace all capability settings in the image manifest.
Also as stated above the options `--caps-retain`, and `--caps-remove` are mutually exclusive.
Only one can be specified at a time.

Capabilities isolators can be added on the command line at run time by
specifying the desired overriding set, as shown in this example:

```
$ sudo rkt run --interactive quay.io/coreos/alpine-sh --caps-retain CAP_NET_BIND_SERVICE
image: using image from file /usr/local/bin/stage1-coreos.aci
image: using image from local store for image name quay.io/coreos/alpine-sh

/ # whoami
root

/ # ping -c 1 8.8.8.8
PING 8.8.8.8 (8.8.8.8): 56 data bytes
ping: permission denied (are you root?)

```

Capability sets are application-specific configuration entries, and in a
`rkt run` command line, they must follow the application container image to
which they apply.
Each application within a pod can have different capability sets.

## Configure capabilities in ACI images

Capability sets are typically defined when creating images, as they are tightly
linked to specific app requirements.

The goal of these examples is to show how to build ACIs with [`acbuild`][acbuild],
where some capabilities are either explicitly blocked or allowed.
For simplicity, the starting point will be the official Alpine Linux image from
CoreOS which ships with `ping` and `nc` commands (from busybox). Those
commands respectively requires `CAP_NET_RAW` and `CAP_NET_BIND_SERVICE`
capabilities in order to perform privileged operations.
To block their usage, capabilities bounding set
can be manipulated via `os/linux/capabilities-remove-set` or
`os/linux/capabilities-retain-set`; both approaches are shown here.

### Removing specific capabilities

This example shows how to block `ping` only, by removing `CAP_NET_RAW` from
capabilities bounding set.

First, a local image is built with an explicit "remove-set" isolator.
This set contains the capabilities that need to be forbidden in order to block
`ping` usage (and only that):

```
$ acbuild begin
$ acbuild set-name localhost/caps-remove-set-example
$ acbuild dependency add quay.io/coreos/alpine-sh
$ acbuild set-exec -- /bin/sh
$ echo '{ "set": ["CAP_NET_RAW"] }' | acbuild isolator add "os/linux/capabilities-remove-set" -
$ acbuild write caps-remove-set-example.aci
$ acbuild end
```

Once properly built, this image can be run in order to check that `ping` usage has
been effectively disabled:

```
$ sudo rkt run --interactive --insecure-options=image caps-remove-set-example.aci
image: using image from file stage1-coreos.aci
image: using image from file caps-remove-set-example.aci
image: using image from local store for image name quay.io/coreos/alpine-sh

/ # whoami
root

/ # ping -c 1 8.8.8.8
PING 8.8.8.8 (8.8.8.8): 56 data bytes
ping: permission denied (are you root?)
```

This means that `CAP_NET_RAW` had been effectively disabled inside the container.
At the same time, `CAP_NET_BIND_SERVICE` is still available in the default bounding
set, so the `nc` command will be able to bind to port 80:

```
$ sudo rkt run --interactive --insecure-options=image caps-remove-set-example.aci
image: using image from file stage1-coreos.aci
image: using image from file caps-remove-set-example.aci
image: using image from local store for image name quay.io/coreos/alpine-sh

/ # whoami
root

/ # nc -v -l -p 80
listening on [::]:80 ...
```

### Allowing specific capabilities

In contrast to the example above, this one shows how to allow `ping` only, by
removing all capabilities except `CAP_NET_RAW` from the bounding set.
This means that all other privileged operations, including binding to port 80
will be blocked.

First, a local image is built with an explicit "retain-set" isolator.
This set contains the capabilities that need to be enabled in order to allowed
`ping` usage (and only that):

```
$ acbuild begin
$ acbuild set-name localhost/caps-retain-set-example
$ acbuild dependency add quay.io/coreos/alpine-sh
$ acbuild set-exec -- /bin/sh
$ echo '{ "set": ["CAP_NET_RAW"] }' | acbuild isolator add "os/linux/capabilities-retain-set" -
$ acbuild write caps-retain-set-example.aci
$ acbuild end
```

Once run, it can be easily verified that `ping` from inside the container is now
functional:

```
$ sudo rkt run --interactive --insecure-options=image caps-retain-set-example.aci
image: using image from file stage1-coreos.aci
image: using image from file caps-retain-set-example.aci
image: using image from local store for image name quay.io/coreos/alpine-sh

/ # whoami
root

/ # ping -c 1 8.8.8.8
PING 8.8.8.8 (8.8.8.8): 56 data bytes
64 bytes from 8.8.8.8: seq=0 ttl=41 time=24.910 ms

--- 8.8.8.8 ping statistics ---
1 packets transmitted, 1 packets received, 0% packet loss
round-trip min/avg/max = 24.910/24.910/24.910 ms
```

However, all others capabilities are now not anymore available to the application.
For example, using `nc` to bind to port 80 will now result in a failure due to
the missing `CAP_NET_BIND_SERVICE` capability:

```
$ sudo rkt run --interactive --insecure-options=image caps-retain-set-example.aci
image: using image from file stage1-coreos.aci
image: using image from file caps-retain-set-example.aci
image: using image from local store for image name quay.io/coreos/alpine-sh

/ # whoami
root

/ # nc -v -l -p 80
nc: bind: Permission denied
```

### Patching images

Image manifests can be manipulated manually, by unpacking the image and editing
the manifest file, or with helper tools like [`actool`][actool].
To override an image's pre-defined capabilities set, replace the existing capabilities
isolators in the image with new isolators defining the desired capabilities.

The `patch-manifest` subcommand to `actool` manipulates the capabilities sets
defined in an image.
`actool patch-manifest --capability` changes the `retain` capabilities set.
`actool patch-manifest --revoke-capability` changes the `remove` set.
These commands take an input image, modify its existing capabilities sets, and
write the changes to an output image, as shown in the example:

```
$ actool cat-manifest caps-retain-set-example.aci
...
    "isolators": [
      {
        "name": "os/linux/capabilities-retain-set",
        "value": {
          "set": [
            "CAP_NET_RAW"
          ]
        }
      }
    ]
...

$ actool patch-manifest -capability CAP_NET_RAW,CAP_NET_BIND_SERVICE caps-retain-set-example.aci caps-retain-set-patched.aci

$ actool cat-manifest caps-retain-set-patched.aci
...
    "isolators": [
      {
        "name": "os/linux/capabilities-retain-set",
        "value": {
          "set": [
            "CAP_NET_RAW",
            "CAP_NET_BIND_SERVICE"
          ]
        }
      }
    ]
...

```

Now run the image to check that the `CAP_NET_BIND_SERVICE` capability added to
the patched image is retained as expected by using `nc` to listen on a
"privileged" port:

```
$ sudo rkt run --interactive --insecure-options=image caps-retain-set-patched.aci
image: using image from file stage1-coreos.aci
image: using image from file caps-retain-set-patched.aci
image: using image from local store for image name quay.io/coreos/alpine-sh

/ # nc -v -l -p 80
listening on [::]:80 ...
```

## Capabilities when running as non-root

The capability isolators (and default capabilities) mentioned in this document operate on the capability bounding set.
When running containers as non-root, capabilities are not added to the effective set of the process, which is the one the kernel will check when the app is attempting to perform a privileged operation.
This means the process won't be able to run the privileged operations enabled by the capabilities granted to the container directly.

For example, including `CAP_NET_RAW` in the retain set when running a container as non-root doesn't allow the container to run `ping` (which uses raw sockets):

```
$ sudo rkt run --interactive kinvolk.io/aci/busybox --user=1000 --group=1000 --caps-retain=CAP_NET_RAW --exec ping -- 8.8.8.8
PING 8.8.8.8 (8.8.8.8): 56 data bytes
ping: permission denied (are you root?)
```

To be able to execute `ping` as a non-root user, the binary needs to have the corresponding file capability.

**Note**: running an image with file capabilities currently requires disabling seccomp in rkt.
This is due to a systemd bug where using seccomp results in enabling [no_new_privs][NNP] (you can track progress in [#3896](https://github.com/rkt/rkt/issues/3896)).

### Building images with file capabilities

Building images that include files with file capabilities is challenging since:

* [build][acbuild] doesn't preserve file capabilities. See [containers/build#197](https://github.com/containers/build/issues/197).
* docker build doesn't preserve file capabilities. See [moby/moby#35699](https://github.com/moby/moby/issues/35699).

However, provided we have an ACI (created for example with [build][acbuild] or with [docker2aci][docker2aci]), we can extract it and add the file capabilities manually.

#### Example

We'll use build to create an Ubuntu ACI with ping installed:

```bash
#!/usr/bin/env bash

acbuild --debug begin docker://ubuntu
acbuild --debug set-name example.com/filecap

# Install ping
acbuild --debug run -- apt-get update
acbuild --debug run -- apt-get install -y inetutils-ping

# ping comes as a setuid file, we don't want that
acbuild --debug run -- chmod -s /bin/ping

acbuild --debug set-exec /bin/bash

acbuild --debug write --overwrite filecaps.aci
```

After running the script, we'll extract the ACI, add file capabilities, and rebuild it.

```
$ sudo tar -xf filecaps.aci
$ sudo setcap cap_net_raw+ep rootfs/bin/ping
$ sudo tar --xattrs -cf filecaps-mod.aci manifest rootfs
```

Now we can run rkt with that image as a non-root user and with the right capability and ping should work fine:

```
$ sudo rkt --insecure-options=image,seccomp run --interactive filecaps-mod.aci --user=1000 --group=1000 --caps-retain=cap_net_raw
groups: cannot find name for group ID 1000
bash: /root/.bashrc: Permission denied
I have no name!@rkt-4a9e66a0-6ee0-496f-8b7a-ab259362cba7:/$ ls -l /bin/ping
-rwxr-xr-x 1 root root 70680 Feb  6  2016 /bin/ping
I have no name!@rkt-4a9e66a0-6ee0-496f-8b7a-ab259362cba7:/$ getcap /bin/ping
/bin/ping = cap_net_raw+ep
I have no name!@rkt-4a9e66a0-6ee0-496f-8b7a-ab259362cba7:/$ ping 8.8.8.8
PING 8.8.8.8 (8.8.8.8): 56 data bytes
64 bytes from 8.8.8.8: icmp_seq=0 ttl=52 time=21.859 ms
64 bytes from 8.8.8.8: icmp_seq=1 ttl=52 time=20.298 ms
64 bytes from 8.8.8.8: icmp_seq=2 ttl=52 time=25.207 ms
^C--- 8.8.8.8 ping statistics ---
3 packets transmitted, 3 packets received, 0% packet loss
round-trip min/avg/max/stddev = 20.298/22.455/25.207/2.048 ms
```

### Ambient capabilities

There's a way in Linux to give capabilities to non-root processes without needing file capabilities to perform the privileged task: ambient capabilities.

When a capability is added to the ambient set, it will be preserved in the effective set of the process executed in the container.
However, this is currently not implemented in rkt.

For more information about ambient capabilities, check [capabilities(7)][man-capabilities].

## Recommendations

As with most security features, capability isolators may require some
application-specific tuning in order to be maximally effective. For this reason,
for security-sensitive environments it is recommended to have a well-specified
set of capabilities requirements and follow best practices:

 1. Always follow the principle of least privilege and, whenever possible,
    avoid running applications as root
 2. Only grant the minimum set of capabilities needed by an application,
    according to its typical usage
 3. Avoid granting overly generic capabilities. For example, `CAP_SYS_ADMIN` and
    `CAP_SYS_PTRACE` are typically bad choices, as they open large attack
    surfaces.
 4. Prefer a whitelisting approach, trying to keep the "retain-set" as small as
    possible.

[acbuild]: https://github.com/containers/build
[docker2aci]: https://github.com/appc/docker2aci
[actool]: https://github.com/appc/spec#building-acis
[capabilities]: https://lwn.net/Kernel/Index/#Capabilities
[default-caps]: https://github.com/appc/spec/blob/master/spec/ace.md#oslinuxcapabilities-remove-set
[grsec-forums]: https://forums.grsecurity.net/viewtopic.php?f=7&t=2522
[man-capabilities]: http://man7.org/linux/man-pages/man7/capabilities.7.html
[NNP]: https://www.kernel.org/doc/Documentation/prctl/no_new_privs.txt
