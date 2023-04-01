




gorm.logger 作用?
> 开启慢查询日志?

govalidator 实现修改用户信息时电话和邮箱格式的校验





#### 未来改进

---

加入自定义中间件

> 还没想好中间件做什么功能





#### Bug记录

---

发送消息对方收不到





群发消息发送消息发不出去





加入群时输入群名报错

> 加入群的时候，通过cominfo 群名/群id进行查找，然后用找到的群的id去查找contact关系，而不应该用输入的cominfo进行查找，因为输入的有可能是群名
>
> ```go
> // 将用户userid加入群comid/群名或者群id => 相当于创建了一个关系contact
> func JoinGroup(userId uint, comId string) (int, string) {
> 	contact := Contact{}
> 	contact.OwnerId = userI
> 	contact.Type = 2
>   
> 	community := Community{}
> 
> 	// 查询群是否存在「通过群名或者群id」
> 	utils.DB.Where("id=? or name=?", comId, comId).Find(&community)
> 	if community.Name == "" {
> 		return -1, "没有找到群"
> 	}
> 	// 拿出群id，应该是用查找到的群id去找contact
> 	groupid := community.ID
> 	utils.DB.Where("owner_id=? and target_id=? and type =2 ", userId, groupid).Find(&contact)
> 
> 	// utils.DB.Where("owner_id=? and target_id=? and type =2 ", userId, comId).Find(&contact)
> 
> 	if !contact.CreatedAt.IsZero() {
> 		return -1, "已加过此群"
> 	} else {
> 		contact.TargetId = community.ID
> 		utils.DB.Create(&contact)
> 		return 0, "加群成功"
> 	}
> }
> ```

---

前端点击了一次，发了两次

> 防止页面请求穿透以及禁止高频发送：设置一个变量，只有变量为true时，可以发送请求，发送完后设置为false,一段时间后变量为true
>
> ```go
> _addfriend: function (dstobj) {
>   //防止一次点击 穿透访问多次
>   if (this.isDisable) { // 如果可以发起请求
>     this.setTimeFlag() // set isDisable
>     var that = this
>     // 发起添加好友请求
>     post("contact/addfriend", { targetName: dstobj, userId: userId() }, function (res) {
>       if (res.Code == 0) {
>         mui.toast("添加成功");
>         that.loadfriends(); // 加载好友列表
>       } else {
>         mui.toast(res.Msg);
>       }
>     })
>   }
> },
> ```
>
> <img src="./README.assets/image-20230401152810182.png" alt="image-20230401152810182" style="zoom:50%;" />






#### 报错记录

---
Incorrect datetime value: '0000-00-00' for column 'login_time' at row 1
> 应该是日期字段在mysql 5.7 之后不能为'0000-00-00 00:00:00'，这里采用设置默认值解决
> ```go
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

> 应该是插入的数据有中文，不支持，而导致的，尝试了下全英文插入，可以插入
> 看了下建表时的语句:
>
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
> 解决方案：
>
> 修改字符集：
>
> - 数据库
>
>   ```sql
>   ALTER DATABASE `test_db` CHARACTER SET 'utf8' COLLATE 'utf8_general_ci';
>   ```
>
> - 数据表
>
>   ```sql
>   ALTER TABLE `test_db`.`user` CHARACTER SET = utf8mb4, COLLATE = utf8mb4_bin;
>   ```
>
> - 字段
>
>   ```sql
>   ALTER TABLE `test_db`.`username`  MODIFY COLUMN `password` varchar(30)  CHARACTER SET utf8mb4 COLLATE utf8mb4_bin;
>   ```
>

