## `TODO` ##

### Requirements ###

* [x] Base.
* [x] Generate configuration from `docker inspect`.
* [x] Come up with a better method than `--net=container` or `--net=host`.

### Features ###

* [x] Allow users to specify port mappings for hidden services.
* [ ] Use the [tor network plugin][tor-network] to automatically route all of the
      target container's traffic through Tor. This might cause some issues, so it'd
      have to be optional. We'd need to disconnect the container from the default
      `bridge` network. How should we package the plugin dependency?
* [x] Allow users to specify a path to an existing private key to be injected
      into the `hidden_service` directory.
* [ ] Consider adding support for OnionCat for exposed UDP ports.

[tor-network]: https://github.com/jfrazelle/onion

### Bugs ###

* [ ] If the Tor container dies before we get the hostname, the waiting loop
      doesn't terminate. We need to check for container death each time we wait
      when getting the hostname.
