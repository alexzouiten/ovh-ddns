# Changelog

## 1.0.0 (2026-10-15)


### Features

* add code ([87771f6](https://github.com/alexzouiten/ovh-ddns/commit/87771f660870a1609dac5f7da97f33cb3f86ee16))
* add timeout on update ([860e4c2](https://github.com/alexzouiten/ovh-ddns/commit/860e4c24e51fd298e01199b3184114e3ad9b8743))
* make update interval configurable ([526c64d](https://github.com/alexzouiten/ovh-ddns/commit/526c64dd440293448eb6bd0e3d58b8c4f4dde0ff))
* periodic update ([e54d45d](https://github.com/alexzouiten/ovh-ddns/commit/e54d45d51d4086217686f75cf5a3f6de97745abc))
* validate log level configuration ([b98f0df](https://github.com/alexzouiten/ovh-ddns/commit/b98f0dfacdfcb6a0c8515f2817fa0208930d91e8))
* wait for in-flight update before exiting ([172489a](https://github.com/alexzouiten/ovh-ddns/commit/172489ab1f6697e82f4a8d77319bc4c917e383c9))


### Bug Fixes

* **config:** fix yaml parsing + end program when conf loading fails ([bf526cc](https://github.com/alexzouiten/ovh-ddns/commit/bf526cc1ef8bcd376835ad0469122eb8be4d2e24))
* fetch only records for empty subdomain ([68bfcd0](https://github.com/alexzouiten/ovh-ddns/commit/68bfcd0a4f3695e4566e3ce4611dfd562f2feca0))
* only tell that dry-run is active if it is ([a2d4900](https://github.com/alexzouiten/ovh-ddns/commit/a2d4900fe00fff734ada3e32287bf29e4eadc903))
* wrong check against IPs cache ([fc5b159](https://github.com/alexzouiten/ovh-ddns/commit/fc5b15919555155ca24567d2905a6b260489b1bc))
