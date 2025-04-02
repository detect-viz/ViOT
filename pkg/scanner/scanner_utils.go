package scanner

import (
	"bufio"
	"context"
	"fmt"
	"os/exec"
	"strings"
	"sync"
	"time"

	"go.uber.org/zap"
)

// FastPingScan 使用 fping 快速掃描大量 IP
func FastPingScan(ctx context.Context, logger *zap.Logger, ipRanges []string, timeout time.Duration) ([]string, error) {
	logger.Info("開始使用 fping 進行快速掃描",
		zap.Strings("ipRanges", ipRanges),
		zap.Duration("timeout", timeout))

	var aliveIPs []string
	var mutex sync.Mutex
	var wg sync.WaitGroup

	// 檢查 fping 是否可用
	_, err := exec.LookPath("fping")
	if err != nil {
		return nil, fmt.Errorf("fping 命令不可用，請安裝 fping: %w", err)
	}

	// 對每個 IP 範圍進行掃描
	for _, ipRange := range ipRanges {
		wg.Add(1)
		go func(ipRange string) {
			defer wg.Done()

			// 展開 IP 範圍為 IP 列表
			ips, err := ExpandIPRange(ipRange)
			if err != nil {
				logger.Error("解析 IP 範圍失敗",
					zap.String("range", ipRange),
					zap.Error(err))
				return
			}

			// 將 IP 列表分批處理，每批最多 500 個
			const batchSize = 500
			for i := 0; i < len(ips); i += batchSize {
				end := i + batchSize
				if end > len(ips) {
					end = len(ips)
				}
				batch := ips[i:end]

				logger.Debug("處理 IP 批次",
					zap.Int("batch", i/batchSize+1),
					zap.Int("batchSize", len(batch)),
					zap.Int("totalIPs", len(ips)))

				// 啟動 fping 命令
				args := []string{
					"-a",      // 只顯示活躍的主機
					"-q",      // 安靜模式，減少輸出
					"-r", "1", // 重試次數
					"-t", fmt.Sprintf("%d", int(timeout.Milliseconds())), // 超時時間（毫秒）
				}
				args = append(args, batch...)

				cmd := exec.CommandContext(ctx, "fping", args...)
				stdout, err := cmd.StdoutPipe()
				if err != nil {
					logger.Error("創建 fping 輸出管道失敗", zap.Error(err))
					continue
				}

				if err := cmd.Start(); err != nil {
					logger.Error("啟動 fping 命令失敗", zap.Error(err))
					continue
				}

				// 讀取活躍 IP
				scanner := bufio.NewScanner(stdout)
				for scanner.Scan() {
					line := scanner.Text()
					ip := strings.TrimSpace(strings.Split(line, " ")[0])
					if ip != "" {
						mutex.Lock()
						aliveIPs = append(aliveIPs, ip)
						mutex.Unlock()
						logger.Debug("發現活躍 IP", zap.String("ip", ip))
					}
				}

				if err := cmd.Wait(); err != nil {
					// fping 即使找到主機也會返回非零退出碼
					logger.Debug("fping 命令完成", zap.Error(err))
				}
			}
		}(ipRange)
	}

	wg.Wait()
	logger.Info("快速掃描完成", zap.Int("aliveIPs", len(aliveIPs)))
	return aliveIPs, nil
}
