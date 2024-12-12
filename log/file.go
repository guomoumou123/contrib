package log

var (
	defaultMaxSize    = 1 << 10 //默认 1G 轮转
	defaultMaxAge     = 29
	defaultMaxBackups = 30
	defaultFileName   = "log/default.log"
	defaultCompress   = true
)

func defaultConfig() *Config {
	return &Config{
		Debug:       true,
		MaxSize:     defaultMaxSize,    //日志文件最大多少兆
		MaxAge:      defaultMaxAge,     //日志文件保留天数
		MaxBackups:  defaultMaxBackups, //保留文件数
		FileName:    defaultFileName,   //日志名字
		Compress:    defaultCompress,   //日志生成压缩包,大幅降低磁盘空间,必要时使用
		RotateByDay: false,             //每天轮转一次,如果开启,maxBackups的值需要>=maxDays
	}
}
