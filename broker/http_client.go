package broker

import (
	"github.com/gin-gonic/gin"
	restful "github.com/sestack/grf"
	"github.com/sestack/smq/model"
)

type clientView struct {
	Cluster *Cluster
}

// @Tags 客户端管理
// @Summary 客户端列表
// @Security ApiKeyAuth
// @accept application/json
// @Produce application/json
// @Success 200 {object} object "{"code": 0, "data": []}"
// @Router /api/v1/client/  [get]
func (self *clientView) List(c *gin.Context) {
	datas := []*model.Client{}
	clients, err := self.Cluster.GetClients()
	if err != nil {
		restful.ErrorData(c, err)
		return
	}
	for _, data := range clients {
		datas = append(datas, data)
	}

	restful.SuccessData(c, datas)

	return
}

// @Tags 客户端管理
// @Summary 查看客户端
// @Security ApiKeyAuth
// @accept application/json
// @Produce application/json
// @Param id path string true "客户端id"
// @Success 200 {object} object "{"code": 0, "data":{}}"
// @Router /api/v1/client/{id}/ [get]
func (self *clientView) Retrieve(c *gin.Context) {
	clientID := c.Param("id")
	client, err := self.Cluster.GetClient(clientID)
	if err != nil {
		restful.NotFound(c)
		return
	}

	restful.SuccessData(c, client)

	return
}
