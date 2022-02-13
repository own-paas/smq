package broker

import (
	"github.com/gin-gonic/gin"
	restful "github.com/sestack/grf"
	"github.com/sestack/smq/model"
)

type subscriptionView struct {
	Cluster *Cluster
}

// @Tags 订阅管理
// @Summary 订阅列表
// @Security ApiKeyAuth
// @accept application/json
// @Produce application/json
// @Success 200 {object} object "{"code": 0, "data": []}"
// @Router /api/v1/subscription/  [get]
func (self *subscriptionView) List(c *gin.Context) {
	datas := []*model.Subscription{}
	subscriptions, err := self.Cluster.GetSubscriptions()
	if err != nil {
		restful.ErrorData(c, err)
		return
	}

	for _, data := range subscriptions {
		datas = append(datas, data)
	}

	restful.SuccessData(c, datas)
	return
}

// @Tags 订阅管理
// @Summary 查看订阅
// @Security ApiKeyAuth
// @accept application/json
// @Produce application/json
// @Param id path string true "订阅id"
// @Success 200 {object} object "{"code": 0, "data":{}}"
// @Router /api/v1/subscription/{id}/ [get]
func (self *subscriptionView) Retrieve(c *gin.Context) {
	subscriptionID := c.Param("id")
	subscription, err := self.Cluster.GetSubscription(subscriptionID)
	if err != nil {
		restful.NotFound(c)
		return
	}

	restful.SuccessData(c, subscription)
	return
}
