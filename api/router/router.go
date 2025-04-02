package router

import (
	"os"
	"path/filepath"

	"viot/api/controller"
	"viot/logger"
	"viot/pkg/config"
	"viot/pkg/webservice"

	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/mux"
	"go.uber.org/zap"
)

// Router 是 API 路由器
type Router struct {
	engine               *gin.Engine
	dataCenterController *controller.DataCenterController
	deployController     *controller.DeployController
	telegrafController   *controller.TelegrafController
	config               config.Config
	logger               logger.Logger
	// Web 服務相關配置
	webEnabled   bool
	webPath      string
	staticPath   string
	templatePath string
	configFile   string
}

// NewRouter 創建一個新的路由器
func NewRouter(
	dataCenterController *controller.DataCenterController,
	deployController *controller.DeployController,
	telegrafController *controller.TelegrafController,
	config config.Config,
	logger logger.Logger,
) *Router {
	return &Router{
		engine:               gin.Default(),
		dataCenterController: dataCenterController,
		deployController:     deployController,
		telegrafController:   telegrafController,
		config:               config,
		logger:               logger.Named("router"),
		// 默認 Web 配置
		webEnabled:   true,
		webPath:      "./web",
		staticPath:   "./web/static",
		templatePath: "./web/templates",
		configFile:   "./web/config.json",
	}
}

// Setup 設置路由
func (r *Router) Setup() {
	r.logger.Info("設置路由")

	// 設置 API 路由組
	r.setupAPIRoutes()

	// 設置 Web 路由（如果啟用）
	if r.webEnabled {
		r.setupWebRoutes()
	} else {
		r.logger.Info("Web 服務已禁用")
	}

	r.logger.Info("路由設置完成")
}

// SetWebConfig 設置 Web 相關配置
func (r *Router) SetWebConfig(enabled bool, webPath, staticPath, templatePath, configFile string) {
	r.webEnabled = enabled
	r.webPath = webPath
	r.staticPath = staticPath
	r.templatePath = templatePath
	r.configFile = configFile
}

// setupAPIRoutes 設置 API 路由
func (r *Router) setupAPIRoutes() {
	// API 路由組
	api := r.engine.Group("/api")
	{
		// PDU 相關路由
		api.GET("/pdu", r.dataCenterController.GetPDUData)
		api.PUT("/pdu/:pdu_name/status", r.dataCenterController.UpdatePDUStatus)
		api.GET("/pdu/:pdu_name/tags", r.dataCenterController.GetPDUTags)
		api.PUT("/pdu/:pdu_name/tags", r.dataCenterController.UpdatePDUTags)

		// 空調機架相關路由
		api.GET("/rack", r.dataCenterController.GetDeviceData)

		// 房間相關路由
		api.GET("/rooms", r.dataCenterController.GetRooms)

		// 掃描和部署相關路由
		api.POST("/scan", r.deployController.StartScan)
		api.GET("/scan/status", r.deployController.GetScanStatus)
		api.POST("/scan/stop", r.deployController.StopScan)
		api.GET("/devices", r.deployController.GetDevices)
		api.POST("/devices/:ip/deploy", r.deployController.DeployDevice)

		// Telegraf 數據相關路由
		api.POST("/telegraf/data", r.telegrafController.ReceiveTelegrafData)
		api.POST("/telegraf/metric", r.telegrafController.ReceiveTelegrafMetric)
		api.GET("/telegraf/stats", r.telegrafController.GetCollectorStats)
	}
}

// setupWebRoutes 設置 Web 路由
func (r *Router) setupWebRoutes() {
	// 獲取當前執行目錄
	wd, err := os.Getwd()
	if err != nil {
		r.logger.Error("無法獲取當前工作目錄", logger.Any("error", err))
		return
	}

	// 使用配置中的路徑
	webPathFull := filepath.Join(wd, r.webPath)
	staticPathFull := filepath.Join(wd, r.staticPath)
	templatePathFull := filepath.Join(wd, r.templatePath)
	configFileFull := filepath.Join(wd, r.configFile)

	r.logger.Info("設置 Web 路由",
		logger.String("webPath", webPathFull),
		logger.String("staticPath", staticPathFull),
		logger.String("templatePath", templatePathFull))

	// 設定靜態文件服務
	r.engine.Static("/static", staticPathFull)

	// 提供 config.json 作為靜態文件
	r.engine.StaticFile("/config.json", configFileFull)

	// 設定 HTML 模板目錄
	r.engine.LoadHTMLGlob(filepath.Join(templatePathFull, "*.html"))

	// 設定首頁路由，返回 HTML 頁面
	r.engine.GET("/", func(c *gin.Context) {
		c.HTML(200, "index.html", gin.H{
			"title": "Home Page",
		})
	})

	r.logger.Info("Web 路由設置完成")
}

// Run 啟動 HTTP 服務器
func (r *Router) Run(addr string) error {
	r.logger.Info("啟動 HTTP 服務器", logger.String("addr", addr))
	return r.engine.Run(addr)
}

// GetEngine 獲取 Gin 引擎
func (r *Router) GetEngine() *gin.Engine {
	return r.engine
}

// SetupRouter 設置路由
func SetupRouter(webManager *webservice.WebManager, logger *zap.Logger) http.Handler {
	router := mux.NewRouter()

	// 創建控制器
	deviceController := controller.NewDeviceController(webManager, logger)
	locationController := controller.NewLocationController(webManager, logger)
	statusController := controller.NewStatusController(webManager, logger)

	// API 前綴
	api := router.PathPrefix("/api/v1").Subrouter()

	// 設備相關路由
	api.HandleFunc("/devices/{room}", deviceController.GetDeviceData).Methods("GET")
	api.HandleFunc("/devices/{name}/status", deviceController.UpdateDeviceStatus).Methods("PUT")
	api.HandleFunc("/devices/{name}/tags", deviceController.GetDeviceTags).Methods("GET")
	api.HandleFunc("/devices/{name}/tags", deviceController.UpdateDeviceTags).Methods("PUT")
	api.HandleFunc("/devices/{name}/status", deviceController.GetDeviceStatus).Methods("GET")
	api.HandleFunc("/devices/status", deviceController.GetAllDeviceStatuses).Methods("GET")

	// 位置相關路由
	api.HandleFunc("/rooms", locationController.GetRooms).Methods("GET")
	api.HandleFunc("/datacenters/{datacenter}/rooms", locationController.GetRoomsByDatacenter).Methods("GET")
	api.HandleFunc("/datacenters", locationController.GetDatacenters).Methods("GET")
	api.HandleFunc("/phases", locationController.GetPhases).Methods("GET")
	api.HandleFunc("/factories", locationController.GetFactories).Methods("GET")
	api.HandleFunc("/racks", locationController.GetRacksByRoom).Methods("GET")

	// 狀態相關路由
	api.HandleFunc("/status/sync", statusController.ForceStatusSync).Methods("POST")

	// 添加中間件
	router.Use(loggingMiddleware(logger))
	router.Use(corsMiddleware())

	return router
}

// loggingMiddleware 日誌中間件
func loggingMiddleware(logger *zap.Logger) mux.MiddlewareFunc {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			logger.Info("HTTP請求",
				zap.String("method", r.Method),
				zap.String("path", r.URL.Path),
				zap.String("remote", r.RemoteAddr),
			)
			next.ServeHTTP(w, r)
		})
	}
}

// corsMiddleware CORS 中間件
func corsMiddleware() mux.MiddlewareFunc {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Access-Control-Allow-Origin", "*")
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

			if r.Method == "OPTIONS" {
				w.WriteHeader(http.StatusOK)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
