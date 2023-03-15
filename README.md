




gorm.logger 作用?
> 开启慢查询日志?


govalidator 实现修改用户信息时电话和邮箱格式的校验


#### 报错记录

---
Incorrect datetime value: '0000-00-00' for column 'login_time' at row 1
> 应该是日期字段在mysql 5.7 之后不能为'0000-00-00 00:00:00'，这里采用设置默认值解决
> ```
> LoginTime     time.Time `gorm:"default:NULL"`
> HeartbeatTime time.Time `gorm:"default:NULL"`
> ```
---
DENIED Redis is running in protected mode because protected mode is enabled, no bind address was specified, no authentication password is requested to clients.
> 错误原因：由于redis的保护模式开启了，并且没有绑定ip地址，没有密码认证

---

error]: dial tcp :6379: connect: connection refused

解决方法：修改redis.config 中的`bind 127.0.0.1 ::1` 为 `bind 0.0.0.0`
>关闭redis :sudo systemctl stop redis-server 

>重启redis :sudo systemctl restart redis-server 


---
