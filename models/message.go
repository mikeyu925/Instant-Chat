package models

import (
	"IM/utils"
	"context"
	"encoding/json"
	"fmt"
	"github.com/fatih/color"
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

// Chat
//
//	@Description: 由用户A发送至用户B消息
//	@param writer
//	@param request
func Chat(writer http.ResponseWriter, request *http.Request) {
	//1.  获取参数 并 检验 token 等合法性
	query := request.URL.Query()
	// 获取用户的ID
	Idstr := query.Get("userId")
	userId, _ := strconv.ParseInt(Idstr, 10, 64)

	isvalida := true // 检验token合法性
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
	currentTime := uint64(time.Now().Unix()) // 获取当前时间
	node := &Node{
		Conn:          conn,
		Addr:          conn.RemoteAddr().String(), //客户端地址
		HeartbeatTime: currentTime,                //心跳时间
		LoginTime:     currentTime,                //登录时间
		DataQueue:     make(chan []byte, 50),
		GroupSets:     set.New(set.ThreadSafe),
	}
	//3. userid 跟 node绑定 并加锁
	rwLocker.Lock()
	clientMap[userId] = node
	rwLocker.Unlock()
	//4.开启发送数据协程
	go sendProc(node)
	//5.开启接收数据协程
	go recvProc(node)
	//6.加入在线用户到缓存
	SetUserOnlineInfo("online_"+Idstr, []byte(node.Addr), time.Duration(viper.GetInt("timeout.RedisOnlineTime"))*time.Hour)
}

// sendProc
//
//	@Description: 发送数据线程
//	@param node 当前一个连接
func sendProc(node *Node) {
	for {
		select {
		case data := <-node.DataQueue: // 如果有数据要发送
			// fmt.Println("[ws]sendProc >>>> msg :", string(data))
			err := node.Conn.WriteMessage(websocket.TextMessage, data)
			if err != nil {
				fmt.Println(err)
				return
			}
		}
	}
}

// recvProc
//
//	@Description: 接收数据线程
//	@param node
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
			node.Heartbeat(currentTime) // 更新用户心跳
		} else {
			dispatch(data) // 进行消息的调度
			broadMsg(data) // 将消息广播到局域网 TODO
			// fmt.Println("[ws] recvProc <<<<< ", string(data))
		}
	}
}

var udpsendChan chan []byte = make(chan []byte, 1024)

// broadMsg
//
//	@Description: 广播消息
//	@param data  广播的数据
func broadMsg(data []byte) {
	udpsendChan <- data
}

// InitUDPProc
//
//	@Description: 初始化UDP相关协程
func InitUDPProc() {
	go udpSendProc()
	go udpRecvProc()
	color.Green("Init UPD Send/Rec Goroutine Successfully!")
}

