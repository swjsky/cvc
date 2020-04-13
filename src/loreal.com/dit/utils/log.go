package utils

import (
	"io"
	"log"
)

//IOLog - IOLog
type IOLog struct {
	Trace   *log.Logger
	Info    *log.Logger
	Warning *log.Logger
	Error   *log.Logger
}

//NewIOLog - New IOLog
func NewIOLog(
	traceHandle io.Writer,
	infoHandle io.Writer,
	warningHandle io.Writer,
	errorHandle io.Writer,
) *IOLog {
	return &IOLog{
		Trace: log.New(
			traceHandle,
			"\n[TRACE] ",
			log.Ldate|
				log.Ltime|
				log.Lshortfile,
		),

		Info: log.New(infoHandle,
			"\n[INFO] ",
			log.Ldate|
				log.Ltime|
				log.Lshortfile,
		),

		Warning: log.New(
			warningHandle,
			"\n[WARNING] ",
			log.Ldate|
				log.Ltime|
				log.Lshortfile,
		),

		Error: log.New(
			errorHandle,
			"\n[ERROR] ",
			log.Ldate|
				log.Ltime|
				log.Lshortfile,
		),
	}
}
