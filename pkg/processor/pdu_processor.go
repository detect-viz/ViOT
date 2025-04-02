package processor

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"viot/logger"
	"viot/models"

	"go.uber.org/zap"
)

// PDUProcessor PDU 數據處理器
type PDUProcessor struct {
	*BaseProcessor
	config       models.PDUProcessorConfig
	scales       []models.PDUScale
	handlers     map[string]map[string]models.PDUManufacturerHandler
	outputRouter models.OutputRouter
}

// NewPDUProcessor 創建 PDU 處理器
func NewPDUProcessor(name string, config models.ProcessorConfig, logger logger.Logger) *PDUProcessor {
	base := NewBaseProcessor(name, config, logger)
	pduConfig := models.PDUProcessorConfig{
		Measurement: "pdu",
	}

	// 從config.Options解析PDU配置
	if options, ok := config.Options["pdu"].(map[string]interface{}); ok {
		if measurement, ok := options["measurement"].(string); ok {
			pduConfig.Measurement = measurement
		}

		// 解析Schema
		if schemaOpts, ok := options["schema"].(map[string]interface{}); ok {
			if field, ok := schemaOpts["total_current"].(string); ok {
				pduConfig.Schema.TotalCurrentField = field
			}
			if field, ok := schemaOpts["total_power"].(string); ok {
				pduConfig.Schema.TotalPowerField = field
			}
			if field, ok := schemaOpts["total_energy"].(string); ok {
				pduConfig.Schema.TotalEnergyField = field
			}
			// 其他字段解析...
		}
	}

	p := &PDUProcessor{
		BaseProcessor: base,
		config:        pduConfig,
		scales:        []models.PDUScale{},
		handlers:      make(map[string]map[string]models.PDUManufacturerHandler),
	}

	// 添加默認處理程序
	p.registerDefaultHandlers()

	return p
}

// SetOutputRouter 設置輸出路由器
func (p *PDUProcessor) SetOutputRouter(router models.OutputRouter) {
	p.outputRouter = router
}

// AddScale 添加比例因子
func (p *PDUProcessor) AddScale(scale models.PDUScale) {
	p.scales = append(p.scales, scale)
}

// RegisterHandler 註冊處理程序
func (p *PDUProcessor) RegisterHandler(manufacturer, model string, handler models.PDUManufacturerHandler) {
	manufacturer = strings.ToLower(manufacturer)
	model = strings.ToLower(model)

	if _, ok := p.handlers[manufacturer]; !ok {
		p.handlers[manufacturer] = make(map[string]models.PDUManufacturerHandler)
	}

	p.handlers[manufacturer][model] = handler
}

// Process 處理PDU數據
func (p *PDUProcessor) Process(data []models.DeviceData) error {
	if len(data) == 0 {
		return fmt.Errorf("沒有提供數據")
	}

	// 創建上下文
	ctx := context.Background()

	// 轉換數據為PDU點
	var points []models.PDUPoint
	for _, d := range data {
		point := models.PDUPoint{
			Name:      d.Name,
			Timestamp: d.Timestamp,
			Tags:      d.Tags,
			Fields:    make(map[string]interface{}),
		}

		// 轉換字段
		for k, v := range d.Fields {
			point.Fields[k] = v
		}

		points = append(points, point)
	}

	// 處理PDU數據點
	_, err := p.ProcessPDUPoints(ctx, points)
	if err != nil {
		p.IncrementErrorCount(1)
		return err
	}

	// 增加處理計數
	p.IncrementProcessedCount(int64(len(points)))

	return nil
}

