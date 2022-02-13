package broker

import (
	"github.com/gin-gonic/gin"
	"github.com/go-basic/uuid"
	restful "github.com/sestack/grf"
	"github.com/sestack/smq/model"
	"time"
)

type aclView struct {
	Cluster *Cluster
}

// @Tags 授权管理
// @Summary 授权列表
// @Security ApiKeyAuth
// @accept application/json
// @Produce application/json
// @Success 200 {object} object "{"code": 0, "data": []}"
// @Router /api/v1/acl/  [get]
func (self *aclView) List(c *gin.Context) {
	datas := []*model.Acl{}
	acls, err := self.Cluster.GetAcls()
	if err != nil {
		restful.ErrorData(c, err)
		return
	}
	for _, data := range acls {
		datas = append(datas, data)
	}

	restful.SuccessData(c, datas)
	return
}

// @Tags 授权管理
// @Summary 查看授权
// @Security ApiKeyAuth
// @accept application/json
// @Produce application/json
// @Param id path string true "授权id"
// @Success 200 {object} object "{"code": 0, "data":{}}"
// @Router /api/v1/acl/{id}/ [get]
func (self *aclView) Retrieve(c *gin.Context) {
	aclID := c.Param("id")
	acl, err := self.Cluster.GetAcl(aclID)
	if err != nil {
		restful.NotFound(c)
		return
	}

	restful.SuccessData(c, acl)
	return
}

// @Tags 授权管理
// @Summary 添加授权
// @Security ApiKeyAuth
// @accept application/json
// @Produce application/json
// @Param user body model.Acl true "传入参数是struct,classify:pub|sub,type:ip|clientid|username,action:allow|deny"
// @Success 200 {object} object "{"code": 0, "message": "操作成功！"}"
// @Router /api/v1/acl/ [post]
func (self *aclView) Create(c *gin.Context) {
	var err error
	acl := model.Acl{}

	if err = c.ShouldBind(&acl); err != nil {
		restful.ErrorData(c, err)
		return
	}

	if acl.Classify != "pub" && acl.Classify != "sub" {
		restful.Error(c)
		return
	}

	if acl.Type != "ip" && acl.Type != "clientid" && acl.Type != "username" {
		restful.Error(c)
		return
	}

	if acl.Action != "allow" && acl.Action != "deny" {
		restful.Error(c)
		return
	}

	if acl.ID == "" {
		acl.ID = uuid.New()
	}

	if acl.Value == "" {
		acl.Value = "*"
	}

	if acl.Topic == "" {
		acl.Topic = "#"
	}

	if acl.CreateAt == 0 {
		acl.CreateAt = time.Now().Unix()
	}

	err = self.Cluster.StoreAddORUpdateAcl(&acl)
	if err != nil {
		restful.ErrorData(c, err)
		return
	}

	restful.Success(c)
}

// @Tags 授权管理
// @Summary 修改授权
// @Security ApiKeyAuth
// @accept application/json
// @Produce application/json
// @Param id path string true "授权id"
// @Param user body model.Acl true "传入参数是struct"
// @Success 200 {object} object "{"code": 0, "message": "操作成功！"}"
// @Router /api/v1/acl/{id}/ [put]
func (self *aclView) Update(c *gin.Context) {
	var err error
	acl := &model.Acl{}
	aclID := c.Param("id")

	acl, err = self.Cluster.GetAcl(aclID)
	if err != nil {
		restful.NotFound(c)
		return
	}

	if err = c.ShouldBind(acl); err != nil {
		restful.ErrorData(c, err)
		return
	}

	if acl.Classify != "pub" && acl.Classify != "sub" {
		restful.Error(c)
		return
	}

	if acl.Type != "ip" && acl.Type != "clientid" && acl.Type != "username" {
		restful.Error(c)
		return
	}

	if acl.Action != "allow" && acl.Action != "deny" {
		restful.Error(c)
		return
	}

	if acl.ID == "" {
		acl.ID = uuid.New()
	}

	if acl.Value == "" {
		acl.Value = "*"
	}

	if acl.Topic == "" {
		acl.Topic = "#"
	}

	err = self.Cluster.StoreAddORUpdateAcl(acl)
	if err != nil {
		restful.ErrorData(c, err)
		return
	}

	restful.Success(c)
}

// @Tags 授权管理
// @Summary 删除授权
// @Security ApiKeyAuth
// @accept application/json
// @Produce application/json
// @Param id path string true "授权id"
// @Success 200 {object} object "{"code": 0, "message": "操作成功！"}"
// @Router /api/v1/acl/{id}/ [delete]
func (self *aclView) Delete(c *gin.Context) {
	aclID := c.Param("id")
	_, err := self.Cluster.GetAcl(aclID)
	if err != nil {
		restful.NotFound(c)
		return
	}

	err = self.Cluster.StoreRemoveAcl(aclID)
	if err != nil {
		restful.ErrorData(c, err)
		return
	}

	restful.Success(c)
}
