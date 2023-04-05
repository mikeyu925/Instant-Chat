package models

import (
	"IM/utils"
	"context"
	"encoding/json"
	"fmt"
	"github.com/fatih/set"
	"github.com/go-redis/redis/v8"
	"github.com/gorilla/websocket"
	"github.com/spf13/viper"
	"gorm.io/gorm"
	"net"
	"net/http"
	"strconv"
	"sync"
	"time"
)

type MessageType int

const (
	One2OneChat MessageType = 1 // 私聊
	GroupChat   MessageType = 2 // 群聊
	HeartBeat   MessageType = 3 //心跳检测
)

type InfoType int

const (
	TextMsg    InfoType = 1 // 文本消息
	EmojiMsg   InfoType = 2 // 表情
	VoiceMsg   InfoType = 3 // 语音
	PictureMsg InfoType = 4 // 图片消息
)

// 消息
type Message struct {
	gorm.Model
	UserId     int64       //发送者
	TargetId   int64       //接受者
	Type       MessageType //发送类型  1私聊  2群聊  3心跳
	Media      InfoType    //消息类型  1文字 2表情包 3语音 4图片
	Content    string      //消息内容
	CreateTime uint64      //创建时间 ==> 防止发送重复内容数据的覆盖问题
	ReadTime   uint64      //读取时间
	Pic        string
	Url        string
	Desc       string
	Amount     int //其他数字统计
}

func (table *Message) TableName() string {
	return "message"
}

// 连接结点
type Node struct {
	Conn          *websocket.Conn //连接
	Addr          string          //客户端地址
	FirstTime     uint64          //首次连接时间
	HeartbeatTime uint64          //心跳时间
	LoginTime     uint64          //登录时间
	DataQueue     chan []byte     //消息
	GroupSets     set.Interface   //好友 / 群
}

// 映射关系 存储连接
var clientMap map[int64]*Node = make(map[int64]*Node, 0)

// 读写锁
var rwLocker sync.RWMutex

// 需要 ：发送者ID ，接受者ID ，消息类型，发送的内容，发送类型
func Chat(writer http.ResponseWriter, request *http.Request) {
	//1.  获取参数 并 检验 token 等合法性
	//token := query.Get("token")
	query := request.URL.Query()
	Idstr := query.Get("userId")
	userId, _ := strconv.ParseInt(Idstr, 10, 64)
	//msgType := query.Get("type")
	//targetId := query.Get("targetId")
	//context := query.Get("context")
	isvalida := true //checkToke()  待.........
	conn, err := (&websocket.Upgrader{
		//token 校验
		CheckOrigin: func(r *http.Request) bool {
			return isvalida
		},
	}).Upgrade(writer, request, nil)

	if err != nil {
		fmt.Println(err)
		return
	}
	//2.获取conn
	currentTime := uint64(time.Now().Unix())
	node := &Node{
		Conn:          conn,
		Addr:          conn.RemoteAddr().String(), //客户端地址
		HeartbeatTime: currentTime,                //心跳时间
		LoginTime:     currentTime,                //登录时间
		DataQueue:     make(chan []byte, 50),
		GroupSets:     set.New(set.ThreadSafe),
	}
	//3.TODO 用户关系

	//4. userid 跟 node绑定 并加锁
	rwLocker.Lock()
	clientMap[userId] = node
	rwLocker.Unlock()
	//5.完成发送逻辑
	go sendProc(node)
	//6.完成接收逻辑
	go recvProc(node)

	//7.加入在线用户到缓存
	SetUserOnlineInfo("online_"+Idstr, []byte(node.Addr), time.Duration(viper.GetInt("timeout.RedisOnlineTime"))*time.Hour)
	//sendMsg(userId, []byte("欢迎进入聊天系统"))
}

// 发送线程
func sendProc(node *Node) {
	for {
		select {
		case data := <-node.DataQueue:
			fmt.Println("[ws]sendProc >>>> msg :", string(data))
			err := node.Conn.WriteMessage(websocket.TextMessage, data)
			if err != nil {
				fmt.Println(err)
				return
			}
		}
	}
}

