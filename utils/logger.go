package utils

import log "github.com/sirupsen/logrus"

func init() {
	log.SetLevel(log.InfoLevel)
	log.SetFormatter(&log.JSONFormatter{
		FieldMap: log.FieldMap{
			log.FieldKeyTime:  "timestamp",
			log.FieldKeyLevel: "level",
			log.FieldKeyMsg:   "message",
		}})
}
