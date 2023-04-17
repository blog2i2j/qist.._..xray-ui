package controller

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/curve25519"
	"time"
	"xray-ui/web/global"
	"xray-ui/web/service"
)

type ServerController struct {
	BaseController

	serverService service.ServerService

	lastStatus        *service.Status
	lastGetStatusTime time.Time

	lastVersions        []string
	lastGetVersionsTime time.Time

	lastGeoipStatus        *service.Status
	lastGeoipGetStatusTime time.Time

	lastGeoipVersions        []string
	lastGeoipGetVersionsTime time.Time

	lastGeositeStatus        *service.Status
	lastGeositeGetStatusTime time.Time

	lastGeositeVersions        []string
	lastGeositeGetVersionsTime time.Time
	xraysecretkey        map[string]string
}

func NewServerController(g *gin.RouterGroup) *ServerController {
	a := &ServerController{
		lastGetStatusTime: time.Now(),
	}
	a.initRouter(g)
	a.startTask()
	return a
}

func (a *ServerController) initRouter(g *gin.RouterGroup) {
	g = g.Group("/server")

	g.Use(a.checkLogin)
	g.POST("/status", a.status)
	g.POST("/getXrayVersion", a.getXrayVersion)
	g.POST("/installXray/:version", a.installXray)
	g.POST("/getGeoipVersion", a.getGeoipVersion)
	g.POST("/installGeoip/:version", a.installGeoip)
	g.POST("/getGeositeVersion", a.getGeositeVersion)
	g.POST("/installGeosite/:version", a.installGeosite)
	g.POST("/xraysecretkey", a.XraySecretKey)
}

func (a *ServerController) refreshStatus() {
	a.lastStatus = a.serverService.GetStatus(a.lastStatus)
}

func (a *ServerController) startTask() {
	webServer := global.GetWebServer()
	c := webServer.GetCron()
	c.AddFunc("@every 2s", func() {
		now := time.Now()
		if now.Sub(a.lastGetStatusTime) > time.Minute*3 {
			return
		}
		a.refreshStatus()
	})
}

func (a *ServerController) status(c *gin.Context) {
	a.lastGetStatusTime = time.Now()

	jsonObj(c, a.lastStatus, nil)
}

func (a *ServerController) getXrayVersion(c *gin.Context) {
	now := time.Now()
	if now.Sub(a.lastGetVersionsTime) <= time.Minute {
		jsonObj(c, a.lastVersions, nil)
		return
	}

	versions, err := a.serverService.GetXrayVersions()
	if err != nil {
		jsonMsg(c, "获取版本", err)
		return
	}

	a.lastVersions = versions
	a.lastGetVersionsTime = time.Now()

	jsonObj(c, versions, nil)
}

func (a *ServerController) installXray(c *gin.Context) {
	version := c.Param("version")
	err := a.serverService.UpdateXray(version)
	jsonMsg(c, "安装 xray", err)
}

func (a *ServerController) getGeoipVersion(c *gin.Context) {
	now := time.Now()
	if now.Sub(a.lastGeoipGetVersionsTime) <= time.Minute {
		jsonObj(c, a.lastGeoipVersions, nil)
		return
	}

	versions, err := a.serverService.GetGeoipVersions()
	if err != nil {
		jsonMsg(c, "获取版本", err)
		return
	}

	a.lastGeoipVersions = versions
	a.lastGeoipGetVersionsTime = time.Now()

	jsonObj(c, versions, nil)
}

func (a *ServerController) installGeoip(c *gin.Context) {
	version := c.Param("version")
	err := a.serverService.UpdateGeoip(version)
	jsonMsg(c, "安装 geoip", err)
}

func (a *ServerController) getGeositeVersion(c *gin.Context) {
	now := time.Now()
	if now.Sub(a.lastGeositeGetVersionsTime) <= time.Minute {
		jsonObj(c, a.lastGeositeVersions, nil)
		return
	}

	versions, err := a.serverService.GetGeositeVersions()
	if err != nil {
		jsonMsg(c, "获取版本", err)
		return
	}

	a.lastGeositeVersions = versions
	a.lastGeositeGetVersionsTime = time.Now()

	jsonObj(c, versions, nil)
}

func (a *ServerController) installGeosite(c *gin.Context) {
	version := c.Param("version")
	err := a.serverService.UpdateGeosite(version)
	jsonMsg(c, "安装 geosite", err)
}

func (a *ServerController) XraySecretKey(c *gin.Context) {
	key := SecretKey()
	a.xraysecretkey = key
	jsonObj(c, a.xraysecretkey, nil)
	//fmt.Println("Private key:", privateKeyBase64)
	//fmt.Println("Public key:", publicKeyBase64)
}
func SecretKey() map[string]string  {
	// 生成私钥
	privateKey := make([]byte, curve25519.ScalarSize)
	if _, err := rand.Read(privateKey); err != nil {
		fmt.Println(err)
		return nil
	}

	// 调整私钥的位数
	privateKey[0] &= 248
	privateKey[31] &= 127
	privateKey[31] |= 64

	// 生成公钥
	publicKey, err := curve25519.X25519(privateKey, curve25519.Basepoint)
	if err != nil {
		fmt.Println(err)
		return nil
	}

	// 用 base64 编码密钥对
	privateKeyBase64 := base64.RawURLEncoding.EncodeToString(privateKey)
	publicKeyBase64 := base64.RawURLEncoding.EncodeToString(publicKey)
        secretkey := make(map[string]string)
	secretkey["key"] = privateKeyBase64
	secretkey["value"] = publicKeyBase64
	return secretkey
}
