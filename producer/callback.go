package producer

import (
	"fmt"
	"free5gc/lib/http_wrapper"
	"free5gc/lib/nas/nasMessage"
	"free5gc/lib/ngap/ngapType"
	"free5gc/lib/openapi/models"
	"free5gc/src/amf/consumer"
	"free5gc/src/amf/context"
	gmm_message "free5gc/src/amf/gmm/message"
	amf_message "free5gc/src/amf/handler/message"
	"free5gc/src/amf/logger"
	"free5gc/src/amf/nas"
	"github.com/mohae/deepcopy"
	"net/http"
	"strconv"
)

func HandleSmContextStatusNotify(httpChannel chan amf_message.HandlerResponseMessage, guti, pduSessionIdString string, body models.SmContextStatusNotification) {
	var problem models.ProblemDetails
	amfSelf := context.AMF_Self()
	ue, _ := amfSelf.AmfUeFindByGuti(guti)

	logger.ProducerLog.Infoln("[AMF] Handle SmContext Status Notify")
	if ue == nil {
		problem.Status = 404
		problem.Cause = "CONTEXT_NOT_FOUND"
		problem.Detail = fmt.Sprintf("Guti[%s] Not Found", guti)
		amf_message.SendHttpResponseMessage(httpChannel, nil, http.StatusNotFound, problem)
		return
	}
	pduSessionID, _ := strconv.Atoi(pduSessionIdString)
	_, ok := ue.SmContextList[int32(pduSessionID)]
	if !ok {
		problem.Status = 404
		problem.Cause = "CONTEXT_NOT_FOUND"
		problem.Detail = fmt.Sprintf("PDUSessionID[%d] Not Found", pduSessionID)
		amf_message.SendHttpResponseMessage(httpChannel, nil, http.StatusNotFound, problem)
		return
	}
	logger.CallbackLog.Debugf("Release PDUSessionId[%d] of UE[%s] By SmContextStatus Notification because of %s", pduSessionID, ue.Supi, body.StatusInfo.Cause)
	pduSessionId := int32(pduSessionID)
	delete(ue.SmContextList, pduSessionId)
	amf_message.SendHttpResponseMessage(httpChannel, nil, http.StatusNoContent, nil)
	if storedSmContext, exist := ue.StoredSmContext[pduSessionId]; exist {

		smContextCreateData := consumer.BuildCreateSmContextRequest(ue, *storedSmContext.PduSessionContext, models.RequestType_INITIAL_REQUEST)

		response, smContextRef, errResponse, problemDetail, err := consumer.SendCreateSmContextRequest(ue, storedSmContext.SmfUri, storedSmContext.Payload, smContextCreateData)
		if response != nil {
			var smContext context.SmContext
			smContext.PduSessionContext = storedSmContext.PduSessionContext
			smContext.PduSessionContext.SmContextRef = smContextRef
			smContext.UserLocation = deepcopy.Copy(ue.Location).(models.UserLocation)
			smContext.SmfUri = storedSmContext.SmfUri
			smContext.SmfId = storedSmContext.SmfId
			ue.SmContextList[pduSessionId] = &smContext
			logger.CallbackLog.Infof("Http create smContext[pduSessionID: %d] Success", pduSessionId)
			// TODO: handle response(response N2SmInfo to RAN if exists)
		} else if errResponse != nil {
			logger.CallbackLog.Warnf("PDU Session Establishment Request is rejected by SMF[pduSessionId:%d]\n", pduSessionId)
			gmm_message.SendDLNASTransport(ue.RanUe[storedSmContext.AnType], nasMessage.PayloadContainerTypeN1SMInfo, errResponse.BinaryDataN1SmInfoToUe, &pduSessionId, 0, nil, 0)
		} else if err != nil {
			logger.CallbackLog.Errorf("Failed to Create smContext[pduSessionID: %d], Error[%s]\n", pduSessionID, err.Error())
		} else {
			logger.CallbackLog.Errorf("Failed to Create smContext[pduSessionID: %d], Error[%v]\n", pduSessionID, problemDetail)
		}
		delete(ue.StoredSmContext, pduSessionId)

	}
}

func HandleAmPolicyControlUpdateNotifyUpdate(request *http_wrapper.Request) (*context.AmfUe, *http_wrapper.Response) {
	logger.ProducerLog.Infoln("Handle AM Policy Control Update Notify [Policy update notification]")

	polAssoID := request.Params["polAssoId"]
	policyUpdate := request.Body.(models.PolicyUpdate)

	ue, problemDetails := AmPolicyControlUpdateNotifyUpdateProcedure(polAssoID, policyUpdate)

	if problemDetails != nil {
		return nil, http_wrapper.NewResponse(int(problemDetails.Status), nil, problemDetails)
	} else {
		return ue, http_wrapper.NewResponse(http.StatusNoContent, nil, nil)
	}
}

