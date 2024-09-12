package utils

import (
	"math/rand"
	"time"
)

const letterBytes = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

func RandString(n int) string {
	rand.NewSource(time.Now().UnixNano()) // 以当前时间作为随机数种子
	b := make([]byte, n)
	for i := range b {
		b[i] = letterBytes[rand.Intn(len(letterBytes))] // 随机选择字母或数字
	}
	return string(b)
}
