package Gj_galaxy

import (
	"database/sql"
	"fmt"
	"net"
	"net/url"
	"time"

	"Gj-galaxy/room"
	"Gj-galaxy/scene"

	"github.com/garyburd/redigo/redis"
	_ "github.com/go-sql-driver/mysql"
)

func PreRun() error {

	logger.Infof("[ Server ] Init with config")
	RegisterConfig(config)

	db, err := loadMysql(config.Db)
	if err != nil {
		return err
	}
	RegisterProvide(db)

	redisPool, err := loadRedis(config.Redis)
	if err != nil {
		return err
	}
	RegisterProvide(redisPool)

	logger.Infof("[ Server ] config process success")

	registerCoreServer()
	return nil
}

func Run() error {
	err := bindServer()
	if err != nil {
		return err
	}
	logger.Infof("[ Server ] Start...")
	err = StartServer()
	return err
}

func Exit() error {
	logger.Infof("[ Server ] Stop...")
	StopServer()
	return nil
}

func loadMysql(config DbConfig) (*sql.DB, error) {
	logger.Debugf("load mysql config")
	u, err := url.Parse(config.Dsn)
	if err != nil {
		return nil, err
	}
	host, port, err := net.SplitHostPort(u.Host)
	if err != nil {
		host = u.Host
	}
	if host == "" {
		return nil, fmt.Errorf("MysqlServer host is empty: %s", u)
	}
	if port == "" {
		port = "3306"
	}
	logger.Debugf("MysqlServer info: host->%s, port->%s", host, port)
	db, err := sql.Open(u.Scheme, fmt.Sprintf("%s@%s/%s", u.User, u.Host, u.Path))
	return db, err
}

func loadRedis(config RedisConfig) (*redis.Pool, error) {
	logger.Debugf("load redis config")
	u, err := url.Parse(config.Dsn)
	if err != nil {
		return nil, err
	}
	host, port, err := net.SplitHostPort(u.Host)
	if err != nil {
		host = u.Host
	}
	if host == "" {
		return nil, fmt.Errorf("RedisServer host is empty: %s", u)
	}
	if port == "" {
		port = "6379"
	}
	logger.Debugf("RedisServer info: host->%s, port->%s", host, port)
	address := net.JoinHostPort(host, port)
	options := []redis.DialOption{}
	password, isSet := u.User.Password()
	if isSet {
		options = append(options, redis.DialPassword(password))
	}

	pool := redis.Pool{
		MaxIdle:     3,
		IdleTimeout: 240 * time.Second,
		Dial: func() (redis.Conn, error) {
			return redis.Dial("tcp", address, options...)
		},
	}
	return &pool, err
}

func registerCoreServer() {

	platformServer := PlatformServer{}
	RegisterServer(&platformServer, "platform")

	webServer := WebServer{}
	RegisterServer(&webServer, "web")

	socketServer := SocketServer{}
	RegisterServer(&socketServer, "socket")

	roomServer := room.Server{}
	RegisterServer(&roomServer, "room")

	sceneServer := scene.Server{}
	RegisterServer(&sceneServer, "scene")

	restfulServer := RestfulServer{}
	RegisterServer(&restfulServer, "restful")
}
