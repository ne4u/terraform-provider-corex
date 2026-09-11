package client

import (
	"context"
	"fmt"
	"time"
)

// applyResponse is the response from POST /config/apply.
type applyResponse struct {
	Status  string `json:"status"`
	Message string `json:"message"`
	TaskID  int    `json:"task_id"`
}

// taskResponse is the response from GET /tasks/{id}.
type taskResponse struct {
	ID        int    `json:"id"`
	TaskType  string `json:"task_type"`
	Status    string `json:"status"`
	Error     string `json:"error"`
	Result    map[string]interface{} `json:"result"`
}

// ApplyConfig triggers POST /config/apply and polls the task until completion.
func (c *Client) ApplyConfig(ctx context.Context) error {
	var resp applyResponse
	if err := c.do(ctx, "POST", "/config/apply", map[string]interface{}{}, &resp); err != nil {
		return fmt.Errorf("failed to apply config: %w", err)
	}
	if resp.TaskID == 0 {
		// Some responses may not include a task_id (synchronous apply).
		return nil
	}
	return c.pollTask(ctx, resp.TaskID)
}

// pollTask polls GET /tasks/{id} every 2 seconds until the task is complete.
func (c *Client) pollTask(ctx context.Context, taskID int) error {
	timeout := time.After(60 * time.Second)
	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-timeout:
			return fmt.Errorf("timed out waiting for task %d to complete", taskID)
		case <-ticker.C:
			var task taskResponse
			if err := c.do(ctx, "GET", fmt.Sprintf("/tasks/%d", taskID), nil, &task); err != nil {
				return fmt.Errorf("failed to poll task %d: %w", taskID, err)
			}
			switch task.Status {
			case "success":
				return nil
			case "failed":
				msg := task.Error
				if msg == "" {
					if r, ok := task.Result["message"]; ok {
						msg = fmt.Sprintf("%v", r)
					}
				}
				if msg == "" {
					msg = "unknown error"
				}
				return fmt.Errorf("config apply task %d failed: %s", taskID, msg)
			}
			// status is "pending" or "running" — keep polling
		}
	}
}