// 完成udp数据发送协程
func udpSendProc() {
	con, err := net.DialUDP("udp", nil, &net.UDPAddr{
		IP:   net.IPv4(192, 168, 0, 255), // 192.168.31.54
		Port: viper.GetInt("port.udp"),
	})
	defer con.Close()
	if err != nil {
		fmt.Println(err)
	}
	for {
		select {
		case data := <-udpsendChan:
			// fmt.Println("udpSendProc  data :", string(data))
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
		Port: viper.GetInt("port.udp"),
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
		// fmt.Println("udpRecvProc  data :", string(buf[0:n]))
		dispatch(buf[0:n])
	}
}

// dispatch
//
//	@Description: 对消息进行调度、转发
//	@param data
func dispatch(data []byte) {
	msg := Message{}
	err := json.Unmarshal(data, &msg)
	msg.CreateTime = uint64(time.Now().Unix()) // 获得消息的创建时间
	if err != nil {
		fmt.Println(err)
		return
	}
	switch msg.Type {
	case One2OneChat: //私信
		sendMsg(msg.TargetId, data)
	case GroupChat: //群发
		sendGroupMsg(msg.TargetId, data)
	}
}

// JoinGroup
//
//	@Description: 通过群id或者群名查找并加入群
//	@param userId  当前用户userid == 申请加入者
//	@param comId  群id/群名
//	@return int  0：成功 -1：失败
//	@return string  响应信息
func JoinGroup(userId uint, comId string) (int, string) {
	contact := Contact{}
	contact.OwnerId = userId
	contact.Type = 2

	community := Community{}
	// 查询群是否存在「通过群名或者群id」
	utils.DB.Where("id=? or name=?", comId, comId).Find(&community)
	if community.Name == "" {
		return -1, "没有找到群"
	}
	// 群存在：拿出群id，应该是用查找到的群id去找contact
	// 判断是否已经加过此群
	utils.DB.Where("owner_id=? and target_id=? and type=?", userId, community.ID, _GroupType).Find(&contact)

	if !contact.CreatedAt.IsZero() {
		return -1, "已加过此群"
	} else {
		contact.TargetId = community.ID
		utils.DB.Create(&contact)
		return 0, "加群成功"
	}
}

// sendGroupMsg
//
//	@Description: 发送群消息
//	@param targetId 群ID
//	@param msg 消息
func sendGroupMsg(targetId int64, msg []byte) {
	// 通过群id找到所有用户的信息
	userIds := SearchUserByGroupId(uint(targetId))
	// 将msg推送给所有群中的用户
	for i := 0; i < len(userIds); i++ {
		//排除给自己的
		if targetId != int64(userIds[i]) {
			sendMsg(int64(userIds[i]), msg)
		}
	}
}

// sendMsg
//
//	@Description: 将消息msg发送至目标用户ID「后端只管发送，前端针对具体的消息类型和targetID进行渲染」
//	@param targetId  目标用户ID
//	@param msg 待发送消息
func sendMsg(targetId int64, msg []byte) {
	// 加锁
	rwLocker.RLock()
	recvNode, ok := clientMap[targetId]
	rwLocker.RUnlock()

	jsonMsg := Message{}
	json.Unmarshal(msg, &jsonMsg)
	// 获取一个上下文
	ctx := context.Background()
	recvIdStr := strconv.Itoa(int(targetId))       // 消息接受者ID
	sendIdStr := strconv.Itoa(int(jsonMsg.UserId)) // 消息发送者ID
	jsonMsg.CreateTime = uint64(time.Now().Unix())
	// 这里应该是判断对方用户是否在线吧
	r, err := utils.RDB.Get(ctx, "online_"+recvIdStr).Result()
	if err != nil {
		fmt.Println(err)
	}

	if r != "" { // 如果对方用户在线
		if ok {
			// fmt.Println("sendMsg >>> ", targetId, "  msg:", string(msg))
			recvNode.DataQueue <- msg // 通过管道发送消息
		}
	}
	// 用户id小的放在前面  这样是为了进行方便的渲染
	var key string
	if targetId > jsonMsg.UserId {
		key = "msg_" + sendIdStr + "_" + recvIdStr
	} else {
		key = "msg_" + recvIdStr + "_" + sendIdStr
	}
	// 查询出所有消息
	res, err := utils.RDB.ZRevRange(ctx, key, 0, -1).Result()
	if err != nil {
		fmt.Println(err)
	}
	score := float64(cap(res)) + 1                                     // 计算出当前的score
	ress, e := utils.RDB.ZAdd(ctx, key, &redis.Z{score, msg}).Result() //jsonMsg

	if e != nil {
		fmt.Println(e)
	}
	fmt.Println(ress)
}

// 获取缓存里面的消息
//
// RedisMsg
//
//	@Description: 读取用户A和用户B从start到end的缓存消息
//	@param userIdA
//	@param userIdB
//	@param start
//	@param end
//	@param isRev 是否倒序读取
//	@return []string 两个用户的聊天数据
func RedisMsg(userIdA int64, userIdB int64, start int64, end int64, isRev bool) []string {

	ctx := context.Background()
	userIdStr := strconv.Itoa(int(userIdA))
	targetIdStr := strconv.Itoa(int(userIdB))
	// 用户id比较小的放在前面
	var key string
	if userIdA > userIdB {
		key = "msg_" + targetIdStr + "_" + userIdStr
	} else {
		key = "msg_" + userIdStr + "_" + targetIdStr
	}

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
	return rels // 直接返回所有缓存消息， 交给前端处理
}

// 需要重写此方法才能完整的msg转byte[]
func (msg Message) MarshalBinary() ([]byte, error) {
	return json.Marshal(msg)
}

// CleanConnection
//
//	@Description: 对超时的连接进行清理
//	@param param
//	@return result
func CleanConnection(param interface{}) (result bool) {
	result = true
	defer func() {
		if r := recover(); r != nil {
			fmt.Println("cleanConnection err", r)
		}
	}()

	currentTime := uint64(time.Now().Unix()) // 当前时间获取
	for i := range clientMap {               // 遍历所有连接，关闭超时连接
		node := clientMap[i]
		if node.IsHeartbeatTimeOut(currentTime) {
			node.Conn.Close()
			// 应该把当前连接从clientMap移除?
			delete(clientMap, i)
		}
	}
	return result
}

// Heartbeat
//
//	@Description: 更新用户的心跳
//	@receiver node
//	@param currentTime
func (node *Node) Heartbeat(currentTime uint64) {
	node.HeartbeatTime = currentTime
	return
}

// IsHeartbeatTimeOut
//
//	@Description: 判断用户是否超时
//	@receiver node
//	@param currentTime
//	@return timeout
func (node *Node) IsHeartbeatTimeOut(currentTime uint64) (timeout bool) {
	if node.HeartbeatTime+viper.GetUint64("timeout.HeartbeatMaxTime") <= currentTime {
		//fmt.Println(node.Addr, ": 心跳超时..... 关闭连接：")
		timeout = true
	}
	return
}
