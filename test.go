package main

import log "github.com/sirupsen/logrus"

func main() {
	log.SetFormatter(&log.TextFormatter{
		ForceColors:   true,
		FullTimestamp: true,
	})

	log.Info("信息")
	log.Warn("警告")
	log.Error("错误")
}
