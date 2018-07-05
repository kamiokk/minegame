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

// InitHelper init this helper
func InitHelper(maxIdle int) {
    host := os.Getenv("REDIS_HOST")
    databaseName := os.Getenv("REDIS_PORT")
    pool = &redis.Pool {
        MaxIdle: maxIdle,
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

// Pool get pool instance
func Pool() *redis.Pool {
    return pool
}

// GetConn get redis.Conn
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

// CloseConn redis.Conn.Close()
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

// EndHelper close the pool
func EndHelper() {
    pool.Close()
}


// GetLockByTimeout get lock with timeout
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

// ReleaseLock release lock by key & id
func ReleaseLock(c redis.Conn,lockKey string,lockID int) {
    unLocker.Do(c,lockKey,lockID)
}

// RandLockId return pseudo-random int
func RandLockId() int {
    return rand.Intn(1000000)
}

// Marshal serialize the struct object
func Marshal(value interface{}) (jsonStr []byte,err error) {
    if jsonStr,err = json.Marshal(value);err != nil {
        logHelper.DebugNoContext("JsonEncodeFailed:%v",err)
        return nil,err
    }
    return jsonStr,nil
}

// Unmarshal unserialize
func Unmarshal(reply []byte,out interface{}) (err error) {
    if err = json.Unmarshal(reply,out);err != nil {
        logHelper.DebugNoContext("JsonDecodeFailed:%v",err)
    }
    return err
}

// SetStructExp serialize a sturct and set into redis with timeout
func SetStructExp(c redis.Conn,key string,value interface{},expire int) error {
    if jsonStr,err := Marshal(value);err != nil {
        return err
    } else {
        if _,err = c.Do("setex",key,expire,jsonStr);err != nil {
            logHelper.DebugNoContext("SetStructExpError:%v",err)
            return err
        }
    }
    return nil
}

// SetStruct serialize a sturct and set into redis
func SetStruct(c redis.Conn,key string,value interface{}) error {
    if jsonStr,err := Marshal(value);err != nil {
        return err
    } else {
        if _,err = c.Do("set",key,jsonStr);err != nil {
            logHelper.DebugNoContext("SetStructExpError:%v",err)
            return err
        }
    }
    return nil
}

// FetchStruct get and unserialize
func FetchStruct(c redis.Conn,key string,out interface{}) error {
    var reply []byte
    var err error
    if reply,err = redis.Bytes(c.Do("get",key)); err != nil {
		logHelper.DebugNoContext("FetchStructError:%v",err)
        return err
    }
    return Unmarshal(reply,out)
}

// LpushStruct serialize a sturct and push into queue
func LpushStruct(c redis.Conn,key string,value interface{}) error {
    if jsonStr,err := Marshal(value);err != nil {
        return err
    } else {
        if _,err = c.Do("lpush",key,jsonStr);err != nil {
            logHelper.DebugNoContext("LpushStructError:%v",err)
            return err
        }
    }
    return nil
}

// LpopStruct pop and unserialize
func LpopStruct(c redis.Conn,key string,out interface{}) (bool,error) {
    if reply,err := redis.Bytes(c.Do("lpop",key)); err != nil {
        return false,err
    } else {
        if len(reply) == 0 {
            return false,nil
        }
        err = Unmarshal(reply,out)
        return true,err
    }
}