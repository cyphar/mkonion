## `mkonion` ##

> **NOTE**: This doesn't actually exist at the moment, I'm working on the code
> while you're reading this. From my guess, this shouldn't be *too* hard to
> implement, but certain cool features might take a while and will probably show
> up later. Feel free to open an issue for a feature request, or open a PR to
> add some cool feature that you needed. This project is free software, so you're
> free to do what you like with it.

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

Will be updated once the code has been written.

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

### `TODO` ###

* [ ] Base.
* [ ] Generate configuration from `docker inspect`.
* [ ] Come up with a better method than `--net=container` or `--net=host`.
* [ ] Consider adding support for OnionCat for exposed UDP ports.
* [ ] Extend `mkonion` to work with services running on the host machine.

### License ###

TBD. Probably either MIT or Apache 2.0 depending on the size of the project.