// ProcessPDUPoints 處理多個PDU數據點
func (p *PDUProcessor) ProcessPDUPoints(ctx context.Context, points []models.PDUPoint) ([]models.PDUData, error) {
	if len(points) == 0 {
		return nil, fmt.Errorf("沒有提供數據點")
	}

	var results []models.PDUData

	for _, point := range points {
		// 檢查是否為PDU數據
		if !p.isPDUData(point) {
			continue
		}

		pduData, err := p.ProcessPDUPoint(ctx, point)
		if err != nil {
			p.GetLogger().Error("處理PDU數據點失敗",
				zap.String("name", point.Name),
				zap.Error(err))
			continue
		}

		results = append(results, pduData)
	}

	// 增加處理計數
	p.IncrementProcessedCount(int64(len(results)))

	// 發送到輸出路由
	if p.outputRouter != nil && len(results) > 0 {
		if err := p.outputRouter.RoutePDUData(ctx, results); err != nil {
			p.GetLogger().Error("PDU數據路由失敗", zap.Error(err))
		}
	}

	return results, nil
}

// ProcessPDUPoint 處理單個PDU數據點
func (p *PDUProcessor) ProcessPDUPoint(ctx context.Context, point models.PDUPoint) (models.PDUData, error) {
	// 創建PDU數據結構
	pduData := models.PDUData{
		Name:      point.Name,
		Timestamp: point.Timestamp,
		Tags:      make(map[string]string),
		Branches:  []models.Branch{},
		Phases:    []models.Phase{},
	}

	// 複製標籤
	for k, v := range point.Tags {
		pduData.Tags[k] = v
	}

	// 獲取製造商和型號
	manufacturer := "unknown"
	model := "unknown"

	if mfr, ok := point.Tags["manufacturer"]; ok {
		manufacturer = strings.ToLower(mfr)
	}

	if mdl, ok := point.Tags["model"]; ok {
		model = strings.ToLower(mdl)
	} else {
		model = p.detectPDUTypeFromFields(point)
	}

	// 獲取處理器
	var handler models.PDUManufacturerHandler

	// 嘗試獲取指定製造商和型號的處理程序
	if manufHandlers, ok := p.handlers[manufacturer]; ok {
		if h, ok := manufHandlers[model]; ok {
			handler = h
		} else if h, ok := manufHandlers["default"]; ok {
			handler = h
		}
	}

	// 如果沒有找到，使用默認處理程序
	if handler == nil {
		handler = &DefaultPDUHandler{
			scale: p.getScaleForManufacturer(manufacturer),
		}
	}

	// 處理字段
	processedFields, err := handler.ProcessFields(point)
	if err != nil {
		p.GetLogger().Warn("處理字段時出錯",
			zap.String("name", point.Name),
			zap.String("manufacturer", manufacturer),
			zap.String("model", model),
			zap.Error(err))
	}

	scale := handler.GetScaleFactor()

	// 處理總體字段
	p.processTotalFields(point, &pduData, scale, processedFields)

	// 處理分支字段
	p.processBranchFields(point, &pduData, scale, processedFields)

	// 處理相位字段
	p.processPhaseFields(point, &pduData, scale, processedFields)

	// 發送到輸出路由
	if p.outputRouter != nil {
		if err := p.outputRouter.RoutePDUData(ctx, pduData); err != nil {
			p.GetLogger().Warn("PDU數據路由失敗",
				zap.String("name", point.Name),
				zap.Error(err))
		}
	}

	return pduData, nil
}

// isPDUData 判斷數據點是否為PDU數據
func (p *PDUProcessor) isPDUData(point models.PDUPoint) bool {
	// 根據測量名稱判斷
	if point.Name == p.config.Measurement {
		return true
	}

	// 根據標籤判斷
	if deviceType, ok := point.Tags["device_type"]; ok && deviceType == "pdu" {
		return true
	}

	// 根據字段判斷
	hasCurrentField := false
	hasVoltageField := false

	for field := range point.Fields {
		switch {
		case strings.Contains(field, "current"):
			hasCurrentField = true
		case strings.Contains(field, "voltage"):
			hasVoltageField = true
		}
	}

	return hasCurrentField && hasVoltageField
}

