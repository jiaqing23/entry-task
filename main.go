package main

import (
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

var (
	router = gin.Default()
)

type Login struct {
	Username string `form:"username" json:"username" xml:"username"  binding:"required"`
	Password string `form:"password" json:"password" xml:"password" binding:"required"`
}

var user = Login{
	Username: "Tan",
	Password: "p@SzWoRD",
}

func getLogin(c *gin.Context) {
	c.HTML(http.StatusOK, "login.html", nil)
}

func postLogin(c *gin.Context) {
	var loginInput Login
	if err := c.ShouldBind(&loginInput); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	fmt.Println(loginInput.Username)
	fmt.Println(loginInput.Password)
	c.Redirect(http.StatusFound, "/profile")
}

func getProfile(c *gin.Context) {
	c.HTML(http.StatusOK, "profile.html", gin.H{
		"username": user.Username,
		"password": user.Password,
		"image":    "image/123.png",
	})
}

func getEdit(c *gin.Context) {
	c.HTML(http.StatusOK, "edit.html", nil)
}

func postEdit(c *gin.Context) {
	username := c.PostForm("username")
	password := c.PostForm("password")

	if password == "" || len(strings.TrimSpace(username)) == 0 {
		c.String(http.StatusInternalServerError, "Require username and password")
		return
	}

	if image, err := c.FormFile("image"); err == nil {
		if err2 := c.SaveUploadedFile(image, "image/"+image.Filename); err2 != nil {
			c.String(http.StatusInternalServerError, err2.Error())
			return
		}
	}

	fmt.Println(username)
	c.String(http.StatusOK, "ok")
}

func main() {
	router.LoadHTMLGlob("html/*")
	router.Static("/image", "./image")
	router.MaxMultipartMemory = 10 << 20 // 10MB
	router.GET("/", getLogin)
	router.POST("/login", postLogin)
	router.GET("/profile", getProfile)
	router.GET("/edit", getEdit)
	router.POST("edit", postEdit)
	log.Fatal(router.Run(":8080"))
}
