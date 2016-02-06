## `mkonion` ##

`mkonion` is a very simple tool to allow you to set up a Tor Onion Service (also
known as a Tor Hidden Service) for an existing Docker container. It takes
advantage of `docker inspect` and other such features to figure out what ports
you might want to add to your hidden service. I have plans to allow you to create
multiple hidden services for a single container (I'm not sure why, but someone
is bound to need that).

The point of this project is for this to be as turn-key as possible, in order to
make it as easy as possible for people to try out Tor Onion Services with their
existing Docker setup.

### Usage ###

The basic usage is the following:

```
% mkonion [-k private_key] [-p [onion:]container]... <container>
```

Simple as that. You don't need to have any Tor setup, as `mkonion` includes
inside it all of the required `Dockerfile` and configuration information to set
up a new Tor container. If you want to take a closer look, check out `fakebuild.go`.

### Requirements ###

`mkonion` depends on first-class networking in the Docker daemon, which means
your Docker daemon must be at least version `1.9.0`. Any earlier versions could
be made to work with some hacks, but without first-class networking it can't work
reliably and easily.

### Recommendations ###

It's recommended to route all of your main container's traffic exclusively
through Tor (using the [Tor networking driver][tor-network]), so if your service
gets hacked the attacker cannot effectively retrieve your external IP address.

If you're planning to use this on a service which requires server anonymity as a
constraint, ensure that you remove all uniquely identifying information. Running
your service in Tor masks your IP address in some senses, if you route all of the
traffic through Tor. But you should also take steps to configure your server to
not leak information, as well as reducing how easily [your writing can be fingerprinted][anonymouth].

[tor-network]: https://github.com/jfrazelle/onion
[anonymouth]: https://github.com/psal/anonymouth

### Overview ###

Basically, `mkonion` automates the following steps:

1. Create a new `bridge` network, connect the target container to the network.
2. Generate a `torrc` which defines a Tor Onion Service that forwards all of the
   exposed ports of the target container to the new network's IP for the target.
3. Start a new Tor daemon in middle relay mode in a container connected to the
   new network.

You could in principle emulate `mkonion` with something like:

```
% docker network create --driver=bridge mkonion
% docker network connect mkonion <target>
% # Manually create a new torrc based on this:
% docker inspect <target>
% docker build -t <tor image> .
% docker run --net=mkonion <tor image>
```

But who wants all of that typing?

### Why Onion Services? ###

There are multiple reasons, the first and foremost being that it is (from my
understanding of the underlying technology) much more privacy-preserving for your
end users. If a user wants to access your website using Tor and you don't have a
problem with that, you should provide an onion address because it will protect
them from certain passive surveillance attacks (as well as misconfigurations that
cause them to accidentally connect directly to your service).

However, there are other cool reasons to use Tor Onion Services:

* NAT Punching means that you can connect to a Tor Onion Service even if it is
  hosted behind a NAT. This is an incredibly useful feature (you can access your
  internal services from the internet with much fewer problems than port-forwarding
  what you need).

* Your host can be hidden, as the current incarnation of Tor Onion Service sets
  up a full Tor circuit for all three introduction points, the HSDir nodes and
  the rendezvous node. This results in both the client and server having quite
  good anonymity properties (although the server's anonymity properties are known
  to be weaker than the client, because the server can be coerced to send data
  through its Tor circuits).

* By making more services available through `.onion` addresses, Tor Onion Services
  become more normalised and people are far less suspicious of such addresses.
  This is critical for the movement for online privacy, making people aware of
  the benefits of Tor and of `.onion` addressing.

Most importantly, because this project is so simple to use and is self-contained,
you lose nothing by starting this up on your services. At the very least, I hope
you'll try this out on your local machine so you can access your local dockerised
services from the internet using Tor.

### License ###

This project is licensed under the MIT/X11 License, as it is a fairly small
project there's no reason to use the Apache 2.0 or GNU GPL licenses.

```
mkonion: create a Tor onion service for existing Docker containers
Copyright (C) 2016 Aleksa Sarai <cyphar@cyphar.com>

Permission is hereby granted, free of charge, to any person obtaining a copy of
this software and associated documentation files (the "Software"), to deal in
the Software without restriction, including without limitation the rights to
use, copy, modify, merge, publish, distribute, sublicense, and/or sell copies of
the Software, and to permit persons to whom the Software is furnished to do so,
subject to the following conditions:

1. The above copyright notice and this permission notice shall be included in
   all copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
SOFTWARE.
```
