module github.com/gojekfarm/xtools/xworker/gocraft

go 1.16

require (
	github.com/Bose/minisentinel v0.0.0-20200130220412-917c5a9223bb
	github.com/alicebob/miniredis/v2 v2.33.0
	github.com/gojek/work v0.7.7
	github.com/gojekfarm/xtools/xworker v0.58.0
	github.com/gomodule/redigo v1.8.9
	github.com/rs/zerolog v1.20.0
	github.com/sethvargo/go-retry v0.1.0
	github.com/stretchr/testify v1.7.0
)

replace github.com/gojekfarm/xtools/xworker => ../
