/*
 * Namf_MT
 *
 * AMF Mobile Termination Service
 *
 * API version: 1.0.0
 * Generated by: OpenAPI Generator (https://openapi-generator.tech)
 */

package Namf_MT

import (
	"free5gc/lib/http_wrapper"
	"free5gc/src/amf/amf_handler/amf_message"
	"github.com/gin-gonic/gin"
)

// ProvideDomainSelectionInfo - Namf_MT Provide Domain Selection Info service Operation
func ProvideDomainSelectionInfo(c *gin.Context) {

	req := http_wrapper.NewRequest(c.Request, nil)
	req.Params["ueContextId"] = c.Params.ByName("ueContextId")

	handlerMsg := amf_message.NewHandlerMessage(amf_message.EventProvideDomainSelectionInfo, req)
	amf_message.SendMessage(handlerMsg)

	rsp := <-handlerMsg.ResponseChan

	HTTPResponse := rsp.HTTPResponse

	c.JSON(HTTPResponse.Status, HTTPResponse.Body)

}
