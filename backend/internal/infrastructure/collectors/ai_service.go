package services

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"railgun-core/internal/config"
	"railgun-core/internal/domain"
	"railgun-core/internal/domain/models"
	engine "railgun-core/internal/engine/detection"
)

type AIService struct {
	analysisRepo     domain.AnalysisRepository
	timelineRepo     domain.TimelineRepository
	cybertServiceURL string
}

func NewAIService(repo domain.AnalysisRepository, timelineRepo domain.TimelineRepository) *AIService {
	return &AIService{
		analysisRepo:     repo,
		timelineRepo:     timelineRepo,
		cybertServiceURL: "http://host.docker.internal:8001/analyze",
	}
}

func (s *AIService) AnalyzeAndSave(ctx context.Context, logLines []string, hostID string) ([]*models.AnalysisResult, error) {
	var results []*models.AnalysisResult

	for _, logLine := range logLines {
		// Вызов Cybert
		cybertResp, err := s.callCybert(ctx, logLine)
		if err != nil {
			return nil, err
		}

		result := &models.AnalysisResult{
			HostID:       hostID,
			InputData:    logLine,
			PredictLabel: cybertResp["label"].(string),
			Score:        cybertResp["score"].(float64),
			IsMalicious:  cybertResp["is_malicious"].(bool),
			CreatedAt:    time.Now(),
		}

		err = s.analysisRepo.Save(ctx, result)
		if err != nil {
			return nil, err
		}

		results = append(results, result)
	}

	return results, nil
}

func (s *AIService) GetAPTTimeline(ctx context.Context, hostID string, start, end time.Time) (*models.APTTimelineResponse, error) {
	// Достаем данные из ES через репозиторий
	events, err := s.timelineRepo.GetHostTimeline(ctx, hostID, start, end)
	if err != nil {
		return nil, err
	}

	if len(events) == 0 {
		return &models.APTTimelineResponse{HostID: hostID}, nil
	}

	// Формируем ответ
	response := &models.APTTimelineResponse{
		HostID:    hostID,
		StartTime: start.Format(time.RFC3339),
		EndTime:   end.Format(time.RFC3339),
		Events:    events,
	}

	return response, nil
}

func (s *AIService) callCybert(ctx context.Context, logLine string) (map[string]interface{}, error) {
	payload := map[string]string{"log_text": logLine}
	jsonPayload, _ := json.Marshal(payload)

	client := &http.Client{Timeout: 10 * time.Second}
	req, _ := http.NewRequestWithContext(ctx, "POST", s.cybertServiceURL, bytes.NewBuffer(jsonPayload))
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&result)

	return result, nil
}

func (s *AIService) ExecuteCounterAttack(ctx context.Context, req models.CounterAttackRequest, config *config.Config) error {
	var repo domain.IncidentRepository

	// Валидация запроса
	if req.TargetIP == "" {
		return errors.New("target IP is required")
	}

	if req.Intensity < 1 || req.Intensity > 5 {
		return errors.New("intensity must be between 1 and 5")
	}

	// Логирование действия
	log.Printf("Executing counter attack: %+v", req)

	// Выполнение контратаки через низкоуровневый модуль
	return engine.NewDetector(config.Detection, repo).RespondToThreat(req.TargetIP, req.Intensity)
}

// Вспомогательная функция для инкремента версии
func incrementVersion(version string) float64 {
	if version == "" {
		return 1.0
	}
	v, err := strconv.ParseFloat(version, 64)
	if err != nil {
		return 1.0
	}
	return v + 0.1
}

// Вспомогательные методы
func (s *AIService) calculateThreatProbability(traffic []models.NetworkTraffic) float64 {
	// Упрощенная логика расчета вероятности угрозы
	// В реальном приложении здесь будет сложный алгоритм анализа
	suspiciousCount := 0
	for _, t := range traffic {
		if t.DstPort == 4444 || t.DstPort == 8080 || t.DstPort == 1337 {
			suspiciousCount++
		}
		if t.Protocol == "IRC" || t.Protocol == "TOR" {
			suspiciousCount++
		}
	}

	if len(traffic) == 0 {
		return 0
	}

	return float64(suspiciousCount) / float64(len(traffic))
}

