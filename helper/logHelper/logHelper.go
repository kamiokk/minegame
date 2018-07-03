package logHelper

import (
    "log"
    "fmt"
	"time"
	"os"
    "github.com/gin-gonic/gin"
)

const (
    //INFO level
    levelInfo = "INFO"
    //WARN level
    levelWarn = "WARN"
    //ERROR level
	levelError = "ERROR"
	//DEBUG level
	levelDebug = "DEBUG"
)

// Do log a record
func Do(c *gin.Context,level,format string,args ...interface{}) {
    now := time.Now().Format("2006-01-02 15:04:05")
    preformat := fmt.Sprintf("[%s] [%s] %s %s %s\n",now,level,format,c.Request.Method,c.Request.RequestURI)
    log.Printf(preformat,args...)
}

// Info is shortcut for Do(c,logHelper.levelInfo,...)
func Info(c *gin.Context,format string,args ...interface{}) {
    Do(c,levelInfo,format,args...)
}

// Warn is shortcut for Do(c,logHelper.levelWarn,...)
func Warn(c *gin.Context,format string,args ...interface{}) {
    Do(c,levelWarn,format,args...)
}

// Error is shortcut for Do(c,logHelper.levelError,...)
func Error(c *gin.Context,format string,args ...interface{}) {
    Do(c,levelError,format,args...)
}

// Debug is shortcut for Do(c,logHelper.levelDebug,...)
func Debug(c *gin.Context,format string,args ...interface{}) {
	if os.Getenv("MINE_GAME_ENVIRONMENT") == "product" {
		return
	}
    Do(c,levelDebug,format,args...)
}

// DebugNoContext is shortcut for Do(c,logHelper.levelDebug,...)
func DebugNoContext(format string,args ...interface{}) {
	if os.Getenv("MINE_GAME_ENVIRONMENT") == "product" {
		return
	}
    now := time.Now().Format("2006-01-02 15:04:05")
    preformat := fmt.Sprintf("[%s] [%s] %s\n",now,levelDebug,format)
    log.Printf(preformat,args...)
}