module SuperH

go 1.15
replace gorm.io/gorm => github.com/go-gorm/gorm v1.9.19

require (
	github.com/go-ini/ini v1.60.2
	github.com/gomodule/redigo v1.8.2
	github.com/jinzhu/now v1.1.1
	github.com/lestrrat-go/file-rotatelogs v2.4.0+incompatible
	github.com/lestrrat-go/strftime v1.0.3 // indirect
	github.com/robfig/cron/v3 v3.0.0
	github.com/sirupsen/logrus v1.6.0
	github.com/smartystreets/goconvey v1.6.4 // indirect
	github.com/spf13/viper v1.7.1
	golang.org/x/net v0.0.0-20200822124328-c89045814202 // indirect
	gopkg.in/ini.v1 v1.60.2 // indirect
	gorm.io/driver/mysql v1.0.0
	gorm.io/gorm v1.9.19
)
