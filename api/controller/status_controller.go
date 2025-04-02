package controller

import (
	"net/http"

	"viot/api/response"
	"viot/pkg/webservice"

	"go.uber.org/zap"
)

// StatusController 狀態控制器
type StatusController struct {
	webManager *webservice.WebManager
	logger     *zap.Logger
}

// NewStatusController 創建狀態控制器
func NewStatusController(webManager *webservice.WebManager, logger *zap.Logger) *StatusController {
	return &StatusController{
		webManager: webManager,
		logger:     logger.Named("status_controller"),
	}
}

// ForceStatusSync 強制同步狀態
func (c *StatusController) ForceStatusSync(w http.ResponseWriter, r *http.Request) {
	if err := c.webManager.ForceStatusSync(); err != nil {
		c.logger.Error("強制同步狀態失敗", zap.Error(err))
		response.RespondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	response.RespondJSON(w, http.StatusOK, map[string]string{"status": "success"})
}
