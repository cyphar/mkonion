## `TODO` ##

### Requirements ###

* [x] Base.
* [x] Generate configuration from `docker inspect`.
* [x] Come up with a better method than `--net=container` or `--net=host`.

### Features ###

* [ ] Use the [tor network plugin][tor-network] to automatically route all of the
      target container's traffic through Tor. This might cause some issues, so it'd
      have to be optional. We'd need to disconnect the container from the default
      `bridge` network.
* [ ] Allow users to specify a volume for the Tor hidden service directory.
* [ ] Consider adding support for OnionCat for exposed UDP ports.

[tor-network]: https://github.com/jfrazelle/onion
