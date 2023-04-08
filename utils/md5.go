package utils

import (
	"crypto/md5"
	"encoding/hex"
	"strings"
)

// 小写加密
func Md5Encode(data string) string {
	h := md5.New()
	h.Write([]byte(data))
	tempStr := h.Sum(nil)
	return hex.EncodeToString(tempStr)
}

// 大写加密
func MD5Encode(data string) string {
	return strings.ToUpper(Md5Encode(data))
}

// 加密
func MakePassword(plainpwd, salt string) string {
	return Md5Encode(plainpwd + salt) // 将用户密码 + 随机数 ==> MD5加密
}

// 解密 再进行一次加密
func ValidPassword(plainpwd, salt string, password string) bool {
	md := Md5Encode(plainpwd + salt)
	return md == password
}
