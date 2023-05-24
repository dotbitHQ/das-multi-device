module das-multi-device

go 1.16

require (
	github.com/dotbitHQ/das-lib v1.0.2-0.20230523130730-f7007e4ab6af
	github.com/fsnotify/fsnotify v1.5.4
	github.com/gin-gonic/gin v1.7.2
	github.com/go-redis/redis v6.15.9+incompatible
	github.com/nervosnetwork/ckb-sdk-go v0.101.3
	github.com/parnurzeal/gorequest v0.2.16
	github.com/scorpiotzh/mylog v1.0.10
	github.com/scorpiotzh/toolib v1.1.5
	github.com/urfave/cli/v2 v2.3.0
	gorm.io/gorm v1.23.6
)

replace github.com/ethereum/go-ethereum v1.9.14 => github.com/ethereum/go-ethereum v1.10.17
