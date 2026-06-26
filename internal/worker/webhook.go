package worker

import (
	"bytes"
	"encoding/json"
	"net/http"
	"time"

	"github.com/code-execution-engine/internal/types"
	"github.com/code-execution-engine/pkg/telemetry"
)

// Webhook payload schema
type webhookPayload struct {
	JobID      string                 `json:"job_id"`
	Status     types.Status           `json:"status"`
	Output     string                 `json:"output,omitempty"`
	Error      string                 `json:"error,omitempty"`
	FatalError string                 `json:"fatal_error,omitempty"`
	TimeMs     int64                  `json:"time_ms,omitempty"`
	MemoryKB   int64                  `json:"memory_kb,omitempty"`
	Passed     int                    `json:"passed,omitempty"`
	Total      int                    `json:"total,omitempty"`
	TestCases  []types.TestCaseResult `json:"test_cases,omitempty"`
}

func triggerWebhookCallback(callbackURL, jobID string, res types.Result) {
	payload := webhookPayload{
		JobID:      jobID,
		Status:     res.Status,
		Output:     res.Output,
		Error:      res.Error,
		FatalError: res.FatalError,
		TimeMs:     res.TimeMs,
		MemoryKB:   res.MemoryKB,
		Passed:     res.Passed,
		Total:      res.Total,
		TestCases:  res.TestCases,
	}

	bytesPayload, err := json.Marshal(payload)
	if err != nil {
		telemetry.Error("Failed to marshal webhook payload", "job_id", jobID, "error", err)
		return
	}

	client := &http.Client{
		Timeout: 5 * time.Second,
	}

	req, err := http.NewRequest("POST", callbackURL, bytes.NewBuffer(bytesPayload))
	if err != nil {
		telemetry.Error("Failed to create webhook request", "job_id", jobID, "error", err)
		return
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "Code-Execution-Engine-Webhook/1.0")

	resp, err := client.Do(req)
	if err != nil {
		telemetry.Error("Webhook callback delivery failed", "job_id", jobID, "url", callbackURL, "error", err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		telemetry.Error("Webhook callback responded with non-2xx status code", "job_id", jobID, "url", callbackURL, "status", resp.Status)
	} else {
		telemetry.Info("Webhook callback delivered successfully", "job_id", jobID, "url", callbackURL, "status", resp.Status)
	}
}