// 處理總體字段
func (p *PDUProcessor) processTotalFields(point models.PDUPoint, pdu *models.PDUData, scale models.PDUScale, processedFields map[string]float64) {
	// 處理電流
	if current, ok := getFieldValue(processedFields, p.config.Schema.TotalCurrentField, "current"); ok {
		pdu.Current = current * scale.Current
	}

	// 處理電壓
	if voltage, ok := getFieldValue(processedFields, "", "voltage"); ok {
		pdu.Voltage = voltage * scale.Voltage
	}

	// 處理功率
	if power, ok := getFieldValue(processedFields, p.config.Schema.TotalPowerField, "power"); ok {
		pdu.Power = power * scale.Power
	}

	// 處理能耗
	if energy, ok := getFieldValue(processedFields, p.config.Schema.TotalEnergyField, "energy"); ok {
		pdu.Energy = energy * scale.Energy
	}
}

// 處理分支字段
func (p *PDUProcessor) processBranchFields(point models.PDUPoint, pdu *models.PDUData, scale models.PDUScale, processedFields map[string]float64) {
	branchMap := make(map[string]*models.Branch)

	// 尋找分支相關字段
	for field, value := range processedFields {
		if !strings.Contains(field, "branch") {
			continue
		}

		// 解析分支ID
		branchID := extractIDFromField(field, "branch")
		if branchID == "" {
			continue
		}

		// 獲取或創建分支
		branch, ok := branchMap[branchID]
		if !ok {
			branch = &models.Branch{ID: branchID}
			branchMap[branchID] = branch
		}

		// 根據字段類型更新分支數據
		switch {
		case strings.Contains(field, "current"):
			branch.Current = value * scale.Current
		case strings.Contains(field, "voltage"):
			branch.Voltage = value * scale.Voltage
		case strings.Contains(field, "power"):
			branch.Power = value * scale.Power
		case strings.Contains(field, "energy"):
			branch.Energy = value * scale.Energy
		}
	}

	// 將分支添加到PDU數據
	for _, branch := range branchMap {
		pdu.Branches = append(pdu.Branches, *branch)
	}
}

// 處理相位字段
func (p *PDUProcessor) processPhaseFields(point models.PDUPoint, pdu *models.PDUData, scale models.PDUScale, processedFields map[string]float64) {
	phaseMap := make(map[string]*models.Phase)

	// 尋找相位相關字段
	for field, value := range processedFields {
		phaseID := ""

		// 檢查是否包含phase或L1/L2/L3
		if strings.Contains(field, "phase") {
			phaseID = extractIDFromField(field, "phase")
		} else if strings.Contains(field, "L1") {
			phaseID = "L1"
		} else if strings.Contains(field, "L2") {
			phaseID = "L2"
		} else if strings.Contains(field, "L3") {
			phaseID = "L3"
		}

		if phaseID == "" {
			continue
		}

		// 獲取或創建相位
		phase, ok := phaseMap[phaseID]
		if !ok {
			phase = &models.Phase{ID: phaseID}
			phaseMap[phaseID] = phase
		}

		// 根據字段類型更新相位數據
		switch {
		case strings.Contains(field, "current"):
			phase.Current = value * scale.Current
		case strings.Contains(field, "voltage"):
			phase.Voltage = value * scale.Voltage
		case strings.Contains(field, "power"):
			phase.Power = value * scale.Power
		case strings.Contains(field, "energy"):
			phase.Energy = value * scale.Energy
		}
	}

	// 將相位添加到PDU數據
	for _, phase := range phaseMap {
		pdu.Phases = append(pdu.Phases, *phase)
	}
}

// getFieldValue 獲取字段值，先嘗試指定的字段名，再試通用名稱
func getFieldValue(fields map[string]float64, specificField, genericName string) (float64, bool) {
	if specificField != "" {
		if value, ok := fields[specificField]; ok {
			return value, true
		}
	}

	for field, value := range fields {
		if strings.Contains(field, genericName) && !strings.Contains(field, "branch") && !strings.Contains(field, "phase") {
			return value, true
		}
	}

	return 0, false
}

