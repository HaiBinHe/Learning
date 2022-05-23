package main

import (
	"context"
	"fmt"
	"github.com/go-redis/redis/v8"
	"log"
	"time"
)

type RedisBucket struct {
	key string	//redis key
	capacity int64   //令牌桶 容量
	rate time.Duration // 定时器
	num int64 // 每隔rate的时间，插入num数据
	tokens int64 //当前桶中令牌数量
	lastTokenSec int64	//上一次放入令牌桶的timestamp
	rdb  *redis.Client
}

func newRedisClient(key string, capacity int64) *redis.Client{
	rdb := redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
		Password: "",
	})
	ping, err := rdb.Ping(context.Background()).Result()
	if err != nil {
		log.Fatalf("redis ping err:%v",err)
		return nil
	}
	log.Println("redis ping:", ping)
	//设置key值,初始值为capacity

	boolCmd := rdb.SetNX(context.Background(), key, capacity, 30 * time.Minute)
	if boolCmd.Err() != nil {
		log.Fatalf("redis SetNX err:%v",boolCmd.Err())
		return nil
	}
	_, err = rdb.Persist(context.Background(), key).Result()
	if err != nil {
		log.Fatalf("redis Persist err :%v", err)
	}
	fmt.Println("redis key :", boolCmd.String())
	return rdb
}
func NewTokenBucket(key string, capacity int64, rate time.Duration, num int64) *RedisBucket{
	return &RedisBucket{
		key: key,
		capacity: capacity,
		rate: rate,
		num: num,
		tokens: capacity, //初始时 bucket是满的
		lastTokenSec: time.Now().Unix(),
		rdb: newRedisClient(key, capacity),
	}
}

func (rb *RedisBucket) Allow() bool {

	now := time.Now().Unix()
	//获取当前bucket中token的数量
	nums, err := rb.rdb.Get(context.Background(), rb.key).Int64()
	if err != nil {
		log.Fatalf("redis get err %v:", err)
		return false
	}
	//更新时间
	rb.lastTokenSec = now
	//根据redis中token的数量
	if nums > 0{
		cmd := rb.rdb.DecrBy(context.Background(), rb.key, 1)
		if cmd.Err() != nil {
			log.Fatalf("redis DecrBy err %v:", cmd.Err())
			return false
		}
		//更新令牌数量
		rb.tokens, _= rb.rdb.Get(context.Background(), rb.key).Int64()
		return true
	}
	return false
}
//定时器 按rate增加令牌
func (rb *RedisBucket) SetTicker() {
	var timer *time.Ticker
	timer = time.NewTicker(time.Second * rb.rate)
	rb.pushTokens(timer)
}
func (rb *RedisBucket) pushTokens(timer *time.Ticker) {
	for{
		select {
		case <-timer.C:
			nums, err := rb.rdb.IncrBy(context.Background(), rb.key, rb.num).Result()
			if err != nil {
				log.Fatalf("redis IncreBy err :%v", err)
			}
			//token数量大于容量
			if nums > rb.capacity{
				err = rb.rdb.DecrBy(context.Background(), rb.key, nums - rb.capacity).Err()
				if err != nil {
					log.Fatalf("redis DecreBy err :%v", err)
				}
			}
		}
	}
}

