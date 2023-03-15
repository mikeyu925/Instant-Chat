




gorm.logger 作用?
> 开启慢查询日志?


#### 报错记录

---
Incorrect datetime value: '0000-00-00' for column 'login_time' at row 1
> 应该是日期字段在mysql 5.7 之后不能为'0000-00-00 00:00:00'，这里采用设置默认值解决
> ```
> LoginTime     time.Time `gorm:"default:NULL"`
> HeartbeatTime time.Time `gorm:"default:NULL"`
> ```
---

