package main

import (
	"fmt"
	"log"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/mysql"
)

// App 表
type App struct {
	ClientToken      string       `json:"client_token"`       // 客户端 token(客户端读取信息用)
	ServerToken      string       `json:"server_token"`       // 服务端 token(服务端读取信息用)
	GatewayRules     string       `json:"gateway_rules"`      // 监听端口,例如: 11:11;22:22;456:456-789
	ServerAddr       string       `json:"server_addr"`        // 服务器的外网地址，例如：192.168.1.1:8888(服务器真实 IP)
	ServerListenAddr string       `json:"server_listen_addr"` // 服务器的监听地址，例如：0.0.0.0:8888
	ClientListenIP   string       `json:"client_listen_ip"`   // 客户端的本地监听IP，例如：127.0.0.2
	Proxies          []Proxies    `json:"proxies"`
	DummyNodes       []DummyNodes `json:"dummy_nodes"`
}

// AppProxyDummyNode 中间件表  举例client_token为 111 的用户拥有 tx1 分组的 proxy 和 dn1 的 dummy_node
type AppProxyDummyNode struct {
	ClientToken  string `json:"client_token"`   // client_token
	ProxyTag     string `json:"proxy_tag"`      // proxy_tag
	DummyNodeTag string `json:"dummy_node_tag"` //dummy_node_tag
}

// Proxies 节点表
type Proxies struct {
	ProxyTag      string    `json:"proxy_tag"`
	Network       string    `json:"network"`         // 节点的类型，tcp/kcp
	Addr          string    `json:"addr"`            // 节点的外网地址
	ListenAddr    string    `json:"listen_addr"`     // 节点的监听地址
	AppUpdateTime time.Time `json:"app_update_time"` // Apps最后更新的时间
}

// DummyNodes 虚拟节点表
type DummyNodes struct {
	DummyNodeTag string `json:"dummy_node_tag"`
	Network      string `json:"network"` // 节点的类型，tcp/kcp/quic
	Addr         string `json:"addr"`    // 节点的外网地址
}

var db *gorm.DB
var err error

func main() {
	db, err = gorm.Open("mysql", "root:123456.ab@/api?charset=utf8&parseTime=True&loc=Local")
	if err != nil {
		log.Fatal("db connect error")
	}

	db.AutoMigrate(&App{}, &Proxies{}, &DummyNodes{}, &AppProxyDummyNode{})
	u1 := App{
		ClientToken:      "111",
		ServerToken:      "333",
		GatewayRules:     "80:80",
		ServerAddr:       "192.168.1.1:8888",
		ServerListenAddr: "0.0.0.0:8888",
		ClientListenIP:   "127.0.0.2"}
	db.Create(&u1)
	u2 := Proxies{
		ProxyTag:      "tx1",
		Network:       "tcp",
		Addr:          "192.192.192.192:111",
		ListenAddr:    "0.0.0.0:111",
		AppUpdateTime: time.Now()}
	u21 := Proxies{
		ProxyTag:   "tx1",
		Network:    "tcp",
		Addr:       "192.191.191.191:222",
		ListenAddr: "0.0.0.0:222",
	}
	db.Create(&u21)
	db.Create(&u2)
	u3 := DummyNodes{
		DummyNodeTag: "dn1",
		Network:      "tcp",
		Addr:         "122.122.122.122"}
	db.Create(&u3)
	u4 := AppProxyDummyNode{
		ClientToken:  "111",
		ProxyTag:     "tx1",
		DummyNodeTag: "dn1",
	}
	db.Create(&u4)

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
		Preload("proxies").
		Preload("dummy_nodes").
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
	err := db.Debug().Table("proxies").
		Where("client_token = ?", CT).
		Find(&app).Error
	db.Model(&app).Association("client_token").Find(&app)
	fmt.Println(err)
	if app.ClientToken == "" {
		c.JSON(404, gin.H{"message": "user not found"})
		return
	}
	c.JSON(200, app)
}
