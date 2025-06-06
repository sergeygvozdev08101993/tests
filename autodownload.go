package handlers

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"gitlab.nodasoft.com/prices/plmiddlewareapi/models"

	"github.com/gorilla/mux"
	log "github.com/sirupsen/logrus"
	"gitlab.nodasoft.com/prices/plmiddlewareapi/models/autodownload"
)

// PriceAutoDownloadAPIArg contains connection string to price autodownload api.
var PriceAutoDownloadAPIArg string

// GetAutoDownloadQueue handles requests to return tasks list from price autodownload api.
func GetAutoDownloadQueue(w http.ResponseWriter, r *http.Request) {

	var (
		tasks autodownload.LogResponse
		err   error
	)

	if err = r.ParseForm(); err != nil {
		log.Errorf("failed to parse form for autodownload log: %v\n", err)
		sendErrorAnswer(w, http.StatusBadRequest, err)
		return
	}

	start, length, err := getPaginationParams(r.Form)
	if err != nil {
		log.Errorf("Could not get pagination params for autodownload log: %v", err)
		sendErrorAnswer(w, http.StatusBadRequest, err)
		return
	}

	qURL := fmt.Sprintf(PriceAutoDownloadAPIArg+"log?skip=%d&limit=%d", start, length)
	qURL = addAutoDownloadQueueQueryParams(r.Form, qURL)

	respStatus, respBody, err := RequestStorage.MakeGetRequest(qURL)
	if err != nil {
		log.Errorf("Could not get resp from price autodownload api: %v", err)
		sendErrorAnswer(w, http.StatusInternalServerError, err)
		return
	}

	if respStatus == http.StatusNotFound {

		notFoundResp := autodownload.LogResponse{
			Meta: models.Meta{
				Status: http.StatusNoContent,
			},
			Response: autodownload.PagesLog{
				Count: 0,
				Log:   []autodownload.PageLog{},
			},
		}

		bytesResp, err := json.Marshal(notFoundResp)
		if err != nil {
			log.Errorf("failed to marshal response data for autodownload log: %v\n", err)
			sendErrorAnswer(w, http.StatusInternalServerError, err)
			return
		}

		sendAnswer(w, http.StatusNoContent, http.Header{}, bytesResp)
		return
	}

	if respStatus != http.StatusOK {
		var errResp models.ErrorResp
		if err = json.Unmarshal(respBody, &errResp); err != nil {
			log.Errorf("failed to unmarshal data to response structure for autodownload log: %v\n", err)
			sendErrorAnswer(w, http.StatusInternalServerError, err)
			return
		}

		sendErrorAnswer(w, errResp.Meta.Status, errors.New(errResp.Meta.Message))
		return
	}

	if err = json.Unmarshal(respBody, &tasks); err != nil {
		log.Errorf("failed to unmarshal data to response structure for autodownload log: %v\n", err)
		sendErrorAnswer(w, http.StatusInternalServerError, err)
		return
	}

	tasks.Response.Log = fixAutoDownloadTasks(tasks.Response.Log)

	tasksBytes, err := json.Marshal(tasks)
	if err != nil {
		log.Errorf("failed to marshal response data for autodownload log: %v\n", err)
		sendErrorAnswer(w, http.StatusInternalServerError, err)
		return
	}

	sendAnswer(w, http.StatusOK, http.Header{}, tasksBytes)
}

