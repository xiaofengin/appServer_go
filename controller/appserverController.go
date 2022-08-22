package controller

import (
	"AppServer/dao"
	"AppServer/model"
	"AppServer/setting"
	"bytes"
	"errors"
	"fmt"
	rtctokenbuilder "github.com/AgoraIO/Tools/DynamicKey/AgoraDynamicKey/go/src/RtcTokenBuilder"
	"github.com/AgoraIO/Tools/DynamicKey/AgoraDynamicKey/go/src/accesstoken2"
	"github.com/gin-gonic/gin"
	"github.com/goccy/go-json"
	"io/ioutil"
	"net/http"
	"strconv"
	"time"
)

type UserForm struct {
	Username string `form:"username" json:"userAccount" uri:"username" xml:"username" binding:"required"`
	Password string `form:"password" json:"userPassword" uri:"password" xml:"password" binding:"required"`
}
type RegisterReq struct {
	Username string `form:"username" json:"username"`
	Password string `form:"password" json:"password"`
}

var AppToken string = ""

func InitAppserver() {
	// 获取APPtoken
	appId := setting.Conf.AgoraAppId
	appCertificate := setting.Conf.AgoraCert
	expire := setting.Conf.AgoraExpire
	accessToken := accesstoken2.NewAccessToken(appId, appCertificate, expire)
	serviceChat := accesstoken2.NewServiceChat("")
	serviceChat.AddPrivilege(accesstoken2.PrivilegeChatApp, expire)
	accessToken.AddService(serviceChat)
	res, err := accessToken.Build()
	if err != nil {
	} else {
		AppToken = res
	}

}
func Register(c *gin.Context) {
	// 获取账号密码参数
	var form UserForm
	if err := c.Bind(&form); err != nil {
		c.JSON(201, gin.H{
			"code":   201,
			"result": err.Error(),
		})
		return
	}
	// 检测数据库是否有该用户名
	if checkIfUserAccountExistsDB(form.Username) {
		c.JSON(201, gin.H{"code": 201, "error": "userAccount " + form.Username + " already exists"})
		return
	}

	// 去AgroaChat那边注册用户
	agoraChatUserUuid, err := RegisterAgoraChatUser(form.Username)

	if err != nil {
		c.JSON(201, gin.H{"code": 201, "error": err.Error()})
		return
	}
	// 生成一个时间戳当成用户的AgoraUid（主要用于音视频通话）
	currentTimestamp := int(time.Now().UTC().UnixNano())
	// 构建model 存本地数据库
	appUserInfo := model.AppUserInfo{
		UserAccount:       form.Username,
		UserPassword:      form.Password,
		AgoraChatUserName: form.Username,
		AgoraChatUserUuid: agoraChatUserUuid,
		AgoraUid:          currentTimestamp,
	}
	err = dao.DB.Create(&appUserInfo).Error
	if err != nil {
		c.JSON(201, gin.H{"code": 201, "error": err.Error()})
	} else {
		c.JSON(http.StatusOK, gin.H{
			"code": "RES_OK",
			"msg":  "success",
		})
	}
}
func Login(c *gin.Context) {

	// 获取账号密码参数
	var form UserForm
	if err := c.Bind(&form); err != nil {
		c.JSON(201, gin.H{
			"code":   201,
			"result": err.Error(),
		})
		return
	}
	// 检测数据库是否有该用户名和对应的密码
	userModel := checkIfUserAndPasswordxistsDB(form.Username, form.Password)
	if len(userModel.UserAccount) == 0 {
		c.JSON(201, gin.H{"code": 201, "error": "userAccount or password error"})
		return
	}

	// 获取用户token
	appId := setting.Conf.AgoraAppId
	appCertificate := setting.Conf.AgoraCert
	expire := setting.Conf.AgoraExpire
	accessToken := accesstoken2.NewAccessToken(appId, appCertificate, expire)
	serviceChat := accesstoken2.NewServiceChat(form.Username)
	serviceChat.AddPrivilege(accesstoken2.PrivilegeChatUser, expire)
	accessToken.AddService(serviceChat)
	res, err := accessToken.Build()

	if err != nil {
		c.JSON(201, gin.H{"code": 201, "error": err.Error()})
	} else {
		c.JSON(http.StatusOK, gin.H{
			"code":             200,
			"accessToken":      res,
			"chatUserName":     userModel.UserAccount,
			"chatUserNickname": userModel.UserAccount,
			"agoraUid":         userModel.AgoraUid,
		})
	}
}
func RtcToken(c *gin.Context) {

	//获取路基参数
	channel := c.Param("channel")
	agorauid, _ := strconv.Atoi(c.Param("agorauid"))
	//userAccount := c.QueryArray("userAccount")
	// 填入项目 App ID
	appID := setting.Conf.AgoraAppId
	// 填入 App 证书
	appCertificate := setting.Conf.AgoraCert
	// Token 过期的时间，单位为秒
	expireTimeInSeconds := setting.Conf.AgoraExpire
	// 获取当前时间戳
	currentTimestamp := uint32(time.Now().UTC().Unix())
	// Token 过期的 Unix 时间戳
	expireTimestamp := currentTimestamp + expireTimeInSeconds

	// 获取音视频的token
	result, err := rtctokenbuilder.BuildTokenWithUID(appID, appCertificate, channel, uint32(agorauid), 1, expireTimestamp)
	if err != nil {
		c.JSON(201, gin.H{
			"code":  201,
			"error": err.Error(),
		})
	} else {
		//输出json结果给调用方
		c.JSON(http.StatusOK, gin.H{
			"code":            "RES_OK",
			"accessToken":     result,
			"expireTimestamp": expireTimestamp,
		})
	}

}
func checkIfUserAccountExistsDB(userAccount string) bool {
	var user model.AppUserInfo
	dao.DB.Where("user_account = ?", userAccount).First(&user)
	return len(user.UserAccount) > 0
}
func checkIfUserAndPasswordxistsDB(userAccount string, password string) model.AppUserInfo {
	var user model.AppUserInfo
	dao.DB.Where("user_account = ? AND user_password = ?", userAccount, password).First(&user)
	return user
}