// 接收线程
func recvProc(node *Node) {
	for {
		_, data, err := node.Conn.ReadMessage()
		if err != nil {
			fmt.Println(err)
			return
		}
		msg := Message{}
		// 解析json字符串
		err = json.Unmarshal(data, &msg)
		if err != nil {
			fmt.Println(err)
		}
		//心跳检测 msg.Media == -1 || msg.Type == 3
		if msg.Type == HeartBeat {
			currentTime := uint64(time.Now().Unix())
			node.Heartbeat(currentTime)
		} else {
			dispatch(data)
			broadMsg(data) //todo 将消息广播到局域网
			fmt.Println("[ws] recvProc <<<<< ", string(data))
		}

	}
}

var udpsendChan chan []byte = make(chan []byte, 1024)

func broadMsg(data []byte) {
	udpsendChan <- data
}

func init() {
	go udpSendProc()
	go udpRecvProc()
	fmt.Println("init goroutine ")
}

// 完成udp数据发送协程
func udpSendProc() {
	con, err := net.DialUDP("udp", nil, &net.UDPAddr{
		IP:   net.IPv4(192, 168, 0, 255), // 192.168.31.54
		Port: 3000,
	})
	defer con.Close()
	if err != nil {
		fmt.Println(err)
	}
	for {
		select {
		case data := <-udpsendChan:
			fmt.Println("udpSendProc  data :", string(data))
			_, err = con.Write(data)
			if err != nil {
				fmt.Println(err)
				return
			}
		}
	}

}

// 完成udp数据接收协程
func udpRecvProc() {
	conn, err := net.ListenUDP("udp", &net.UDPAddr{
		IP:   net.IPv4zero,
		Port: 3000,
	})
	if err != nil {
		fmt.Println(err)
	}
	defer conn.Close()
	for {
		var buf [512]byte
		n, err := conn.Read(buf[0:])
		if err != nil {
			fmt.Println(err)
			return
		}
		fmt.Println("udpRecvProc  data :", string(buf[0:n]))
		dispatch(buf[0:n])
	}
}

// 后端调度逻辑处理--> 进行转发
func dispatch(data []byte) {
	msg := Message{}
	msg.CreateTime = uint64(time.Now().Unix()) // 获得消息的创建时间
	err := json.Unmarshal(data, &msg)
	if err != nil {
		fmt.Println(err)
		return
	}
	switch msg.Type {
	case One2OneChat: //私信
		fmt.Println("dispatch  data :", string(data))
		sendMsg(msg.TargetId, data)
	case GroupChat: //群发
		sendGroupMsg(msg.TargetId, data) //发送的群ID ，消息内容
		//case HeartBeat: // 心跳
		//	node.Heartbeat()
	}
}

// 将用户userid加入群comid/群名或者群id => 相当于创建了一个关系contact
func JoinGroup(userId uint, comId string) (int, string) {
	contact := Contact{}
	contact.OwnerId = userId
	//contact.TargetId = comId
	contact.Type = 2
	community := Community{}

	// 查询群是否存在「通过群名或者群id」
	utils.DB.Where("id=? or name=?", comId, comId).Find(&community)
	if community.Name == "" {
		return -1, "没有找到群"
	}
	// 拿出群id，应该是用查找到的群id去找contact
	utils.DB.Where("owner_id=? and target_id=? and type =2 ", userId, community.ID).Find(&contact)

	//utils.DB.Where("owner_id=? and target_id=? and type =2 ", userId, comId).Find(&contact)

	if !contact.CreatedAt.IsZero() {
		return -1, "已加过此群"
	} else {
		contact.TargetId = community.ID
		utils.DB.Create(&contact)
		return 0, "加群成功"
	}
}

func sendGroupMsg(targetId int64, msg []byte) {
	fmt.Println("开始群发消息")
	// 通过群id找到所有用户的信息
	userIds := SearchUserByGroupId(uint(targetId))
	for i := 0; i < len(userIds); i++ {
		//排除给自己的
		if targetId != int64(userIds[i]) {
			sendMsg(int64(userIds[i]), msg)
		}
	}
}