// extractIDFromField 從字段名提取ID
func extractIDFromField(fieldName, prefix string) string {
	prefixPattern := prefix + "_"
	if !strings.Contains(fieldName, prefixPattern) {
		prefixPattern = prefix
	}

	parts := strings.Split(fieldName, prefixPattern)
	if len(parts) <= 1 {
		return ""
	}

	idPart := parts[1]
	if idx := strings.Index(idPart, "_"); idx > 0 {
		return idPart[:idx]
	}
	return idPart
}

// getScaleForManufacturer 獲取製造商對應的比例因子
func (p *PDUProcessor) getScaleForManufacturer(manufacturer string) models.PDUScale {
	for _, scale := range p.scales {
		if strings.EqualFold(scale.Manufacturer, manufacturer) {
			return scale
		}
	}
	return models.PDUScale{
		Manufacturer: "default",
		Current:      1.0,
		Voltage:      1.0,
		Power:        1.0,
		Energy:       1.0,
	}
}

// detectPDUTypeFromFields 嘗試從字段檢測PDU型號
func (p *PDUProcessor) detectPDUTypeFromFields(point models.PDUPoint) string {
	// 根據字段模式檢測型號
	hasL1 := false
	hasL2 := false
	hasL3 := false
	hasBranch := false

	for field := range point.Fields {
		if strings.Contains(field, "L1") {
			hasL1 = true
		}
		if strings.Contains(field, "L2") {
			hasL2 = true
		}
		if strings.Contains(field, "L3") {
			hasL3 = true
		}
		if strings.Contains(field, "branch") {
			hasBranch = true
		}
	}

	// 基於字段特性推斷型號
	if hasBranch && hasL1 && hasL2 && hasL3 {
		return "pdue428" // Delta PDU E428
	} else if hasL1 && hasL2 && hasL3 {
		return "pdu1315" // Delta PDU 1315
	}

	return "unknown"
}

// registerDefaultHandlers 註冊默認處理程序
func (p *PDUProcessor) registerDefaultHandlers() {
	// 註冊Delta處理程序
	deltaScale := models.PDUScale{
		Manufacturer: "delta",
		Current:      0.1,   // 10分之1安培
		Voltage:      0.1,   // 10分之1伏特
		Power:        1.0,   // 瓦特
		Energy:       0.001, // 千瓦時
	}

	// Delta PDU4425 (Modbus)
	p.RegisterHandler("delta", "pdu4425", &DeltaPDUHandler{
		scale:     deltaScale,
		modelType: "PDU4425",
		slaveID:   1,
	})

	// Delta PDU1315 (Modbus)
	p.RegisterHandler("delta", "pdu1315", &DeltaPDUHandler{
		scale:     deltaScale,
		modelType: "PDU1315",
		slaveID:   1,
	})

	// Delta PDUE428 (SNMP)
	p.RegisterHandler("delta", "pdue428", &DeltaPDUHandler{
		scale:     deltaScale,
		modelType: "PDUE428",
		slaveID:   0,
	})

	// 註冊Vertiv處理程序
	vertivScale := models.PDUScale{
		Manufacturer: "vertiv",
		Current:      0.1,   // 10分之1安培
		Voltage:      0.1,   // 10分之1伏特
		Power:        1.0,   // 瓦特
		Energy:       0.001, // 千瓦時
	}

	// Vertiv 6PS56 (SNMP)
	p.RegisterHandler("vertiv", "6ps56", &VertivPDUHandler{
		scale:     vertivScale,
		modelType: "6PS56",
	})
}

// DefaultPDUHandler 默認PDU處理程序
type DefaultPDUHandler struct {
	scale models.PDUScale
}

// CanHandle 檢查是否可以處理
func (h *DefaultPDUHandler) CanHandle(manufacturer, model string) bool {
	return true // 默認處理程序可以處理任何PDU
}