// GetAutoDownloadConfigs handles the request to return auto download configs list from price autodownload api.
func GetAutoDownloadConfigs(w http.ResponseWriter, r *http.Request) {

	var (
		configs autodownload.ConfigsResponse
		err     error
	)

	if err = r.ParseForm(); err != nil {
		log.Errorf("failed to parse form for autodownload configs: %v\n", err)
		sendErrorAnswer(w, http.StatusBadRequest, err)
		return
	}

	start, length, err := getPaginationParams(r.Form)
	if err != nil {
		log.Errorf("Could not get pagination params for autodownload configs: %v", err)
		sendErrorAnswer(w, http.StatusBadRequest, err)
		return
	}

	qURL := fmt.Sprintf(PriceAutoDownloadAPIArg+"configs?skip=%d&limit=%d", start, length)
	qURL = addAutoDownloadConfigsQueryParams(r.Form, qURL)

	respStatus, respBody, err := RequestStorage.MakeGetRequest(qURL)
	if err != nil {
		log.Errorf("Could not get resp from price autodownload api: %v", err)
		sendErrorAnswer(w, http.StatusInternalServerError, err)
		return
	}

	if respStatus == http.StatusNotFound {

		notFoundResp := autodownload.ConfigsResponse{
			Meta: models.Meta{
				Status: http.StatusNoContent,
			},
			Response: autodownload.Configs{
				Count:   0,
				Configs: []autodownload.Config{},
			},
		}

		bytesResp, err := json.Marshal(notFoundResp)
		if err != nil {
			log.Errorf("failed to marshal response data for autodownload configs: %v\n", err)
			sendErrorAnswer(w, http.StatusInternalServerError, err)
			return
		}

		sendAnswer(w, http.StatusNoContent, http.Header{}, bytesResp)
		return
	}

	if respStatus != http.StatusOK {
		var errResp models.ErrorResp
		if err = json.Unmarshal(respBody, &errResp); err != nil {
			log.Errorf("failed to unmarshal data to response structure for autodownload configs: %v\n", err)
			sendErrorAnswer(w, http.StatusInternalServerError, err)
			return
		}

		sendErrorAnswer(w, errResp.Meta.Status, errors.New(errResp.Meta.Message))
		return
	}

	if err = json.Unmarshal(respBody, &configs); err != nil {
		log.Errorf("failed to unmarshal data to response structure for autodownload configs: %v\n", err)
		sendErrorAnswer(w, http.StatusInternalServerError, err)
		return
	}

	if length != 1 {
		configs.Response.Configs = fixAutoDownloadConfigs(configs.Response.Configs)
	}

	configsBytes, err := json.Marshal(configs)
	if err != nil {
		log.Errorf("failed to marshal response data for autodownload configs: %v\n", err)
		sendErrorAnswer(w, http.StatusInternalServerError, err)
		return
	}

	sendAnswer(w, http.StatusOK, http.Header{}, configsBytes)
}

// DeleteAutoDownloadConfig handles the request to delete autodownload config from price autodownload_api.
func DeleteAutoDownloadConfig(w http.ResponseWriter, r *http.Request) {

	var (
		resp autodownload.DeleteConfigResponse
		err  error
	)

	v := mux.Vars(r)

	resellerID, err := getInt64Param("resellerId", v["resellerId"])
	if err != nil {
		log.Errorf("Could not get 'resellerId' param for deletion autodownload config: %v", err)
		sendErrorAnswer(w, http.StatusBadRequest, err)
		return
	}

	distributorID, err := getInt64Param("distributorId", v["distributorId"])
	if err != nil {
		log.Errorf("Could not get 'distributorId' param for deletion autodownload config: %v", err)
		sendErrorAnswer(w, http.StatusBadRequest, err)
		return
	}

	qURL := fmt.Sprintf(PriceAutoDownloadAPIArg+"configs/%d/%d/delete", resellerID, distributorID)

	respStatus, respBody, err := RequestStorage.MakePostRequest(qURL, "", "")
	if err != nil {
		log.Errorf("Could not get resp from autodownload api: %v", err)
		sendErrorAnswer(w, http.StatusInternalServerError, err)
		return
	}

	if respStatus != http.StatusOK {
		var errResp models.ErrorResp
		if err = json.Unmarshal(respBody, &errResp); err != nil {
			log.Errorf("failed to unmarshal data to response structure for delete autodownload config: %v\n", err)
			sendErrorAnswer(w, http.StatusInternalServerError, err)
			return
		}

		sendErrorAnswer(w, errResp.Meta.Status, errors.New(errResp.Meta.Message))
		return
	}

	if err = json.Unmarshal(respBody, &resp); err != nil {
		log.Errorf("failed to unmarshal data to response structure for delete autodownload config: %v\n", err)
		sendErrorAnswer(w, http.StatusInternalServerError, err)
		return
	}

	bytesResp, err := json.Marshal(resp)
	if err != nil {
		log.Errorf("failed to marshal response data for delete autodownload config: %v\n", err)
		sendErrorAnswer(w, http.StatusInternalServerError, err)
		return
	}

	sendAnswer(w, http.StatusOK, http.Header{}, bytesResp)
}

