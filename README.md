# tinygo-w5500-driver

_Work in progress_

Wiznet [W5500](https://www.wiznet.io/product-item/w5500/) chip driver implementation for [TinyGO](https://tinygo.org/) stack.

Partially inspired Arduino's Ethernet implementation, mostly based on [official documentation](http://wizwiki.net/wiki/lib/exe/fetch.php/products:w5500:w5500_ds_v109e.pdf).

## Usage

TBD

## Examples

- [TCP HTTP client](examples/http_client/main.go)

## TODO

- [ ] organise package
- [ ] tests
- [ ] multiple sockets support
- [ ] better error handling
- [ ] UDP tested
- [ ] DHCP client
- [ ] DNS client
- [ ] prepare to be moved into tinygo-org/drivers repository
    - [ ] rework `net` package
    - [ ] refactor `DeviceDriver` in order to support multiple sockets