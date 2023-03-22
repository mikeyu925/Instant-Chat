




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


Mysql：Incorrect string value: '\xE4\xBA\xA4\xE6\xB5\x81...' for column 'name' at row 1

> 应该是插入的数据有中午，不支持，而导致的，尝试了下全英文插入，可以插入
> 看了下建表时的语句:
> ```sql
> CREATE TABLE `community` (
> `id` bigint(20) unsigned NOT NULL AUTO_INCREMENT,
> `created_at` datetime(3) DEFAULT NULL,
> `updated_at` datetime(3) DEFAULT NULL,
> `deleted_at` datetime(3) DEFAULT NULL,
> `name` longtext,
> `owner_id` bigint(20) unsigned DEFAULT NULL,
> `img` longtext,
> `desc` longtext,
> PRIMARY KEY (`id`),
> KEY `idx_community_deleted_at` (`deleted_at`)
> ) ENGINE=InnoDB DEFAULT CHARSET=latin1
> ```
>
> 因为一些一键安装包的环境, `my.ini` 默认配置的字符集是 `latin1` 或者其他, 如果此时一旦不注意, 使用sql语句去创建数据库, 表 默认都是 `latin1`, 因为有些字符集是不能存储中文的,如果需要存储中文, 需要使用GBK,utf8...等字符集
解决方案：

修改字符集：

- 数据库

  ```sql
  ALTER DATABASE `test_db` CHARACTER SET 'utf8' COLLATE 'utf8_general_ci';
  ```

- 数据表

  ```sql
  ALTER TABLE `test_db`.`user` CHARACTER SET = utf8mb4, COLLATE = utf8mb4_bin;
  ```

- 字段

  ```sql
  ALTER TABLE `test_db`.`username`  MODIFY COLUMN `password` varchar(30)  CHARACTER SET utf8mb4 COLLATE utf8mb4_bin;
  ```

  