// ProcessFields 處理字段
func (h *DefaultPDUHandler) ProcessFields(point models.PDUPoint) (map[string]float64, error) {
	// 直接返回原始字段轉換成float64
	return convertToFloat64Map(point.Fields), nil
}

// GetScaleFactor 獲取比例因子
func (h *DefaultPDUHandler) GetScaleFactor() models.PDUScale {
	return h.scale
}

// DeltaPDUHandler Delta PDU處理程序
type DeltaPDUHandler struct {
	scale     models.PDUScale
	modelType string
	slaveID   int
}

// CanHandle 檢查是否可以處理
func (h *DeltaPDUHandler) CanHandle(manufacturer, model string) bool {
	return strings.EqualFold(manufacturer, "delta")
}

// ProcessFields 處理Delta PDU字段
func (h *DeltaPDUHandler) ProcessFields(point models.PDUPoint) (map[string]float64, error) {
	fields := convertToFloat64Map(point.Fields)

	// 檢查slave_id是否匹配（如果有設置）
	if h.slaveID != 0 {
		if slaveIDStr, exists := point.Tags["slave_id"]; exists {
			slaveID, err := strconv.Atoi(slaveIDStr)
			if err == nil && slaveID != h.slaveID {
				return fields, fmt.Errorf("slave_id不匹配: 期望 %d, 實際 %d", h.slaveID, slaveID)
			}
		}
	}

	// 根據不同型號進行特殊處理
	switch h.modelType {
	case "PDU4425":
		requiredFields := []string{
			"current_L1", "current_L2", "current_L3",
			"voltage_L1", "voltage_L2", "voltage_L3",
			"power_L1", "power_L2", "power_L3",
		}
		if !checkRequiredFields(fields, requiredFields) {
			return fields, fmt.Errorf("delta pdu4425 缺少必需字段")
		}

	case "PDU1315":
		requiredFields := []string{
			"current_L1", "current_L2", "current_L3",
			"voltage_L1", "voltage_L2", "voltage_L3",
		}
		if !checkRequiredFields(fields, requiredFields) {
			return fields, fmt.Errorf("delta pdu1315 缺少必需字段")
		}

	case "PDUE428":
		requiredFields := []string{
			"current_L1", "current_L2", "current_L3",
			"voltage_L1", "voltage_L2", "voltage_L3",
			"power_L1", "power_L2", "power_L3",
		}
		if !checkRequiredFields(fields, requiredFields) {
			return fields, fmt.Errorf("delta pdue428 缺少必需字段")
		}
	}

	return fields, nil
}

// GetScaleFactor 獲取比例因子
func (h *DeltaPDUHandler) GetScaleFactor() models.PDUScale {
	return h.scale
}

// VertivPDUHandler Vertiv PDU處理程序
type VertivPDUHandler struct {
	scale     models.PDUScale
	modelType string
}

// CanHandle 檢查是否可以處理
func (h *VertivPDUHandler) CanHandle(manufacturer, model string) bool {
	return strings.EqualFold(manufacturer, "vertiv")
}

// ProcessFields 處理Vertiv PDU字段
func (h *VertivPDUHandler) ProcessFields(point models.PDUPoint) (map[string]float64, error) {
	fields := convertToFloat64Map(point.Fields)

	switch h.modelType {
	case "6PS56":
		requiredFields := []string{
			"current_L1", "current_L2", "current_L3",
			"voltage_L1", "voltage_L2", "voltage_L3",
			"power_L1", "power_L2", "power_L3",
			"energy_L1", "energy_L2", "energy_L3",
		}
		if !checkRequiredFields(fields, requiredFields) {
			return fields, fmt.Errorf("vertiv 6ps56 缺少必需字段")
		}
	}

	return fields, nil
}

// GetScaleFactor 獲取比例因子
func (h *VertivPDUHandler) GetScaleFactor() models.PDUScale {
	return h.scale
}

