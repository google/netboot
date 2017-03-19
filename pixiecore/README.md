# Pixiecore

Pixiecore is an all-in-one tool to manage network booting of
machines. It can be used either as a simple tool for ad-hoc network
boots, or as a building block of machine management infrastructure.

[![license](https://img.shields.io/github/license/google/netboot.svg?maxAge=2592000)](https://github.com/google/netboot/blob/master/LICENSE) [![Travis](https://img.shields.io/travis/google/netboot.svg?maxAge=2592000)](https://travis-ci.org/google/netboot)  ![api](https://img.shields.io/badge/api-unstable-red.svg) ![cli](https://img.shields.io/badge/cli-stable-green.svg) [![cli](https://img.shields.io/badge/godoc-reference-blue.svg)](https://godoc.org/go.universe.tf/netboot/pixiecore)

![Pixiecore UI](https://cdn.rawgit.com/google/netboot/master/pixiecore/pixiecore-ui.png)

## Why?

Booting a Linux system over the network is quite tedious. You have to
set up a TFTP server, reconfigure your DHCP server to recognize PXE
clients, and send them the right set of magical options to get them to
boot, often fighting rubbish PXE ROM implementations.

Pixiecore aims to simplify this process, by packing the whole process
into a single binary that can cooperate with your network's existing
DHCP server. You don't need to reconfigure anything else in the
network.

If you're curious about the whole process that Pixiecore manages, you
can read the details in [README.booting](README.booting.md).

## Installation

Pixiecore is available in a variety of forms. All of them
automatically track this repository, so you always get the latest
build.

### Go get

Build the latest Pixiecore via `go get`:

```shell
go get go.universe.tf/netboot/cmd/pixiecore
```

### Debian/Ubuntu

A Debian/Ubuntu package is available from
[packagecloud.io](https://packagecloud.io/danderson/pixiecore/install). They
have extensive configuration instructions for a variety of mechanisms,
but the quick version is:

```shell
sudo apt-get install -y apt-transport-https
curl -L https://packagecloud.io/danderson/pixiecore/gpgkey | sudo apt-key add -
echo "deb https://packagecloud.io/danderson/pixiecore/debian stretch main" >/etc/apt/sources.list.d/pixiecore.list
sudo apt-get update
sudo apt-get install pixiecore
```

Note that you should reference debian/stretch regardless of your
actual distro. The pixiecore binary is built statically and should
work fine on all distros, so we only build one variant of the
package. Please file a bug if you hit problems with this setup.

### Container images

Docker and ACI autobuilds are available. They track the latest code
from this repository.

 - Docker image on Docker Hub: [danderson/pixiecore](https://hub.docker.com/r/danderson/pixiecore/)
 - Rkt ACI image on Quay.io: [quay.io/pixiecore/pixiecore](https://quay.io/repository/pixiecore/pixiecore)

## Using Pixiecore in static mode ("I just want to boot a machine")

Run the pixiecore binary, passing it a kernel and initrd, and
optionally some extra kernel commandline arguments. For example,
here's how you make all machines in your network netboot into the
alpha release of CoreOS, with automatic login:

```shell
sudo pixiecore boot \
  https://alpha.release.core-os.net/amd64-usr/current/coreos_production_pxe.vmlinuz \
  https://alpha.release.core-os.net/amd64-usr/current/coreos_production_pxe_image.cpio.gz \
  --cmdline='coreos.autologin'
```

That's it! Any machine that tries to boot from the network will now
boot into CoreOS.

That's a bit slow to boot, because Pixiecore is refetching the images
from core-os.net each time a machine tries to boot. We can also
download the files and use those:

```shell
wget https://alpha.release.core-os.net/amd64-usr/current/coreos_production_pxe.vmlinuz
wget https://alpha.release.core-os.net/amd64-usr/current/coreos_production_pxe_image.cpio.gz
sudo pixiecore boot \
  coreos_production_pxe.vmlinuz \
  coreos_production_pxe_image.cpio.gz \
  --cmdline='coreos.autologin'
```

Sometimes, you want to give extra files to the booting OS. For
example, CoreOS lets you pass a Cloud Init file via the
`cloud-config-url` kernel commandline parameter. That's fine if you
have a URL, but what if you have a local file?

For this, Pixiecore lets you specify that you want an additional file
served over HTTP to the booting OS, via a template function. Let's
grab a [cloud-config.yml](https://goo.gl/7HzZf2) that sets the
hostname to `pixiecore-test`, and serve it:

```shell
wget -O my-cloud-config.yml https://goo.gl/7HzZf2
sudo pixiecore boot \
  coreos_production_pxe.vmlinuz \
  coreos_production_pxe_image.cpio.gz \
  --cmdline='coreos.autologin cloud-config-url={{ ID "./my-cloud-config.yml" }}'
```

Pixiecore will transform the template invocation into a URL that, when
fetched, serves `my-cloud-config.yml`. Similarly to the kernel and
initrd arguments, you can also pass a URL to the `ID` template
function.

## Pixiecore in API mode

Think of Pixiecore in API mode as a "PXE to HTTP" translator. Whenever
Pixiecore sees a machine trying to netboot, it will ask a remote HTTP
API (which you implement) what to do. The API server can tell
Pixiecore to ignore the machine, or tell it to boot into a given
kernel/initrd/commandline.

Effectively, Pixiecore in API mode lets you pretend that your machines
speak a simple JSON protocol when trying to netboot. This makes it
_far_ easier to play with netbooting in your own software.

To start Pixiecore in API mode, pass it the URL of your API endpoint:

```shell
sudo pixiecore api https://foo.example/pixiecore
```

The endpoint you provide must implement the Pixiecore boot API, as
described in the [API spec](README.api.md).

You can find a sample API server implementation in the `api-example`
subdirectory. The code is not production-grade, but gives a short
illustration of how the protocol works by reimplementing a subset of
Pixiecore's static mode as an API server.

## Running in containers

Pixiecore is available both as an ACI image for `rkt`, and as a Docker
image for Docker Engine. Both images are automatically built whenever
the github repository changes.

Because Pixiecore needs to listen for DHCP traffic, it has to run with
access to the host's networking stack. Both Rkt and Docker do this
with the `--net=host` commandline flag.

```shell
sudo rkt run --net=host \
  --volume images,kind=host,source=/var/images \
  --mount volume=images,target=/image \
  quay.io/pixiecore/pixiecore -- \
    boot /image/coreos_production_pxe.vmlinuz /image/coreos_production_pxe_image.cpio.gz

sudo docker run \
  --net=host \
  -v .:/image \
  danderson/pixiecore \
    boot /image/coreos_production_pxe.vmlinuz /image/coreos_production_pxe_image.cpio.gz
```

## Demos and users

Pixiecore was used alongside
[waitron](https://github.com/jhaals/waitron) in a
[presentation](https://youtu.be/QyGHZ2HCwqY?t=440) at the OpenStack
summit in 2016).

If you use Pixiecore, we'd love to hear about it, and know more about
how you're using it. You can open a pull request to be added to this
list, file an issue for me to add you, or just email me at
dave(at)natulte.net if you'd like to give feedback privately.

- [waitron](https://github.com/jhaals/waitron) uses Pixiecore to
  manage automated server installation based on machine templates.
