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
	"fmt"
	"free5gc/lib/http_wrapper"
	"free5gc/lib/openapi"
	"free5gc/lib/openapi/models"
	"free5gc/src/amf/logger"
	"github.com/gin-gonic/gin"
	"net/http"
	"strings"

	"free5gc/src/amf/handler/amf_message"
)

// N1N2MessageTransfer - Namf_Communication N1N2 Message Transfer (UE Specific) service Operation
func N1N2MessageTransfer(c *gin.Context) {
	var request models.N1N2MessageTransferRequest

	request.JsonData = new(models.N1N2MessageTransferReqData)
	s := strings.Split(c.GetHeader("Content-Type"), ";")
	var err error
	switch s[0] {
	case "application/json":
		err = fmt.Errorf("N1 and N2 datas are both Empty in N1N2MessgeTransfer")
	case "multipart/related":
		err = c.ShouldBindWith(&request, openapi.MultipartRelatedBinding{})
	}

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
	req.Params["ueContextId"] = c.Params.ByName("ueContextId")
	req.Params["reqUri"] = c.Request.RequestURI

	handlerMsg := amf_message.NewHandlerMessage(amf_message.EventN1N2MessageTransfer, req)
	amf_message.SendMessage(handlerMsg)

	rsp := <-handlerMsg.ResponseChan

	HTTPResponse := rsp.HTTPResponse

	for key, val := range HTTPResponse.Header {
		c.Header(key, val[0])
	}
	c.JSON(HTTPResponse.Status, HTTPResponse.Body)
}

func N1N2MessageTransferStatus(c *gin.Context) {

	req := http_wrapper.NewRequest(c.Request, nil)
	req.Params["ueContextId"] = c.Params.ByName("ueContextId")
	req.Params["reqUri"] = c.Request.RequestURI
	handlerMsg := amf_message.NewHandlerMessage(amf_message.EventN1N2MessageTransferStatus, req)
	amf_message.SendMessage(handlerMsg)

	rsp := <-handlerMsg.ResponseChan

	HTTPResponse := rsp.HTTPResponse

	c.JSON(HTTPResponse.Status, HTTPResponse.Body)
}
