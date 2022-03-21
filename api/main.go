package main

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/mysql"
)

// App  这只是个返回的结构,不是表结构
type App struct {
	Id               int         `json:"id" gorm:"id"`
	ClientToken      string      `json:"client_token" gorm:"client_token"`             // 客户端 token(客户端读取信息用)
	ServerToken      string      `json:"server_token" gorm:"server_token"`             // 服务端 token(服务端读取信息用)
	GatewayRules     string      `json:"gateway_rules" grom:"gateway_rules"`           // 监听端口,例如: 11:11;22:22;456:456-789
	ServerAddr       string      `json:"server_addr" gorm:"server_addr"`               // 服务器的外网地址，例如：192.168.1.1:8888(服务器真实 IP)
	ServerListenAddr string      `json:"server_listen_addr" gorm:"server_listen_addr"` // 服务器的监听地址，例如：0.0.0.0:8888
	ClientListenIP   string      `json:"client_listen_ip" gorm:"client_listen_ip"`     // 客户端的本地监听IP，例如：127.0.0.2
	Proxies          []Proxy     `json:"proxies"`
	DummyNodes       []DummyNode `json:"dummy_nodes"`
}

func (o App) TableName() string {
	//todo 补充表名
	return ""
}

//AppProxyDummyNode 中间件表  举例client_token为 111 的用户拥有 tx1 分组的 proxy 和 dn1 的 dummy_node
type AppProxyDummyNode struct {
	ClientToken  string `json:"client_token"`   // client_token
	ProxyTag     string `json:"proxy_tag"`      // proxy_tag
	DummyNodeTag string `json:"dummy_node_tag"` //dummy_node_tag
}

func (o AppProxyDummyNode) TableName() string {
	//todo 补充表名
	return ""
}

// Proxy 节点表
type Proxy struct {
	Id            int       `json:"id" gorm:"id"`
	ProxyTag      string    `gorm:"proxy_tag" json:"proxy_tag"`             // 指定外键
	Network       string    `json:"network" gorm:"network"`                 // 节点的类型，tcp/kcp
	Addr          string    `json:"addr" gorm:"addr"`                       // 节点的外网地址
	ListenAddr    string    `json:"listen_addr" gorm:"listen_addr"`         // 节点的监听地址
	AppUpdateTime time.Time `json:"app_update_time" gorm:"app_update_time"` // Apps最后更新的时间
}

func (o Proxy) TableName() string {
	//todo 补充表名
	return ""
}

// DummyNode 虚拟节点表
type DummyNode struct {
	Id           int    `json:"id" gorm:"id"`
	DummyNodeTag string `gorm:"dummy_node_tag" json:"dummy_node_tag"` // 指定外键
	Network      string `json:"network" gorm:"network"`               // 节点的类型，tcp/kcp/quic
	Addr         string `json:"addr" gorm:"addr"`                     // 节点的外网地址
}

func (o DummyNode) TableName() string {
	//todo 补充表名
	return ""
}

var db *gorm.DB
var err error

func main() {
	db, err = gorm.Open("mysql", "root:123456.ab@/api?charset=utf8&parseTime=True&loc=Local")
	if err != nil {
		panic(err)
	}

	db.AutoMigrate(&App{}, &Proxy{}, &DummyNode{}, &AppProxyDummyNode{})
	u1 := App{
		ClientToken:      "111",
		ServerToken:      "333",
		GatewayRules:     "80:80",
		ServerAddr:       "192.168.1.1:8888",
		ServerListenAddr: "0.0.0.0:8888",
		ClientListenIP:   "127.0.0.2",
	}
	if e := db.Create(&u1).Error; e != nil {
		panic(e)
	}
	mapping := AppProxyDummyNode{
		ClientToken:  "111",
		ProxyTag:     "tx1",
		DummyNodeTag: "dn1",
	}
	if e := db.Model(mapping).Create(&mapping).Error; e != nil {
		panic(e)
	}
	u2 := Proxy{
		ProxyTag:      "tx1",
		Network:       "tcp",
		Addr:          "192.192.192.192:111",
		ListenAddr:    "0.0.0.0:111",
		AppUpdateTime: time.Now()}
	if e := db.Create(&u2).Error; e != nil {
		panic(e)
	}
	u3 := DummyNode{
		DummyNodeTag: "dn1",
		Network:      "tcp",
		Addr:         "122.122.122.122"}
	if e := db.Create(&u3).Error; e != nil {
		panic(e)
	}
	defer db.Close() //延时调用函数

	r := gin.Default()
	r.GET("/client", index)               //获取所有用户
	r.GET("/client/:client_token", show1) //根据id获取用户
	_ = r.Run()
}

//获取所有用户
func index(c *gin.Context) {
	var users []App
	db.Find(&users)
	c.JSON(200, users)
}

//根据id获取用户
func show(c *gin.Context) {
	CT := c.Params.ByName("client_token")
	var app App
	err := db.Debug().Model(&app).
		Related("proxies").
		Related("dummy_nodes").
		Where("client_token = ?", CT).
		Find(&app).Error
	fmt.Println(err)
	if app.ClientToken == "" {
		c.JSON(404, gin.H{"message": "user not found"})
		return
	}
	c.JSON(200, app)
}

//保存新用户
func store(c *gin.Context) {
	var user App
	_ = c.BindJSON(&user) //绑定一个请求主体到一个类型
	db.Create(&user)
	c.JSON(200, user)
}

func show1(c *gin.Context) {
	CT := c.Params.ByName("client_token")
	var app App
	//get from app table
	if e := db.Model(App{}).Where("client_token=?", CT).First(&app).Error; e != nil {
		panic(e)
	}

	//get mapping related from middle table
	var mapping AppProxyDummyNode
	if e := db.Model(mapping).Where("client_token=?", CT).First(&mapping).Error; e != nil {
		panic(e)
	}

	//get record from proxy table
	if e := db.Model(Proxy{}).Where("proxy_tag=?", mapping.ProxyTag).Find(&app).Error; e != nil {
		panic(e)
	}

	//get record from dummy_node table
	if e := db.Model(DummyNode{}).Where("dummy_node_tag=?", mapping.DummyNodeTag).Find(&app).Error; e != nil {
		panic(e)
	}

	if app.ClientToken == "" {
		c.JSON(404, gin.H{"message": "user not found"})
		return
	}
	c.JSON(200, app)
}

func jsonOutput(param interface{}) string {
	rs, _ := json.MarshalIndent(param, "  ", "  ")
	return string(rs)
}
