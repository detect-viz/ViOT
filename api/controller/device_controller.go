package controller

import (
	"encoding/json"
	"net/http"

	"viot/api/response"
	"viot/pkg/webservice"

	"github.com/gorilla/mux"
	"go.uber.org/zap"
)

// DeviceController 設備控制器
type DeviceController struct {
	webManager *webservice.WebManager
	logger     *zap.Logger
}

// NewDeviceController 創建設備控制器
func NewDeviceController(webManager *webservice.WebManager, logger *zap.Logger) *DeviceController {
	return &DeviceController{
		webManager: webManager,
		logger:     logger.Named("device_controller"),
	}
}

// GetDeviceData 獲取設備數據
func (c *DeviceController) GetDeviceData(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	room := vars["room"]

	data, err := c.webManager.GetDeviceData(room)
	if err != nil {
		c.logger.Error("獲取設備數據失敗", zap.String("room", room), zap.Error(err))
		response.RespondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	response.RespondJSON(w, http.StatusOK, data)
}

// UpdateDeviceStatus 更新設備狀態
func (c *DeviceController) UpdateDeviceStatus(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	name := vars["name"]

	var data struct {
		Status string `json:"status"`
	}

	if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
		c.logger.Error("解析請求數據失敗", zap.Error(err))
		response.RespondError(w, http.StatusBadRequest, "無效的請求數據")
		return
	}

	if err := c.webManager.UpdateDeviceStatus(name, data.Status); err != nil {
		c.logger.Error("更新設備狀態失敗", zap.String("name", name), zap.Error(err))
		response.RespondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	response.RespondJSON(w, http.StatusOK, map[string]string{"status": "success"})
}

// GetDeviceTags 獲取設備標籤
func (c *DeviceController) GetDeviceTags(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	name := vars["name"]

	tags, err := c.webManager.GetDeviceTags(name)
	if err != nil {
		c.logger.Error("獲取設備標籤失敗", zap.String("name", name), zap.Error(err))
		response.RespondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	response.RespondJSON(w, http.StatusOK, tags)
}

// UpdateDeviceTags 更新設備標籤
func (c *DeviceController) UpdateDeviceTags(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	name := vars["name"]

	var tags []string
	if err := json.NewDecoder(r.Body).Decode(&tags); err != nil {
		c.logger.Error("解析請求數據失敗", zap.Error(err))
		response.RespondError(w, http.StatusBadRequest, "無效的請求數據")
		return
	}

	if err := c.webManager.UpdateDeviceTags(name, tags); err != nil {
		c.logger.Error("更新設備標籤失敗", zap.String("name", name), zap.Error(err))
		response.RespondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	response.RespondJSON(w, http.StatusOK, map[string]string{"status": "success"})
}

// GetDeviceStatus 獲取設備狀態
func (c *DeviceController) GetDeviceStatus(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	name := vars["name"]

	data, err := c.webManager.GetDeviceStatusByName(name)
	if err != nil {
		c.logger.Error("獲取設備狀態失敗", zap.String("name", name), zap.Error(err))
		response.RespondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	response.RespondJSON(w, http.StatusOK, data)
}

// GetAllDeviceStatuses 獲取所有設備狀態
func (c *DeviceController) GetAllDeviceStatuses(w http.ResponseWriter, r *http.Request) {
	data, err := c.webManager.GetAllDeviceStatuses()
	if err != nil {
		c.logger.Error("獲取所有設備狀態失敗", zap.Error(err))
		response.RespondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	response.RespondJSON(w, http.StatusOK, data)
}