// 將interface{}類型的map轉換為float64類型的map
func convertToFloat64Map(fields map[string]interface{}) map[string]float64 {
	result := make(map[string]float64)
	for k, v := range fields {
		switch value := v.(type) {
		case float64:
			result[k] = value
		case float32:
			result[k] = float64(value)
		case int:
			result[k] = float64(value)
		case int64:
			result[k] = float64(value)
		case string:
			if f, err := strconv.ParseFloat(value, 64); err == nil {
				result[k] = f
			}
		}
	}
	return result
}

// 檢查是否包含所有必需字段
func checkRequiredFields(fields map[string]float64, requiredFields []string) bool {
	for _, field := range requiredFields {
		if _, ok := fields[field]; !ok {
			return false
		}
	}
	return true
}

// GetStatus 獲取處理器狀態
func (p *PDUProcessor) GetStatus() models.ProcessorStatus {
	return models.ProcessorStatus{
		Running:        p.IsRunning(),
		StartTime:      p.GetLastProcessTime(),
		ProcessedCount: p.GetProcessedCount(),
		ErrorCount:     p.GetErrorCount(),
		LastError:      p.GetLastError(),
	}
}

// IsRunning 檢查處理器是否正在運行
func (p *PDUProcessor) IsRunning() bool {
	return p.BaseProcessor.IsRunning()
}

// GetLastProcessTime 獲取最後處理時間
func (p *PDUProcessor) GetLastProcessTime() time.Time {
	return p.BaseProcessor.GetLastProcessTime()
}

// GetProcessedCount 獲取處理計數
func (p *PDUProcessor) GetProcessedCount() int64 {
	return p.BaseProcessor.GetProcessedCount()
}

// GetErrorCount 獲取錯誤計數
func (p *PDUProcessor) GetErrorCount() int64 {
	return p.BaseProcessor.GetErrorCount()
}

// GetLastError 獲取最後錯誤
func (p *PDUProcessor) GetLastError() string {
	return p.BaseProcessor.GetLastError()
}

// HandlePDUData 處理PDU數據
func (p *PDUProcessor) HandlePDUData(ctx context.Context, data []models.PDUData) error {
	if len(data) == 0 {
		return fmt.Errorf("沒有提供PDU數據")
	}

	// 轉換PDU數據為設備數據
	var deviceData []models.DeviceData
	for _, pdu := range data {
		device := models.DeviceData{
			Name:      pdu.Name,
			Timestamp: pdu.Timestamp,
			Tags:      pdu.Tags,
			Fields:    make(map[string]interface{}),
		}

		// 添加總體數據
		device.Fields["current"] = pdu.Current
		device.Fields["voltage"] = pdu.Voltage
		device.Fields["power"] = pdu.Power
		device.Fields["energy"] = pdu.Energy

		// 添加分支數據
		for _, branch := range pdu.Branches {
			prefix := fmt.Sprintf("branch_%s_", branch.ID)
			device.Fields[prefix+"current"] = branch.Current
			device.Fields[prefix+"voltage"] = branch.Voltage
			device.Fields[prefix+"power"] = branch.Power
			device.Fields[prefix+"energy"] = branch.Energy
		}

		// 添加相位數據
		for _, phase := range pdu.Phases {
			prefix := fmt.Sprintf("phase_%s_", phase.ID)
			device.Fields[prefix+"current"] = phase.Current
			device.Fields[prefix+"voltage"] = phase.Voltage
			device.Fields[prefix+"power"] = phase.Power
			device.Fields[prefix+"energy"] = phase.Energy
		}

		deviceData = append(deviceData, device)
	}

	// 處理轉換後的設備數據
	return p.Process(deviceData)
}

// 添加介面實現確認
var _ models.Processor = (*PDUProcessor)(nil)

// Start 啟動處理器
func (p *PDUProcessor) Start(ctx context.Context) error {
	p.BaseProcessor.SetRunning(true)
	p.BaseProcessor.SetLastProcessTime(time.Now())
	return nil
}