func RegisterAgoraChatUser(chatUserName string) (string, error) {

	orgName := setting.Conf.OrgName
	appName := setting.Conf.AppName
	domain := setting.Conf.BaseUri

	loginInfo := &RegisterReq{
		Username: chatUserName,
		Password: "1",
	}
	info, err := json.Marshal(loginInfo)
	url := "http://" + domain + "/" + orgName + "/" + appName + "/users"
	resp, err := http.NewRequest("POST", url, bytes.NewReader(info))
	resp.Header.Add("Authorization", "Bearer "+AppToken)
	resp.Header.Add("Content-Type", "application/json")

	if err != nil {
		fmt.Println("get failed, err", err)
	}
	client := &http.Client{}
	res, err := client.Do(resp)
	if err != nil {
		fmt.Println("get failed, err", err)
	}

	//关闭Body
	defer resp.Body.Close()
	//读取body内容
	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		fmt.Println("read from resp.Body failed, err", err)
	}
	reqSucceed := ReqSucceed{}
	_ = json.Unmarshal(body, &reqSucceed)
	if len(reqSucceed.Path) == 0 {
		reqErrer := ReqErrer{}
		_ = json.Unmarshal(body, &reqErrer)
		return "", errors.New(reqErrer.ErrorDescription)
	} else {
		return reqSucceed.Entities[0].Uuid, nil
	}

}

type ReqErrer struct {
	Error            string `json:"error"`
	Exception        string `json:"exception"`
	Timestamp        int64  `json:"timestamp"`
	Duration         int    `json:"duration"`
	ErrorDescription string `json:"error_description"`
}
type ReqSucceed struct {
	Path         string `json:"path"`
	Uri          string `json:"uri"`
	Timestamp    int64  `json:"timestamp"`
	Organization string `json:"organization"`
	Application  string `json:"application"`
	Entities     []struct {
		Uuid      string `json:"uuid"`
		Type      string `json:"type"`
		Created   int64  `json:"created"`
		Modified  int64  `json:"modified"`
		Username  string `json:"username"`
		Activated bool   `json:"activated"`
	} `json:"entities"`
	Action          string        `json:"action"`
	Data            []interface{} `json:"data"`
	Duration        int           `json:"duration"`
	ApplicationName string        `json:"applicationName"`
}
