package controller

import (
	"net/http"

	"viot/api/response"
	"viot/logger"
	"viot/pkg/webservice"

	"github.com/gin-gonic/gin"
)

// DataCenterController 處理數據中心相關的 API 請求
type DataCenterController struct {
	service webservice.WebManager
	logger  logger.Logger
}

// NewDataCenterController 創建一個新的數據中心控制器
func NewDataCenterController(service webservice.WebManager, logger logger.Logger) *DataCenterController {
	return &DataCenterController{
		service: service,
		logger:  logger.Named("datacenter-controller"),
	}
}

// GetPDUData 獲取 PDU 數據
// @Summary 獲取 PDU 數據
// @Description 獲取指定房間的 PDU 數據
// @Tags PDU
// @Accept json
// @Produce json
// @Param room query string false "房間名稱，如果不指定則獲取所有房間的數據"
// @Success 200 {object} response.Response
// @Failure 400 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /api/pdu [get]
func (c *DataCenterController) GetPDUData(ctx *gin.Context) {
	room := ctx.Query("room")
	c.logger.Debug("獲取 PDU 數據", logger.String("room", room))

	// 獲取 PDU 數據
	pduData, err := c.service.GetPDUData(room)
	if err != nil {
		c.logger.Error("獲取 PDU 數據失敗", logger.String("room", room), logger.Any("error", err))
		response.Fail(ctx, http.StatusInternalServerError, "獲取 PDU 數據失敗", err.Error())
		return
	}

	response.Success(ctx, "獲取 PDU 數據成功", pduData)
}

// UpdatePDUStatus 更新 PDU 狀態
// @Summary 更新 PDU 狀態
// @Description 更新指定 PDU 的狀態
// @Tags PDU
// @Accept json
// @Produce json
// @Param pdu_name path string true "PDU 名稱"
// @Param status body string true "PDU 狀態"
// @Success 200 {object} response.Response
// @Failure 400 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /api/pdu/{pdu_name}/status [put]
func (c *DataCenterController) UpdatePDUStatus(ctx *gin.Context) {
	pduName := ctx.Param("pdu_name")
	var status struct {
		Status string `json:"status" binding:"required"`
	}

	if err := ctx.ShouldBindJSON(&status); err != nil {
		c.logger.Error("請求參數無效", logger.String("pdu", pduName), logger.Any("error", err))
		response.Fail(ctx, http.StatusBadRequest, "請求參數無效", err.Error())
		return
	}

	c.logger.Debug("更新 PDU 狀態", logger.String("pdu", pduName), logger.String("status", status.Status))

	// 更新 PDU 狀態
	err := c.service.UpdatePDUStatus(pduName, status.Status)
	if err != nil {
		c.logger.Error("更新 PDU 狀態失敗", logger.String("pdu", pduName), logger.String("status", status.Status), logger.Any("error", err))
		response.Fail(ctx, http.StatusInternalServerError, "更新 PDU 狀態失敗", err.Error())
		return
	}

	response.Success(ctx, "更新 PDU 狀態成功", nil)
}

// UpdatePDUTags 更新 PDU 標籤
// @Summary 更新 PDU 標籤
// @Description 更新指定 PDU 的標籤
// @Tags PDU
// @Accept json
// @Produce json
// @Param pdu_name path string true "PDU 名稱"
// @Param tags body []string true "PDU 標籤"
// @Success 200 {object} response.Response
// @Failure 400 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /api/pdu/{pdu_name}/tags [put]
func (c *DataCenterController) UpdatePDUTags(ctx *gin.Context) {
	pduName := ctx.Param("pdu_name")
	var tags struct {
		Tags []string `json:"tags" binding:"required"`
	}

	if err := ctx.ShouldBindJSON(&tags); err != nil {
		c.logger.Error("請求參數無效", logger.String("pdu", pduName), logger.Any("error", err))
		response.Fail(ctx, http.StatusBadRequest, "請求參數無效", err.Error())
		return
	}

	c.logger.Debug("更新 PDU 標籤", logger.String("pdu", pduName), logger.Any("tags", tags.Tags))

	// 更新 PDU 標籤
	err := c.service.UpdatePDUTags(pduName, tags.Tags)
	if err != nil {
		c.logger.Error("更新 PDU 標籤失敗", logger.String("pdu", pduName), logger.Any("tags", tags.Tags), logger.Any("error", err))
		response.Fail(ctx, http.StatusInternalServerError, "更新 PDU 標籤失敗", err.Error())
		return
	}

	response.Success(ctx, "更新 PDU 標籤成功", nil)
}

// GetPDUTags 獲取 PDU 標籤
// @Summary 獲取 PDU 標籤
// @Description 獲取指定 PDU 的標籤
// @Tags PDU
// @Accept json
// @Produce json
// @Param pdu_name path string true "PDU 名稱"
// @Success 200 {object} response.Response
// @Failure 400 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /api/pdu/{pdu_name}/tags [get]
func (c *DataCenterController) GetPDUTags(ctx *gin.Context) {
	pduName := ctx.Param("pdu_name")
	c.logger.Debug("獲取 PDU 標籤", logger.String("pdu", pduName))

	// 獲取 PDU 標籤
	tags, err := c.service.GetPDUTags(pduName)
	if err != nil {
		c.logger.Error("獲取 PDU 標籤失敗", logger.String("pdu", pduName), logger.Any("error", err))
		response.Fail(ctx, http.StatusInternalServerError, "獲取 PDU 標籤失敗", err.Error())
		return
	}

	response.Success(ctx, "獲取 PDU 標籤成功", tags)
}

// GetDeviceData 獲取空調機架數據
// @Summary 獲取空調機架數據
// @Description 獲取指定房間的空調機架數據
// @Tags ACRack
// @Accept json
// @Produce json
// @Param room query string false "房間名稱，如果不指定則獲取所有房間的數據"
// @Success 200 {object} response.Response
// @Failure 400 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /api/rack [get]
func (c *DataCenterController) GetDeviceData(ctx *gin.Context) {
	room := ctx.Query("room")
	c.logger.Debug("獲取空調機架數據", logger.String("room", room))

	// 獲取空調機架數據
	rackData, err := c.service.GetDeviceData(room)
	if err != nil {
		c.logger.Error("獲取空調機架數據失敗", logger.String("room", room), logger.Any("error", err))
		response.Fail(ctx, http.StatusInternalServerError, "獲取空調機架數據失敗", err.Error())
		return
	}

	response.Success(ctx, "獲取空調機架數據成功", rackData)
}

// GetRooms 獲取所有房間
// @Summary 獲取所有房間
// @Description 獲取所有房間
// @Tags Room
// @Accept json
// @Produce json
// @Success 200 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /api/rooms [get]
func (c *DataCenterController) GetRooms(ctx *gin.Context) {
	c.logger.Debug("獲取所有房間")

	// 獲取所有房間
	rooms, err := c.service.GetRooms()
	if err != nil {
		c.logger.Error("獲取所有房間失敗", logger.Any("error", err))
		response.Fail(ctx, http.StatusInternalServerError, "獲取所有房間失敗", err.Error())
		return
	}

	response.Success(ctx, "獲取所有房間成功", rooms)
}
