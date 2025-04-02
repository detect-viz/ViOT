package controller

import (
	"net/http"
	"time"

	"viot/pkg/processor"
	"viot/logger"
	"viot/models"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// TelegrafController 處理Telegraf數據的控制器
type TelegrafController struct {
	collector  *models.Processor
	logger     logger.Logger
	lastError  error
	lastAccess time.Time
}

// NewTelegrafController 創建一個新的Telegraf控制器
func NewTelegrafController(collector *processor.Collector, log logger.Logger) *TelegrafController {
	return &TelegrafController{
		collector:  collector,
		logger:     log.Named("telegraf-controller"),
		lastError:  nil,
		lastAccess: time.Now(),
	}
}

// ReceiveTelegrafData 接收來自Telegraf的數據
func (tc *TelegrafController) ReceiveTelegrafData(c *gin.Context) {
	tc.lastAccess = time.Now()

	// 從請求正文中讀取數據
	err := tc.collector.ProcessDataFromReader(c.Request.Body)
	if err != nil {
		tc.lastError = err
		tc.logger.Error("處理Telegraf數據失敗", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
	})
}

// ReceiveTelegrafMetric 接收單個Telegraf指標
func (tc *TelegrafController) ReceiveTelegrafMetric(c *gin.Context) {
	tc.lastAccess = time.Now()

	var point models.RestAPIPoint
	if err := c.BindJSON(&point); err != nil {
		tc.lastError = err
		tc.logger.Error("解析Telegraf指標失敗", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	err := tc.collector.ProcessSinglePoint(&point)
	if err != nil {
		tc.lastError = err
		tc.logger.Error("處理Telegraf指標失敗", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
	})
}

// GetCollectorStats 獲取收集器處理統計
func (tc *TelegrafController) GetCollectorStats(c *gin.Context) {
	stats := tc.collector.GetStats()

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"stats":   stats,
		"status": gin.H{
			"last_access": tc.lastAccess,
			"error":       tc.lastError != nil,
			"error_msg":   tc.getLastErrorMessage(),
		},
	})
}

// getLastErrorMessage 獲取最後的錯誤消息
func (tc *TelegrafController) getLastErrorMessage() string {
	if tc.lastError == nil {
		return ""
	}
	return tc.lastError.Error()
}