// UpdateAutoDownloadConfig handles the request to update auto download config from price autodownload api.
func UpdateAutoDownloadConfig(w http.ResponseWriter, r *http.Request) {

	var (
		config autodownload.UpdateOrAddConfigResponse
		err    error
	)

	inputConfig, err := getAutoDownloadConfigReqBody(r.Body)
	if err != nil {
		log.Errorf("Could not get autodownload request body config for update config route: %v", err)
		sendErrorAnswer(w, http.StatusBadRequest, err)
		return
	}

	if inputConfig.ResellerID == 0 {
		sendErrorAnswer(w, http.StatusBadRequest, errors.New("config parameter 'resellerId' cannot be null"))
		return
	}

	if inputConfig.DistributorID == 0 {
		sendErrorAnswer(w, http.StatusBadRequest, errors.New("config parameter 'distributorId' cannot be null"))
		return
	}

	qURL := fmt.Sprintf(PriceAutoDownloadAPIArg+"configs/%d/%d/update", inputConfig.ResellerID, inputConfig.DistributorID)

	inputConfigBytes, err := json.Marshal(inputConfig)
	if err != nil {
		log.Errorf("failed to marshal config parameter for update autodownload config: %v\n", err)
		sendErrorAnswer(w, http.StatusInternalServerError, err)
		return
	}

	respStatus, respBody, err := RequestStorage.MakePostRequest(qURL, fmt.Sprintf(`config=%s`, string(inputConfigBytes)), appFormURLEncoded)
	if err != nil {
		log.Errorf("Could not get resp from autodownload api: %v", err)
		sendErrorAnswer(w, http.StatusInternalServerError, err)
		return
	}

	if respStatus != http.StatusOK {
		var errResp models.ErrorResp
		if err = json.Unmarshal(respBody, &errResp); err != nil {
			log.Errorf("failed to unmarshal data to response structure for update of autodownload config: %v\n", err)
			sendErrorAnswer(w, http.StatusInternalServerError, err)
			return
		}

		sendErrorAnswer(w, errResp.Meta.Status, errors.New(errResp.Meta.Message))
		return
	}

	if err = json.Unmarshal(respBody, &config); err != nil {
		log.Errorf("failed to unmarshal data to response structure for update autodownload config: %v\n", err)
		sendErrorAnswer(w, http.StatusInternalServerError, err)
		return
	}

	configBytes, err := json.Marshal(config)
	if err != nil {
		log.Errorf("failed to marshal response data for update autodownload config: %v\n", err)
		sendErrorAnswer(w, http.StatusInternalServerError, err)
		return
	}

	sendAnswer(w, http.StatusOK, http.Header{}, configBytes)
}

// AddAutoDownloadConfig handles the request to add auto download config from price autodownload api.
func AddAutoDownloadConfig(w http.ResponseWriter, r *http.Request) {

	var (
		config autodownload.UpdateOrAddConfigResponse
		err    error
	)

	inputConfig, err := getAutoDownloadConfigReqBody(r.Body)
	if err != nil {
		log.Errorf("Could not get autodownload request body config for adding config route: %v", err)
		sendErrorAnswer(w, http.StatusBadRequest, err)
		return
	}

	qURL := PriceAutoDownloadAPIArg + "configs/new"

	inputConfigBytes, err := json.Marshal(inputConfig)
	if err != nil {
		log.Errorf("failed to marshal config parameter for adding autodownload config: %v\n", err)
		sendErrorAnswer(w, http.StatusInternalServerError, err)
		return
	}

	respStatus, respBody, err := RequestStorage.MakePostRequest(qURL, fmt.Sprintf(`config=%s`, string(inputConfigBytes)), appFormURLEncoded)
	if err != nil {
		log.Errorf("Could not get resp from price autodownload api: %v", err)
		sendErrorAnswer(w, http.StatusInternalServerError, err)
		return
	}

	if respStatus != http.StatusOK {
		var errResp models.ErrorResp
		if err = json.Unmarshal(respBody, &errResp); err != nil {
			log.Errorf("failed to unmarshal data to response structure for adding of autodownload config: %v\n", err)
			sendErrorAnswer(w, http.StatusInternalServerError, err)
			return
		}

		sendErrorAnswer(w, errResp.Meta.Status, errors.New(errResp.Meta.Message))
		return
	}

	if err = json.Unmarshal(respBody, &config); err != nil {
		log.Errorf("failed to unmarshal data to response structure for adding of autodownload config: %v\n", err)
		sendErrorAnswer(w, http.StatusInternalServerError, err)
		return
	}

	configBytes, err := json.Marshal(config)
	if err != nil {
		log.Errorf("failed to marshal response data for adding of autodownload config: %v\n", err)
		sendErrorAnswer(w, http.StatusInternalServerError, err)
		return
	}

	sendAnswer(w, http.StatusOK, http.Header{}, configBytes)
}