func (s *AIService) detectPatterns(traffic []models.NetworkTraffic) []string {
	// Упрощенная логика обнаружения шаблонов атак
	patterns := make([]string, 0)

	// Проверяем наличие сканирования портов
	portScanDetected := false
	for _, t := range traffic {
		if t.BytesSent < 100 && t.Duration < 0.1 {
			portScanDetected = true
			break
		}
	}
	if portScanDetected {
		patterns = append(patterns, "T1046") // Port Scanning
	}

	// Проверяем наличие C2 коммуникаций
	c2Detected := false
	for _, t := range traffic {
		if t.DstPort == 4444 || t.DstPort == 8080 {
			c2Detected = true
			break
		}
	}
	if c2Detected {
		patterns = append(patterns, "T1071") // C2 Communication
	}

	return patterns
}

func (s *AIService) determineThreatType(traffic []models.NetworkTraffic) string {
	// Упрощенная логика определения типа угрозы
	for _, t := range traffic {
		if t.DstPort == 4444 {
			return "APT"
		}
		if t.Protocol == "IRC" {
			return "Botnet"
		}
	}
	return "Unknown"
}

func (s *AIService) extractIndicators(traffic []models.NetworkTraffic) []string {
	// Упрощенная логика извлечения индикаторов компрометации
	indicators := make([]string, 0)

	for _, t := range traffic {
		if t.DstPort == 4444 {
			indicators = append(indicators, "C2 Communication")
		}
		if t.BytesSent > 1000000 {
			indicators = append(indicators, "Data Exfiltration")
		}
	}

	return indicators
}

func (s *AIService) determineAPTStage(report models.ThreatReport) string {
	// Определение стадии APT атаки на основе обнаруженных шаблонов
	for _, pattern := range report.DetectedPatterns {
		switch pattern {
		case "T1046", "T1595": // Port Scanning, Active Scanning
			return "Reconnaissance"
		case "T1190", "T1133": // Exploit Public-Facing Application, External Remote Services
			return "Initial Access"
		case "T1059", "T1053": // Command and Scripting Interpreter, Scheduled Task/Job
			return "Execution"
		case "T1078", "T1098": // Valid Accounts, Account Manipulation
			return "Persistence"
		case "T1134", "T1068": // Access Token Manipulation, Exploitation for Privilege Escalation
			return "Privilege Escalation"
		case "T1027", "T1140": // Obfuscated Files or Information, Deobfuscate/Decode Files or Information
			return "Defense Evasion"
		case "T1071", "T1105": // Application Layer Protocol, Ingress Tool Transfer
			return "Command and Control"
		case "T1048", "T1567": // Exfiltration Over Alternative Protocol, Exfiltration Over Web Service
			return "Exfiltration"
		case "T1485", "T1486": // Data Destruction, Data Encrypted for Impact
			return "Impact"
		}
	}
	return "Unknown"
}

func (s *AIService) getMainIndicator(report models.ThreatReport) string {
	if len(report.Indicators) > 0 {
		return report.Indicators[0]
	}
	return "Unknown"
}

func (s *AIService) calculateThreatLevel(data []string) int {
	// Простая логика определения уровня угрозы
	if len(data) == 0 {
		return 0
	}

	threatLevel := 0
	for _, item := range data {
		if strings.Contains(strings.ToLower(item), "error") {
			threatLevel += 1
		}
		if strings.Contains(strings.ToLower(item), "warning") {
			threatLevel += 1
		}
		if strings.Contains(strings.ToLower(item), "critical") {
			threatLevel += 2
		}
		if strings.Contains(strings.ToLower(item), "attack") {
			threatLevel += 3
		}
	}

	// Ограничиваем максимальный уровень угрозы
	if threatLevel > 5 {
		threatLevel = 5
	}

	return threatLevel
}
