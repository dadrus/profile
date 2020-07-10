package main

import (
	"errors"
	"fmt"
	authn "github.com/dadrus/gin-authn"
	"github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"
	"net/http"
	"os"
	"profile/model"
	"strconv"
)

var router *gin.Engine

var port = getEnv("PORT", "8090")
var login_url = getEnv("LOGIN_URL", "http://127.0.0.1:8081/login")
var own_url = getEnv("OWN_URL", "http://127.0.0.1:8091")
var main_url = getEnv("MAIN_URL", "http://127.0.0.1:8081")

func getEnv(key string, defaultValue string) string {
	if val, present := os.LookupEnv(key); present {
		return val
	} else {
		return defaultValue
	}
}

func main() {
	router = gin.Default()
	router.LoadHTMLGlob("web/templates/*")
	router.Use(authn.OAuth2Aware())

	initRoutes()

	if tlsConfig, err := getTlsConfig(); err == nil {
		fmt.Println("Starting with TLS mode")
		err = router.RunTLS(":"+port, tlsConfig.TlsCertFile, tlsConfig.TlsKeyFile)
		fmt.Println("Error: " + err.Error())
	} else {
		fmt.Println("Could not load TLS settings because of " + err.Error())
		err = router.Run(":" + port)
		fmt.Println("Error: " + err.Error())
	}
}

type tlsConfig struct {
	TlsKeyFile  string
	TlsCertFile string
}

func getTlsConfig() (*tlsConfig, error) {
	tlsKeyFile := os.Getenv("TLS_KEY")
	if len(tlsKeyFile) == 0 {
		return nil, errors.New("no TLS key configured")
	}
	if _, err := os.Stat(tlsKeyFile); err != nil {
		return nil, errors.New("configured TLS key not available")
	}

	tlsCertFile := os.Getenv("TLS_CERT")
	if len(tlsCertFile) == 0 {
		return nil, errors.New("no TLS cert configured")
	}
	if _, err := os.Stat(tlsCertFile); err != nil {
		return nil, errors.New("configured TLS cert not available")
	}

	return &tlsConfig{
		TlsKeyFile:  tlsKeyFile,
		TlsCertFile: tlsCertFile,
	}, nil
}

func initRoutes() {
	router.GET("/register", ShowRegisterPage)
	router.POST("/register", Register)
	router.GET("/profile/:id", authn.ScopesAllowed("profile"), GetProfile)
	router.POST("/profile/:id", authn.ScopesAllowed("profile"), UpdateProfile)
	router.PUT("/profile/:id", UpdateProfile)
	router.POST("/authenticate", AuthenticateUser)
}

type AuthenticationRequest struct {
	UserName string `form:"user_name" json:"user_name"`
	Password string `form:"password" json:"password"`
}

func AuthenticateUser(c *gin.Context) {
	var authRequest AuthenticationRequest
	if err := c.ShouldBind(&authRequest); err != nil {
		c.JSON(http.StatusBadRequest, nil)
		return
	}

	customer, err := model.FindCustomerByUserName(authRequest.UserName)
	if err != nil {
		c.JSON(http.StatusUnauthorized, nil)
		return
	}

	if customer.Password != authRequest.Password {
		c.JSON(http.StatusUnauthorized, nil)
		return
	}

	c.JSON(http.StatusOK, struct {
		User       *model.Customer `json:"user"`
		ProfileUrl string          `json:"profile_url"`
	}{User: customer, ProfileUrl: own_url + "/profile/" + strconv.Itoa(customer.ID)})
}

type registerForm struct {
	Email            string `form:"email" binding="required"`
	Password         string `form:"password" binding="required"`
	RepeatedPassword string `form:"repeated_password" binding="required"`
}

func render(c *gin.Context, code int, templateName string, data gin.H, key string) {
	switch c.Request.Header.Get("Accept") {
	case "application/json":
		// Respond with JSON
		c.JSON(code, data[key])
	case "application/xml":
		// Respond with XML
		c.XML(code, data[key])
	default:
		// Respond with HTML
		c.HTML(code, templateName, data)
	}
}

func ShowRegisterPage(c *gin.Context) {
	c.HTML(http.StatusOK, "register.html", gin.H{
		"title":     "Register",
		"login_url": login_url,
	})
}

func Register(c *gin.Context) {
	var registerData registerForm
	if err := c.ShouldBind(&registerData); err != nil {
		c.HTML(http.StatusBadRequest, "register.html", gin.H{"title": "Register"})
		return
	}

	if registerData.Password != registerData.RepeatedPassword {
		c.HTML(http.StatusBadRequest, "register.html", gin.H{
			"title": "Register",
			"error": "Password mismatch",
			"email": registerData.Email,
		})
		return
	}

	if model.CustomerExistsForEmail(registerData.Email) {
		c.HTML(http.StatusBadRequest, "register.html", gin.H{
			"title": "Register",
			"error": "You have already an account",
		})
		return
	}

	model.NewCustomer(registerData.Email, registerData.Password)
	c.Redirect(http.StatusSeeOther, main_url + "/login")
}

func GetProfile(c *gin.Context) {
	var userClaims map[string]interface{}
	if token, present := c.Get("id_token"); present {
		userClaims = token.(*jwt.Token).Claims.(jwt.MapClaims)
	}

	strId := c.Param("id")
	id, err := strconv.Atoi(strId)
	if err != nil {
		c.Redirect(http.StatusSeeOther, "/register")
		return
	}

	customer, err := model.FindCustomerById(id)
	if err != nil {
		c.Redirect(http.StatusSeeOther, "/register")
		return
	}

	render(c, http.StatusOK, "profile.html", gin.H{
		"title":    "Profile",
		"user": userClaims,
		"customer": customer,
		"home": main_url,
	}, "customer")
}

func UpdateProfile(c *gin.Context) {
	strId := c.Param("id")
	id, err := strconv.Atoi(strId)
	if err != nil {
		render(c, http.StatusBadRequest, "register.html", gin.H{"title": "Register"}, "no")
		return
	}

	customer, err := model.FindCustomerById(id)
	if err != nil {
		render(c, http.StatusNotFound, "register.html", gin.H{"title": "Register"}, "no")
		return
	}

	if err := c.ShouldBind(customer); err != nil {
		render(c, http.StatusBadRequest, "register.html", gin.H{"title": "Register"}, "no")
		return
	}

	if customer.Birthday != nil && customer.Birthday.IsZero() {
		customer.Birthday = nil
	}

	c.Redirect(http.StatusSeeOther, "/profile/"+strconv.Itoa(customer.ID))
}
