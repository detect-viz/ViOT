package controller

import (
	"net/http"

	"viot/api/response"
	"viot/pkg/webservice"

	"github.com/gorilla/mux"
	"go.uber.org/zap"
)

// LocationController 位置控制器
type LocationController struct {
	webManager *webservice.WebManager
	logger     *zap.Logger
}

// NewLocationController 創建位置控制器
func NewLocationController(webManager *webservice.WebManager, logger *zap.Logger) *LocationController {
	return &LocationController{
		webManager: webManager,
		logger:     logger.Named("location_controller"),
	}
}

// GetRooms 獲取所有房間
func (c *LocationController) GetRooms(w http.ResponseWriter, r *http.Request) {
	rooms, err := c.webManager.GetRooms()
	if err != nil {
		c.logger.Error("獲取房間列表失敗", zap.Error(err))
		response.RespondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	response.RespondJSON(w, http.StatusOK, rooms)
}

// GetRoomsByDatacenter 獲取指定數據中心的房間
func (c *LocationController) GetRoomsByDatacenter(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	datacenter := vars["datacenter"]

	rooms, err := c.webManager.GetRoomsByDatacenter(datacenter)
	if err != nil {
		c.logger.Error("獲取數據中心房間列表失敗", zap.String("datacenter", datacenter), zap.Error(err))
		response.RespondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	response.RespondJSON(w, http.StatusOK, rooms)
}

// GetDatacenters 獲取所有數據中心
func (c *LocationController) GetDatacenters(w http.ResponseWriter, r *http.Request) {
	datacenters, err := c.webManager.GetDatacenters()
	if err != nil {
		c.logger.Error("獲取數據中心列表失敗", zap.Error(err))
		response.RespondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	response.RespondJSON(w, http.StatusOK, datacenters)
}

// GetPhases 獲取所有階段
func (c *LocationController) GetPhases(w http.ResponseWriter, r *http.Request) {
	phases, err := c.webManager.GetPhases()
	if err != nil {
		c.logger.Error("獲取階段列表失敗", zap.Error(err))
		response.RespondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	response.RespondJSON(w, http.StatusOK, phases)
}

// GetFactories 獲取所有工廠
func (c *LocationController) GetFactories(w http.ResponseWriter, r *http.Request) {
	factories, err := c.webManager.GetFactories()
	if err != nil {
		c.logger.Error("獲取工廠列表失敗", zap.Error(err))
		response.RespondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	response.RespondJSON(w, http.StatusOK, factories)
}

// GetRacksByRoom 獲取指定房間的機架
func (c *LocationController) GetRacksByRoom(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()
	factory := query.Get("factory")
	phase := query.Get("phase")
	datacenter := query.Get("datacenter")
	room := query.Get("room")

	racks, err := c.webManager.GetRacksByRoom(factory, phase, datacenter, room)
	if err != nil {
		c.logger.Error("獲取機架列表失敗",
			zap.String("factory", factory),
			zap.String("phase", phase),
			zap.String("datacenter", datacenter),
			zap.String("room", room),
			zap.Error(err))
		response.RespondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	response.RespondJSON(w, http.StatusOK, racks)
}
