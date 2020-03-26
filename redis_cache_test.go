package tiptop

import (
	"fmt"
	"testing"
	"time"
)

func TestRedisCache_Clean(t *testing.T) {
	client, _ := newRedisClient(&Config{
		RedisAddr: "172.18.29.81:6379",
	})
	if err := client.Ping().Err(); err != nil {
		fmt.Println(err.Error())
	}
	client.Set(KeyPrefix+"1", "test1", time.Minute)
	client.Set(KeyPrefix+"2", "test2", time.Minute)
	client.Set(KeyPrefix+"3", "test3", time.Minute)
	client.Set(KeyPrefix+"4", "test4", time.Minute)
	iterator := client.Scan(0, KeyPrefix+"*", 2).Iterator()
	for iterator.Next() {
		fmt.Println(iterator.Val())
	}
	fmt.Println("开始删除")
	iterator = client.Scan(0, KeyPrefix+"*", 2).Iterator()
	for iterator.Next() {
		client.Del(iterator.Val())
	}
	fmt.Println("删除成功")
	iterator = client.Scan(0, KeyPrefix+"*", 2).Iterator()
	for iterator.Next() {
		fmt.Println(iterator.Val())
	}
}
