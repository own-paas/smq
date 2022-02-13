package broker

import (
	"github.com/gin-gonic/gin"
	restful "github.com/sestack/grf"
	"github.com/sestack/smq/model"
)

type nodeView struct {
	Cluster *Cluster
}

// @Tags 节点管理
// @Summary 节点列表
// @Security ApiKeyAuth
// @accept application/json
// @Produce application/json
// @Success 200 {object} object "{"code": 0, "data": []}"
// @Router /api/v1/node/  [get]
func (self *nodeView) List(c *gin.Context) {
	datas := []*model.Node{}
	nodes, err := self.Cluster.GetNodes()
	if err != nil {
		restful.ErrorData(c, err)
		return
	}
	for _, data := range nodes {
		datas = append(datas, data)
	}

	restful.SuccessData(c, datas)
	return
}

// @Tags 节点管理
// @Summary 查看节点
// @Security ApiKeyAuth
// @accept application/json
// @Produce application/json
// @Param id path string true "节点id"
// @Success 200 {object} object "{"code": 0, "data":{}}"
// @Router /api/v1/node/{id}/ [get]
func (self *nodeView) Retrieve(c *gin.Context) {
	nodeID := c.Param("id")
	node, err := self.Cluster.GetNode(nodeID)
	if err != nil {
		restful.NotFound(c)
		return
	}

	restful.SuccessData(c, node)
	return
}