func sendMsg(userId int64, msg []byte) {
	// 加锁
	rwLocker.RLock()
	node, ok := clientMap[userId]
	rwLocker.RUnlock()

	jsonMsg := Message{}
	json.Unmarshal(msg, &jsonMsg)
	// 获取一个上下文
	ctx := context.Background()
	targetIdStr := strconv.Itoa(int(userId))
	userIdStr := strconv.Itoa(int(jsonMsg.UserId))
	jsonMsg.CreateTime = uint64(time.Now().Unix())
	r, err := utils.RDB.Get(ctx, "online_"+userIdStr).Result()
	if err != nil {
		fmt.Println(err)
	}
	// 如果用户在线
	if r != "" {
		if ok {
			fmt.Println("sendMsg >>> ", userId, "  msg:", string(msg))
			node.DataQueue <- msg // 通过管道发送消息
		}
	}
	var key string
	if userId > jsonMsg.UserId { // 这样是为了进行方便的渲染
		key = "msg_" + userIdStr + "_" + targetIdStr
	} else {
		key = "msg_" + targetIdStr + "_" + userIdStr
	}
	res, err := utils.RDB.ZRevRange(ctx, key, 0, -1).Result()
	if err != nil {
		fmt.Println(err)
	}
	score := float64(cap(res)) + 1
	ress, e := utils.RDB.ZAdd(ctx, key, &redis.Z{score, msg}).Result() //jsonMsg
	//res, e := utils.Red.Do(ctx, "zadd", key, 1, jsonMsg).Result() //备用 后续拓展 记录完整msg
	if e != nil {
		fmt.Println(e)
	}
	fmt.Println(ress)
}

// 需要重写此方法才能完整的msg转byte[]
func (msg Message) MarshalBinary() ([]byte, error) {
	return json.Marshal(msg)
}

// 每隔一段时间进行超时连接清理
func CleanConnection(param interface{}) (result bool) {
	result = true
	defer func() {
		if r := recover(); r != nil {
			fmt.Println("cleanConnection err", r)
		}
	}()
	//fmt.Println("定时任务,清理超时连接 ", param)
	//node.IsHeartbeatTimeOut()
	currentTime := uint64(time.Now().Unix()) // 当前时间获取
	for i := range clientMap {               // 遍历所有连接，关闭超时连接
		node := clientMap[i]
		if node.IsHeartbeatTimeOut(currentTime) {
			node.Conn.Close()
			// 应该把当前连接从clientMap移除?
			//delete(clientMap, i)
		}
	}
	return result
}

// 更新用户心跳
func (node *Node) Heartbeat(currentTime uint64) {
	node.HeartbeatTime = currentTime
	return
}

// 用户心跳是否超时
func (node *Node) IsHeartbeatTimeOut(currentTime uint64) (timeout bool) {
	if node.HeartbeatTime+viper.GetUint64("timeout.HeartbeatMaxTime") <= currentTime {
		fmt.Println(node.Addr, ": 心跳超时..... 关闭连接：")
		timeout = true
	}
	return
}

// 获取缓存里面的消息
func RedisMsg(userIdA int64, userIdB int64, start int64, end int64, isRev bool) []string {
	rwLocker.RLock()
	//node, ok := clientMap[userIdA]
	rwLocker.RUnlock()
	//jsonMsg := Message{}
	//json.Unmarshal(msg, &jsonMsg)
	ctx := context.Background()
	userIdStr := strconv.Itoa(int(userIdA))
	targetIdStr := strconv.Itoa(int(userIdB))
	var key string
	if userIdA > userIdB {
		key = "msg_" + targetIdStr + "_" + userIdStr
	} else {
		key = "msg_" + userIdStr + "_" + targetIdStr
	}
	//key = "msg_" + userIdStr + "_" + targetIdStr
	//rels, err := utils.Red.ZRevRange(ctx, key, 0, 10).Result()  //根据score倒叙

	var rels []string
	var err error
	if isRev {
		rels, err = utils.RDB.ZRange(ctx, key, start, end).Result()
	} else {
		rels, err = utils.RDB.ZRevRange(ctx, key, start, end).Result()
	}
	if err != nil {
		fmt.Println(err) //没有找到
	}
	// 发送推送消息
	/**
	// 后台通过websoket 推送消息
	for _, val := range rels {
		fmt.Println("sendMsg >>> userID: ", userIdA, "  msg:", val)
		node.DataQueue <- []byte(val)
	}**/
	return rels
}
