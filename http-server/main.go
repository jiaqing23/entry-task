package main

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"os"
	"path"
	"strings"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/fatih/pool"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"github.com/twinj/uuid"
	"google.golang.org/protobuf/proto"

	"github.com/go-redis/redis/v7"
)

const (
	connHost = "localhost"
	connPort = "8080"
	connType = "tcp"
)

var (
	router      = gin.Default()
	redisClient *redis.Client
	tcpPool     pool.Pool
	ROOT        = os.Getenv("ENTRY_TASK_DEPLOY_PATH")
)

type Login struct {
	Username string `form:"username" json:"username" xml:"username"  binding:"required"`
	Password string `form:"password" json:"password" xml:"password" binding:"required"`
}

type User struct {
	Username string `form:"username" json:"username" xml:"username"`
	Nickname string `form:"nickname" json:"nickname" xml:"nickname"`
}

func init() {
	//Initializing redis
	dsn := os.Getenv("REDIS_DSN")
	if len(dsn) == 0 {
		dsn = "localhost:6379"
	}
	redisClient = redis.NewClient(&redis.Options{
		Addr: dsn,
	})
	_, err := redisClient.Ping().Result()
	if err != nil {
		log.Fatal(err)
	}
}

func main() {
	if err := godotenv.Load(path.Join(os.Getenv("ENTRY_TASK_DEPLOY_PATH"), ".env")); err != nil {
		log.Fatalf("Error loading .env file")
	}

	var err error
	tcpPool, err = pool.NewChannelPool(10, 15, connCreator)
	if err != nil {
		log.Fatal("Error connecting: ", err)
	}
	defer tcpPool.Close()

	router.LoadHTMLGlob(path.Join(ROOT, "html/*"))
	// router.Static("/image", "./image")
	router.MaxMultipartMemory = 10 << 20 // 10MB
	router.GET("/", getLogin)
	router.POST("/login", postLogin)
	router.GET("/profile", TokenAuthMiddleware(), getProfile)
	router.GET("/edit", TokenAuthMiddleware(), getEdit)
	router.POST("/edit", TokenAuthMiddleware(), postEdit)
	router.GET("/logout", TokenAuthMiddleware(), logout)
	router.GET("/image/:imageName", showPic)
	log.Fatal(router.Run(":3000"))
}

func connCreator() (net.Conn, error) {
	return net.Dial(connType, connHost+":"+connPort)
}

func showPic(c *gin.Context) {
	file, err := os.Open(path.Join(ROOT, "image", c.Param("imageName")+".png"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	defer file.Close()

	buff, err := ioutil.ReadAll(file)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.Data(http.StatusOK, "image/png", buff)
}

func logout(c *gin.Context) {
	temp, exist := c.Get("accessUuid")
	var accessUuid string = temp.(string)
	if !exist {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Unexpected error when retrieving accessUuid info from middleware",
		})
		return
	}

	redisClient.Del(accessUuid)

	c.Redirect(http.StatusFound, "/")
}

func getLogin(c *gin.Context) {

	if _, _, err := VerifyToken(c); err != nil {

		c.HTML(http.StatusOK, "login.html", nil)
		return
	} else {
		c.Redirect(http.StatusFound, "/profile")
		return
	}
}

func postLogin(c *gin.Context) {
	var loginInput Login
	if err := c.ShouldBind(&loginInput); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	req := &Request{
		Username:    loginInput.Username,
		Password:    loginInput.Password,
		RequestType: Request_AUTH,
	}
	res, err := sendTCP(req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if res.ResponseType != Request_SUCCESS {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Username or password wrong."})
		return
	}

	token, err := CreateToken(&User{
		Username: res.Username,
		Nickname: res.Nickname,
	})
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.SetCookie("Authorization", token.AccessToken, 3600, "/", "localhost", false, true)
	c.Redirect(http.StatusFound, "/profile")
}

func getProfile(c *gin.Context) {
	temp, exist := c.Get("user")
	var user *User = temp.(*User)
	if !exist {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Unexpected error when retrieving user info from middleware",
		})
		return
	}

	c.HTML(http.StatusOK, "profile.html", gin.H{
		"username": user.Username,
		"nickname": user.Nickname,
		"image":    "image/" + user.Username,
	})
}

