package models

import (
	"time"

	"github.com/garyburd/redigo/redis"
	"github.com/streadway/amqp"

	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/mysql"
	"github.com/micross/codingtest/utils"
	"github.com/spf13/viper"
)

var DB *gorm.DB
var RabbitMQ *amqp.Connection
var RedisPool *redis.Pool

func InitDB() {
	dsn := viper.GetString("mysql")
	db, err := gorm.Open("mysql", dsn)
	utils.FailOnError(err, "Failed to connect to MySQL")
	env := viper.GetString("env")
	if env == DevelopmentMode {
		db.LogMode(true)
	}
	DB = db
}

func InitRabbitMQ() {
	url := viper.GetString("rabbitmq")
	conn, err := amqp.Dial(url)
	utils.FailOnError(err, "Failed to connect to RabbitMQ")
	RabbitMQ = conn
}

func InitRedis() {
	url := viper.GetString("redis")
	RedisPool = &redis.Pool{
		MaxIdle:     100,
		MaxActive:   500,
		IdleTimeout: 480 * time.Second,
		Wait:        true,
		Dial: func() (redis.Conn, error) {
			c, err := redis.DialURL(url)
			if err != nil {
				return nil, err
			}
			return c, nil
		},
	}
}
