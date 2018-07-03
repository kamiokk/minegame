package redis

import (
    "github.com/gin-gonic/gin"
    "os"
    "time"
    "math/rand"
    "encoding/json"
	"github.com/gomodule/redigo/redis"
	"github.com/kamiokk/minegame/helper/logHelper"
)

const getLockScript string = `
    if redis.call('exists', KEYS[1]) == 0 
        then return redis.call('setex', KEYS[1], ARGV[1], ARGV[2])
    end`
const releaseLockScript string = `
    if redis.call('get', KEYS[1]) == ARGV[1] then
        return redis.call('del', KEYS[1]) or true
    end`

var pool *redis.Pool
var locker *redis.Script
var unLocker *redis.Script

func InitHelper() {
    host := os.Getenv("REDIS_HOST")
    databaseName := os.Getenv("REDIS_PORT")
    pool = &redis.Pool {
        MaxIdle: 30,
        IdleTimeout: 240 * time.Second,
        Dial: func () (redis.Conn, error) { return redis.Dial("tcp", host + ":" + databaseName ) },
    }
    //init lua script
    locker = redis.NewScript(1,getLockScript)
    unLocker = redis.NewScript(1,releaseLockScript)
    c := pool.Get()
    locker.Load(c)
    unLocker.Load(c)
    c.Close()
}

func Pool() *redis.Pool {
    return pool
}

func GetConn(c *gin.Context) redis.Conn {
    rc,ok := c.Get("__redisConn")
    if ok {
        _,ok = rc.(redis.Conn)
        if ok {
            logHelper.Debug(c,"Get Redis Connect From Context")
            return rc.(redis.Conn)
        }
    }
    logHelper.Debug(c,"New Redis Connetion")
    conn := pool.Get()
    c.Set("__redisConn",conn)
    return conn
}

func CloseConn(c *gin.Context) {
    rc,ok := c.Get("__redisConn")
    if ok {
        _,ok = rc.(redis.Conn)
        if ok {
            logHelper.Debug(c,"Closing Redis Connect In Context")
            rc.(redis.Conn).Close()
        }
    }
}

func EndHelper() {
    pool.Close()
}

func GetLockByTimeout(c redis.Conn,timeout time.Duration,lockKey string,lockID int,lockTime uint) bool {
    var re string
    var err error
    begin := time.Now()
    for re != "OK" && time.Since(begin) < timeout {
        re,err = redis.String(locker.Do(c,lockKey,lockTime,lockID))
        if err != nil {
			logHelper.DebugNoContext("GetLockError:%v",err)
            break
        }
        time.Sleep(time.Microsecond)
    }
    return re == "OK"
}

func ReleaseLock(c redis.Conn,lockKey string,lockID int) {
    unLocker.Do(c,lockKey,lockID)
}

func RandLockId() int {
    return rand.Intn(1000000)
}

func SetStructExp(c redis.Conn,key string,value interface{},expire int) error {
    if jsonStr,err := json.Marshal(value);err != nil {
        logHelper.DebugNoContext("SetStructExpError:%v",err)
        return err
    } else {
        if _,err := c.Do("setex",key,expire,jsonStr);err != nil {
            logHelper.DebugNoContext("SetStructExpError:%v",err)
            return err
        }
    }
    return nil
}

func SetStruct(c redis.Conn,key string,value interface{}) error {
    if jsonStr,err := json.Marshal(value);err != nil {
        logHelper.DebugNoContext("SetStructError:%v",err)
        return err
    } else {
        if _,err := c.Do("set",key,jsonStr);err != nil {
            logHelper.DebugNoContext("SetStructError:%v",err)
            return err
        }
    }
    return nil
}

func FetchStruct(c redis.Conn,key string,out interface{}) error {
    if reply,err := redis.Bytes(c.Do("get",key)); err != nil {
		logHelper.DebugNoContext("FetchStructError:%v",err)
        return err
    } else {
        if err := json.Unmarshal(reply,out);err != nil {
			logHelper.DebugNoContext("FetchStructError:%v",err)
            return err
        }
    }
    return nil
}