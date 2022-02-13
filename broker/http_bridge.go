package broker

import (
	"github.com/gin-gonic/gin"
	restful "github.com/sestack/grf"
)

type bridgeView struct {
	Cluster *Cluster
}

// @Tags 桥接管理
// @Summary 桥接列表
// @Security ApiKeyAuth
// @accept application/json
// @Produce application/json
// @Success 200 {object} object "{"code": 0, "data": []}"
// @Router /api/v1/bridge/  [get]
func (self *bridgeView) List(c *gin.Context) {
	results := []map[string]interface{}{}
	bridges, err := self.Cluster.GetBridges()
	if err != nil {
		restful.ErrorData(c, err)
		return
	}

	for k, v := range bridges {
		results = append(results, map[string]interface{}{
			"topic": k,
			"queue": v,
		})
	}

	restful.SuccessData(c, results)
	return
}

// @Tags 桥接管理
// @Summary 查看桥接
// @Security ApiKeyAuth
// @accept application/json
// @Produce application/json
// @Param topic path string true "桥接topic"
// @Success 200 {object} object "{"code": 0, "data":{}}"
// @Router /api/v1/bridge/{topic}/ [get]
func (self *bridgeView) Retrieve(c *gin.Context) {
	topicName := c.Param("topic")
	data, err := self.Cluster.GetBridge(topicName)
	if err != nil {
		restful.NotFound(c)
		return
	}
	result := map[string]interface{}{
		"topic": topicName,
		"queue": data,
	}
	restful.SuccessData(c, result)
	return
}

// @Tags 桥接管理
// @Summary 添加桥接
// @Security ApiKeyAuth
// @accept application/json
// @Produce application/json
// @Param user body   model.Bridge  true "传入参数是map"
// @Success 200 {object} object "{"code": 0, "message": "操作成功！"}"
// @Router /api/v1/bridge/ [post]
func (self *bridgeView) Create(c *gin.Context) {
	var err error
	bridge := map[string]string{}

	if err = c.ShouldBind(&bridge); err != nil {
		restful.ErrorData(c, err)
		return
	}

	topic, ok := bridge["topic"]
	if !ok {
		restful.ErrorData(c, "failed to parse data")
		return
	}

	queue, ok := bridge["queue"]
	if !ok {
		restful.ErrorData(c, "failed to parse data")
		return
	}

	if topic == "" || queue == "" {
		restful.ErrorData(c, "failed to parse data")
		return
	}

	err = self.Cluster.StoreAddORUpdateBridge(topic, queue)
	if err != nil {
		restful.ErrorData(c, err)
		return
	}

	restful.Success(c)
}

// @Tags 桥接管理
// @Summary 修改交接
// @Security ApiKeyAuth
// @accept application/json
// @Produce application/json
// @Param topic path string true "桥接topic"
// @Param user body  model.Bridge true "传入参数是map"
// @Success 200 {object} object "{"code": 0, "message": "操作成功！"}"
// @Router /api/v1/bridge/{topic}/ [put]
func (self *bridgeView) Update(c *gin.Context) {
	var err error
	topicName := c.Param("topic")
	_, err = self.Cluster.GetBridge(topicName)
	if err != nil {
		restful.NotFound(c)
		return
	}


	bridge := map[string]string{}

	if err = c.ShouldBind(&bridge); err != nil {
		restful.ErrorData(c, err)
		return
	}


	queue, ok := bridge["queue"]
	if !ok {
		restful.ErrorData(c, "failed to parse data")
		return
	}

	if  queue == "" {
		restful.ErrorData(c, "failed to parse data")
		return
	}

	err = self.Cluster.StoreAddORUpdateBridge(topicName, queue)
	if err != nil {
		restful.ErrorData(c, err)
		return
	}

	restful.Success(c)
}

// @Tags 桥接管理
// @Summary 删除桥接
// @Security ApiKeyAuth
// @accept application/json
// @Produce application/json
// @Param topic path string true "桥接topic"
// @Success 200 {object} object "{"code": 0, "message": "操作成功！"}"
// @Router /api/v1/bridge/{topic}/ [delete]
func (self *bridgeView) Delete(c *gin.Context) {
	var err error
	topicName := c.Param("topic")
	_, err = self.Cluster.GetBridge(topicName)
	if err != nil {
		restful.NotFound(c)
		return
	}

	err = self.Cluster.StoreRemoveBridge(topicName)
	if err != nil {
		restful.ErrorData(c, err)
		return
	}

	restful.Success(c)
}