func AmPolicyControlUpdateNotifyUpdateProcedure(polAssoID string, policyUpdate models.PolicyUpdate) (ue *context.AmfUe, problemDetails *models.ProblemDetails) {
	amfSelf := context.AMF_Self()

	ue, ok := amfSelf.AmfUeFindByPolicyAssociationID(polAssoID)
	if !ok {
		problemDetails = &models.ProblemDetails{
			Status: http.StatusNotFound,
			Cause:  "CONTEXT_NOT_FOUND",
			Detail: fmt.Sprintf("Policy Association ID[%s] Not Found", polAssoID),
		}
		return
	}

	ue.AmPolicyAssociation.Triggers = policyUpdate.Triggers
	ue.RequestTriggerLocationChange = false

	for _, trigger := range policyUpdate.Triggers {
		if trigger == models.RequestTrigger_LOC_CH {
			ue.RequestTriggerLocationChange = true
		}
		if trigger == models.RequestTrigger_PRA_CH {
			// TODO: Presence Reporting Area handling (TS 23.503 6.1.2.5, TS 23.501 5.6.11)
		}
	}

	if policyUpdate.ServAreaRes != nil {
		ue.AmPolicyAssociation.ServAreaRes = policyUpdate.ServAreaRes
	}

	if policyUpdate.Rfsp != 0 {
		ue.AmPolicyAssociation.Rfsp = policyUpdate.Rfsp
	}

	return
}

func HandleAmPolicyControlUpdateNotifyTerminate(httpChannel chan amf_message.HandlerResponseMessage, polAssoId string, body models.TerminationNotification) {
	logger.ProducerLog.Infoln("Handle AM Policy Control Update Notify [Request for termination of the policy association]")

	var problem models.ProblemDetails
	amfSelf := context.AMF_Self()
	ue, ok := amfSelf.AmfUeFindByPolicyAssociationID(polAssoId)

	if !ok {
		problem.Status = 404
		problem.Cause = "CONTEXT_NOT_FOUND"
		problem.Detail = fmt.Sprintf("Policy Association ID[%s] Not Found", polAssoId)
		amf_message.SendHttpResponseMessage(httpChannel, nil, http.StatusNotFound, problem)
		return
	}

	logger.CallbackLog.Warnf("Cause of AM Policy termination[%+v]", body.Cause)

	amf_message.SendHttpResponseMessage(httpChannel, nil, http.StatusNoContent, nil)

	problemDetails, err := consumer.AMPolicyControlDelete(ue)
	if problemDetails != nil {
		logger.GmmLog.Errorf("AM Policy Control Delete Failed Problem[%+v]", problemDetails)
	} else if err != nil {
		logger.GmmLog.Errorf("AM Policy Control Delete Error[%v]", err.Error())
	}
}

// TS 23.502 4.2.2.2.3 Registration with AMF re-allocation
func HandleN1MessageNotify(httpChannel chan amf_message.HandlerResponseMessage, body models.N1MessageNotify) {
	logger.ProducerLog.Infoln("[AMF] Handle N1 Message Notify")

	logger.ProducerLog.Debugf("request body: %+v", body)

	amfSelf := context.AMF_Self()

	registrationCtxtContainer := body.JsonData.RegistrationCtxtContainer
	if registrationCtxtContainer.UeContext == nil {
		problem := models.ProblemDetails{
			Status: 400,
			Cause:  "MANDATORY_IE_MISSING", // Defined in TS 29.500 5.2.7.2
			Detail: "Missing IE [UeContext] in RegistrationCtxtContainer",
		}
		amf_message.SendHttpResponseMessage(httpChannel, nil, http.StatusBadRequest, problem)
		return
	}

	amf_message.SendHttpResponseMessage(httpChannel, nil, http.StatusNoContent, nil)

	var amfUe *context.AmfUe
	ueContext := registrationCtxtContainer.UeContext
	if ueContext.Supi != "" {
		amfUe = amfSelf.NewAmfUe(ueContext.Supi)
	} else {
		amfUe = amfSelf.NewAmfUe("")
	}
	amfUe.CopyDataFromUeContextModel(*ueContext)

	ran := amfSelf.RanIdPool[*registrationCtxtContainer.RanNodeId]
	if ran == nil {
		problem := models.ProblemDetails{
			Status: 400,
			Cause:  "MANDATORY_IE_INCORRECT",
			Detail: fmt.Sprintf("Can not find RAN[RanId: %+v]", *registrationCtxtContainer.RanNodeId),
		}
		amf_message.SendHttpResponseMessage(httpChannel, nil, http.StatusBadRequest, problem)
		return
	}

	ranUe := ran.RanUeFindByRanUeNgapID(int64(registrationCtxtContainer.AnN2ApId))

	ranUe.Location = *registrationCtxtContainer.UserLocation
	amfUe.Location = *registrationCtxtContainer.UserLocation
	ranUe.UeContextRequest = registrationCtxtContainer.UeContextRequest
	ranUe.OldAmfName = registrationCtxtContainer.InitialAmfName

	if registrationCtxtContainer.AllowedNssai != nil {
		allowedNssai := registrationCtxtContainer.AllowedNssai
		amfUe.AllowedNssai[allowedNssai.AccessType] = allowedNssai.AllowedSnssaiList
	}

	if len(registrationCtxtContainer.ConfiguredNssai) > 0 {
		amfUe.ConfiguredNssai = registrationCtxtContainer.ConfiguredNssai
	}

	amfUe.AttachRanUe(ranUe)

	nas.HandleNAS(ranUe, ngapType.ProcedureCodeInitialUEMessage, body.BinaryDataN1Message)
}