func getEdit(c *gin.Context) {
	c.HTML(http.StatusOK, "edit.html", nil)
}

func postEdit(c *gin.Context) {
	temp, exist := c.Get("user")
	var user *User = temp.(*User)
	if !exist {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Unexpected error when retrieving user info from middleware",
		})
		return
	}

	nickname := c.PostForm("nickname")
	if len(strings.TrimSpace(nickname)) == 0 {
		c.String(http.StatusInternalServerError, "Require nickname")
		return
	}

	if image, err := c.FormFile("image"); err == nil {
		if err2 := c.SaveUploadedFile(image, path.Join(ROOT, "image", user.Username+".png")); err2 != nil {
			c.String(http.StatusInternalServerError, err2.Error())
			return
		}
	}

	req := &Request{
		Username:    user.Username,
		Nickname:    nickname,
		RequestType: Request_EDIT_NICKNAME,
	}
	res, err := sendTCP(req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if res.ResponseType != Request_SUCCESS {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Request fail"})
		return
	}

	err = redisClient.Set(res.Username, res.Nickname, time.Hour*1).Err()
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err})
	}

	c.Redirect(http.StatusFound, "/profile")
}

func sendTCP(req *Request) (*Request, error) {
	conn, err := tcpPool.Get()
	if err != nil {
		log.Print("Error get connection from pool: ", err)
		return nil, err
	}
	defer conn.Close()

	out, err := proto.Marshal(req)
	if err != nil {
		log.Print("Failed to parse request: ", err)
		return nil, err
	}
	//log.Print(append(out, []byte{0}...))
	conn.Write(append(out, []byte{0}...))

	message, _ := bufio.NewReader(conn).ReadBytes(0)
	res := &Request{}
	if err := proto.Unmarshal(message[:len(message)-1], res); err != nil {
		log.Print("Failed to parse response: ", err)
		return nil, err
	}

	//log.Print("Server response: ", res)
	return res, nil
}

func VerifyToken(c *gin.Context) (*User, string, error) {
	tokenString, err := c.Cookie("Authorization")
	if err != nil {
		tokenString = ""
	}

	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		//Make sure that the token method conform to "SigningMethodHMAC"
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(os.Getenv("ACCESS_SECRET")), nil
	})
	if err != nil {
		return nil, "", err
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok || !token.Valid {
		return nil, "", err
	}
	accessUuid, ok := claims["access_uuid"].(string)
	if !ok {
		return nil, accessUuid, err
	}
	username, err := redisClient.Get(accessUuid).Result()
	if err != nil {
		return nil, accessUuid, err
	}
	nickname, err := redisClient.Get(username).Result()
	if err != nil {
		return nil, accessUuid, err
	}
	return &User{
		Username: username,
		Nickname: nickname,
	}, accessUuid, nil
}

type TokenDetails struct {
	AccessToken string
	AccessUuid  string
}

func CreateToken(user *User) (*TokenDetails, error) {
	td := &TokenDetails{}
	td.AccessUuid = uuid.NewV4().String()

	var err error
	atClaims := jwt.MapClaims{}
	atClaims["access_uuid"] = td.AccessUuid
	atClaims["username"] = user.Username
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, atClaims)
	td.AccessToken, err = token.SignedString([]byte(os.Getenv("ACCESS_SECRET")))
	if err != nil {
		return nil, err
	}

	expires := time.Unix(time.Now().Add(time.Minute*60).Unix(), 0) //converting Unix to UTC(to Time object)
	now := time.Now()
	log.Print(expires.Sub(now))
	err = redisClient.Set(td.AccessUuid, user.Username, expires.Sub(now)).Err()
	if err != nil {
		return nil, err
	}
	err = redisClient.Set(user.Username, user.Nickname, expires.Sub(now)).Err()
	if err != nil {
		return nil, err
	}
	return td, nil
}

func TokenAuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		user, accessUuid, err := VerifyToken(c)
		if err != nil {
			c.JSON(http.StatusUnauthorized, "Unauthorized")
			c.Abort()
			return
		}
		c.Set("user", user)
		c.Set("accessUuid", accessUuid)
		c.Next()
	}
}
