/*
 * Namf_Communication
 *
 * AMF Communication Service
 *
 * API version: 1.0.0
 * Generated by: OpenAPI Generator (https://openapi-generator.tech)
 */

package communication

import (
	"free5gc/lib/http_wrapper"
	"free5gc/lib/openapi/models"
	"free5gc/src/amf/amf_handler/amf_message"
	"free5gc/src/amf/logger"
	"github.com/gin-gonic/gin"
	"net/http"
)

// AMFStatusChangeSubscribeModfy - Namf_Communication AMF Status Change Subscribe Modify service Operation
func AMFStatusChangeSubscribeModfy(c *gin.Context) {
	var request models.SubscriptionData

	err := c.ShouldBindJSON(&request)
	if err != nil {
		problemDetail := "[Request Body] " + err.Error()
		rsp := models.ProblemDetails{
			Title:  "Malformed request syntax",
			Status: http.StatusBadRequest,
			Detail: problemDetail,
		}
		logger.CommLog.Errorln(problemDetail)
		c.JSON(http.StatusBadRequest, rsp)
		return
	}

	req := http_wrapper.NewRequest(c.Request, request)
	req.Params["subscriptionId"] = c.Params.ByName("subscriptionId")

	handlerMsg := amf_message.NewHandlerMessage(amf_message.EventAMFStatusChangeSubscribeModfy, req)
	amf_message.SendMessage(handlerMsg)

	rsp := <-handlerMsg.ResponseChan

	HTTPResponse := rsp.HTTPResponse

	c.JSON(HTTPResponse.Status, HTTPResponse.Body)
}

// AMFStatusChangeUnSubscribe - Namf_Communication AMF Status Change UnSubscribe service Operation
func AMFStatusChangeUnSubscribe(c *gin.Context) {

	req := http_wrapper.NewRequest(c.Request, nil)
	req.Params["subscriptionId"] = c.Params.ByName("subscriptionId")

	handlerMsg := amf_message.NewHandlerMessage(amf_message.EventAMFStatusChangeUnSubscribe, req)
	amf_message.SendMessage(handlerMsg)

	rsp := <-handlerMsg.ResponseChan

	HTTPResponse := rsp.HTTPResponse

	c.JSON(HTTPResponse.Status, HTTPResponse.Body)

}