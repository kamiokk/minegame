package logHepler

import (
	"log"
	"fmt"
	"time"
	"github.com/gin-gonic/gin"
)

const (
	//INFO level
	levelInfo = "INFO"
	//WARN level
	levelWarn = "WARN"
	//ERROR level
	levelError = "ERROR"
)

// Do log a record
func Do(c *gin.Context,level,format string,args ...interface{}) {
	now := time.Now().Format("2006-01-02 15:04:05")
	preformat := fmt.Sprintf("[%s] [%s] %s %s %s\n",now,level,format,c.Request.Method,c.Request.RequestURI)
	log.Printf(preformat,args)
}

// Info is shortcut for Do(c,logHelper.INFO,...)
func Info(c *gin.Context,format string,args ...interface{}) {
	Do(c,levelInfo,format,args)
}

// Warn is shortcut for Do(c,logHelper.WARN,...)
func Warn(c *gin.Context,format string,args ...interface{}) {
	Do(c,levelWarn,format,args)
}

// Error is shortcut for Do(c,logHelper.ERROR,...)
func Error(c *gin.Context,format string,args ...interface{}) {
	Do(c,levelError,format,args)
}