package broker

import (
	"github.com/gin-gonic/gin"
	restful "github.com/sestack/grf"
	"github.com/sestack/smq/global"
	"github.com/sestack/smq/model"
	"github.com/sestack/smq/utils"
	"github.com/sestack/smq/utils/jwt"
	"time"
)

type userView struct {
	Cluster *Cluster
}

// @Tags 用户管理
// @Summary 用户列表
// @Security ApiKeyAuth
// @accept application/json
// @Produce application/json
// @Success 200 {object} object "{"code": 0, "data": []}"
// @Router /api/v1/user/  [get]
func (self *userView) List(c *gin.Context) {
	results := []map[string]interface{}{}
	users, err := self.Cluster.GetUsers()
	if err != nil {
		restful.ErrorData(c, err)
		return
	}
	for _, data := range users {
		results = append(results, map[string]interface{}{
			"client_id": data.ClientID,
			"name":      data.Name,
			"role":      data.Role,
			"create_at": data.CreateAt,
		})
	}

	restful.SuccessData(c, results)
	return
}

// @Tags 用户管理
// @Summary 查看用户
// @Security ApiKeyAuth
// @accept application/json
// @Produce application/json
// @Param name path string true "用户名"
// @Success 200 {object} object "{"code": 0, "data":{}}"
// @Router /api/v1/user/{name}/ [get]
func (self *userView) Retrieve(c *gin.Context) {
	userName := c.Param("name")
	user, err := self.Cluster.GetUser(userName)
	if err != nil {
		restful.NotFound(c)
		return
	}

	result := map[string]interface{}{
		"client_id": user.ClientID,
		"name":      user.Name,
		"role":      user.Role,
		"create_at": user.CreateAt,
	}

	restful.SuccessData(c, result)
}

// @Tags 用户管理
// @Summary 添加用户
// @Security ApiKeyAuth
// @accept application/json
// @Produce application/json
// @Param user body model.User true "传入参数是struct"
// @Success 200 {object} object "{"code": 0, "message": "操作成功！"}"
// @Router /api/v1/user/ [post]
func (self *userView) Create(c *gin.Context) {
	var err error
	user := model.User{}

	if err = c.ShouldBind(&user); err != nil {
		restful.ErrorData(c, err)
		return
	}

	if user.Name == "" || user.Password == "" {
		restful.Error(c)
		return
	}

	if user.CreateAt == 0 {
		user.CreateAt = time.Now().Unix()
	}

	pwd, err := utils.HashSaltPassword(user.Password)
	if err != nil {
		restful.ErrorData(c, err)
		return
	}

	user.Password = pwd
	err = self.Cluster.StoreAddORUpdateUser(&user)
	if err != nil {
		restful.ErrorData(c, err)
		return
	}

	restful.Success(c)
}

// @Tags 用户管理
// @Summary 修改用户
// @Security ApiKeyAuth
// @accept application/json
// @Produce application/json
// @Param name path string true "用户名"
// @Param user body model.User true "传入参数是struct"
// @Success 200 {object} object "{"code": 0, "message": "操作成功！"}"
// @Router /api/v1/user/{name}/ [put]
func (self *userView) Update(c *gin.Context) {
	var err error
	newUser := model.User{}
	userName := c.Param("name")

	if err = c.ShouldBind(&newUser); err != nil {
		restful.ErrorData(c, err)
		return
	}

	oldUser, err := self.Cluster.GetUser(userName)
	if err != nil {
		restful.ErrorData(c, err)
		return
	}

	oldUser.Name = newUser.Name
	oldUser.ClientID = newUser.ClientID
	oldUser.Role = newUser.Role
	oldUser.CreateAt = newUser.CreateAt

	err = self.Cluster.StoreAddORUpdateUser(oldUser)
	if err != nil {
		restful.ErrorData(c, err)
		return
	}

	restful.Success(c)
}

// @Tags 用户管理
// @Summary 删除用户
// @Security ApiKeyAuth
// @accept application/json
// @Produce application/json
// @Param name path string true "用户名"
// @Success 200 {object} object "{"code": 0, "message": "操作成功！"}"
// @Router /api/v1/user/{name}/ [delete]
func (self *userView) Delete(c *gin.Context) {
	userName := c.Param("name")
	_, err := self.Cluster.GetUser(userName)
	if err != nil {
		restful.NotFound(c)
		return
	}

	err = self.Cluster.StoreRemoveUser(userName)
	if err != nil {
		restful.NotFound(c)
		return
	}

	restful.Success(c)
}

// @Tags 认证登陆
// @Summary 登录
// @accept application/json
// @Produce application/json
// @Param user body model.LoginForm true "传入参数是struct"
// @Success 200 {object} object "{"code": 0, "data":{"token": ""}}"
// @Router /api/v1/login/ [post]
func (self *userView) Login(c *gin.Context) {
	var err error
	params := model.LoginForm{}
	if err = c.ShouldBind(&params); err != nil {
		restful.ErrorData(c, err)
		return
	}

	if params.Username == "" || params.Password == "" {
		restful.ErrorData(c, "username and password cannot be empty")
		return
	}

	user, err := self.Cluster.GetUser(params.Username)
	if err != nil {
		restful.ErrorData(c, err)
		return
	}

	ok := utils.ComparePassword(user.Password, params.Password)
	if !ok {
		restful.ErrorData(c, "wrong user name or password")
		return
	}

	if user.Role != 1 {
		restful.ErrorData(c, "insufficient permissions, please contact the administrator")
		return
	}

	jwtObj := jwt.NewJWT(global.CONFIG.JWT)
	token, err := jwtObj.CreateToken(jwtObj.CreateClaims(jwt.JWTBaseClaims{Username: user.Name, Role: user.Role}))
	if err != nil {
		restful.ErrorData(c, err)
		return
	}

	restful.SuccessData(c, gin.H{"token": token})
}

// @Tags 用户管理
// @Summary 修改密码
// @Security ApiKeyAuth
// @accept application/json
// @Produce application/json
// @Param user body model.LoginForm true "传入参数是struct"
// @Success 200 {object} object "{"code": 0, "data":{"token": ""}}"
// @Router /api/v1/change_password/ [post]
func (self *userView) ChangePassword(c *gin.Context) {
	var err error
	var params model.LoginForm
	if err = c.ShouldBind(&params); err != nil {
		restful.ErrorData(c, err)
		return
	}

	if params.Username == "" || params.Password == "" {
		restful.ErrorData(c, "username and password cannot be empty")
		return
	}

	user, err := self.Cluster.GetUser(params.Username)
	if err != nil {
		restful.ErrorData(c, err)
		return
	}

	pwd, err := utils.HashSaltPassword(params.Password)
	if err != nil {
		restful.ErrorData(c, err)
		return
	}

	user.Password = pwd
	err = self.Cluster.StoreAddORUpdateUser(user)
	if err != nil {
		restful.ErrorData(c, err)
		return
	}

	restful.Success(c)
}
