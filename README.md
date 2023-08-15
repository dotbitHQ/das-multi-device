* [Prerequisites](#prerequisites)
* [Install &amp; Run](#install--run)
    * [Source Compile](#source-compile)
    * [Docker](#docker)
* [Usage](#usage)

# das-multi-device
Multi device service, providing functions such as adding and removing backup devices through webauhtn

## Prerequisites

* Ubuntu 18.04 or newer
* MYSQL >= 8.0
* go version >= 1.17.0
* [CKB Node](https://github.com/nervosnetwork/ckb)

## Install & Run

### Source Compile

```bash
# get the code
git clone git@github.com:dotbitHQ/das-multi-device.git

# edit config/config.yaml and run unipay_svr

# compile and run
cd das-multi-device
make device
./das_multi_device --config=config/config.yaml
```

### Docker
* docker >= 20.10
* docker-compose >= 2.2.2

```bash
sudo curl -L "https://github.com/docker/compose/releases/download/v2.2.2/docker-compose-$(uname -s)-$(uname -m)" -o /usr/local/bin/docker-compose
sudo chmod +x /usr/local/bin/docker-compose
sudo ln -s /usr/local/bin/docker-compose /usr/bin/docker-compose
docker-compose up -d
```

_if you already have a mysql installed, just run_
```bash
docker run -dv $PWD/config/config.yaml:/app/config/config.yaml --name unipay_svr dotbitteam/unipay:latest
```