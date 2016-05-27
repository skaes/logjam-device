# Logjam Tools

A collection of programs and daemons to build the server side
infrastructure for logjam (see https://github.com/skaes/logjam_app).

<a href="https://scan.coverity.com/projects/3357">
  <img alt="Coverity Scan Build Status"
       src="https://scan.coverity.com/projects/3357/badge.svg"/>
</a>

Currently the following programs are provided:

## logjam-device

A daemon which offers a ZeroMQ PULL socket endpoint for applications
to connect to and a ZeroMQ SUB socket for forwarding. It optionally
subscribes to a RabbitMQ server to collect application messages from
there and republishes them on the PUB socket. It can compress the log
stream on the fly to reduce network traffic, although compressing it
at the producer is preferable. You can run as many of those devices as
needed to scale the logging infrastructure.

## logjam-importer

A multithreaded daemon using CZMQ's actor framework which has replaced
all of the ruby importer code in logjam. It's much less resource
intensive than the ruby code and a _lot_ faster.

## logjam-httpd

A daemon which takes frontend performance data via HTTP GET requests
and publishes it on on ZeroMQ PUB socket for the importer to pick up.

## logjam-graylog-forwarder

A daemon which subscribes to PUB sockets of logjam-devices and
forwards GELF messages to a graylog GELF socket endpoint.

## logjam-dump

A utility program to capture messages sent from a logjam device and
log them to disk.

## logjam-replay

A utility program to replay messages captured by logjam-dump. Useful
in determining maximum system throughput. Mimics a logjam-device.

## logjam-pubsub-bridge

A utility program which subscribes to a logjam-device PUB socket,
decompresses and forwards messages to a PUSH socket. Only to be used
when writing the decompression logic is to cmplex for a consumer.

## logjam-logger

A utility program which reads lines from stdin and publishes them on
a PUB socket, optionally using a given topic.

## logjam-tail

A utility program which connects to a logjam-logger PUB socket and
displays lines matching an optional list of topics on stdout.


# Dependencies

* librabbitmq (0.7.0)
* libzmq (4.1.4)
* libczmq (3.0.2)
* mongo-c-driver (1.1.2)
* libbson (included in mongo-c-driver as a submodule)
* json-c (0.12 patched)
* libsnappy (1.1.3)

# Installation

## Ubuntu packages

Ubuntu packages are available from
<a href="https://packagecloud.io/stkaes/logjam">
<img src="doc/packagecloud-logo-med-dark.png" height="16">
</a>.

The are two types of packages available: `logjam-tools` will install
in `/opt/logjam/`, `logjam-tools-usr-local` in `/usr/local`. The tools
package uses very recent and sometimes patched libraries. Installing
in `/usr/local` might cause problems with other applications. In this
case, use `-opt-logjam` packages. However, you will need to set some
enviroment variables to make use of the libararies provided by logjam:
Set `ZMQ_LIB_PATH` to `/opt/logjam/lib` in order to use `libzmq`from
`ffi-rzmq`, add `/opt/logjam/lib/pkgconfig` to `PKG_CONFIG_PATH` and
`/opt/logjam/bin` to `PATH`.

Installation instructions how to add the package cloud apt repository
can be found [here](https://packagecloud.io/stkaes/logjam/install).

The final step is then `apt-get install logjam-tools`.

Currently, 12.04 LTS and 14.04 LTS are supported.

## From source

Start by cloning the repository:
```
git clone git://github.com/skaes/logjam-tools.git
cd logjam-tools
```

Then, run `./bin/install-libs` to install all dependecies in `/usr/local`.

If you want to install everything into a separate hierarchy, you can
use the `--prefix` argument like so:

```
./bin/install-libs --prefix=/opt/logjam
```

Or install them manually:
* Download and install rabbitmq-c 0.7.0 from https://github.com/alanxz/rabbitmq-c/releases/tag/v0.7.0
* Download and install zmq 4.1.4 from http://zeromq.org/intro:get-the-software
* Dowmload and install czmq 3.0.2 from http://czmq.zeromq.org/page:get-the-software
* Clone https://github.com/skaes/json-c.git, checkout
  36be1c4c7ade78fae8ef67280cd4f98ff9f81016, build and install
* Clone https://github.com/mongodb/mongo-c-driver, checkout
  048c851919dc32c98b5f8a9a0270e69af9385db3, build and install
* Clone https://github.com/skaes/snappy.git, checkout
  10c7088336f490e646de7d40e9ace0958b269047, build and install

Finally
```
sh autogen.sh
make
sudo make install
```

The generated `./configure` script will try to use `pkg-config` to find the
required libraries. If `pkg-config` is not installed, it assumes the
headers and libraries are installed under `/opt/logjam`, `/usr/local` or
`/opt/local`. If they're somewhere else, you can specify
`--with-opt-dir=dir1:dir2:dir3` as argument to `sh autogen.sh` (or
`./configure`).

`autogen.sh` accepts the usual configure arguments, such as
`--prefix`. Thus, if you have installed the libraries under
`/opt/logjam`, and want to install the logjam tools in the same place,
run `sh autogen.sh --prefix=/opt/logjam`

If you want to get rid of the installed software, run
```
sudo make uninstall
./bin/install-libs uninstall
```

# Profiling with gperftools

Install Google perftools on your machine (https://code.google.com/p/gperftools/).

Set environment variable CPUPROFILE to the name of the profile data
file you want to use. Reconfigure and recompile everything:

```
CPUPROFILE=logjam.prof sh autogen.sh
make clean
make
```

Then invoke the command you want to profile. For example:

```
CPUPROFILE=logjam.prof ./logjam-device -c logjam.conf
pprof --web ./logjam-device logjam.prof
```

On Ubuntu, you will likely need to add `LD_PRELOAD=<path to libprofile.so>`
to make this work.

# License

GPL v3. See LICENSE.txt.
