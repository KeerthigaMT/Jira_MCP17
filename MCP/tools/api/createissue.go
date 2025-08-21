package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/jira-7-6-1/mcp-server/config"
	"github.com/jira-7-6-1/mcp-server/models"
	"github.com/mark3labs/mcp-go/mcp"
)

func CreateissueHandler(cfg *config.APIConfig) func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		url := fmt.Sprintf("%s/api/2/issue", cfg.BaseURL)
		req, err := http.NewRequest("POST", url, nil)
		if err != nil {
			return mcp.NewToolResultErrorFromErr("Failed to create request", err), nil
		}
		// No authentication required for this endpoint
		req.Header.Set("Accept", "application/json")

		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			return mcp.NewToolResultErrorFromErr("Request failed", err), nil
		}
		defer resp.Body.Close()

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return mcp.NewToolResultErrorFromErr("Failed to read response body", err), nil
		}

		if resp.StatusCode >= 400 {
			return mcp.NewToolResultError(fmt.Sprintf("API error: %s", body)), nil
		}
		// Use properly typed response
		var result map[string]interface{}
		if err := json.Unmarshal(body, &result); err != nil {
			// Fallback to raw text if unmarshaling fails
			return mcp.NewToolResultText(string(body)), nil
		}

		prettyJSON, err := json.MarshalIndent(result, "", "  ")
		if err != nil {
			return mcp.NewToolResultErrorFromErr("Failed to format JSON", err), nil
		}

		return mcp.NewToolResultText(string(prettyJSON)), nil
	}
}

func CreateCreateissueTool(cfg *config.APIConfig) models.Tool {
	tool := mcp.NewTool("post_api_2_issue",
		mcp.WithDescription("Creates an issue or a sub-task from a JSON representation.
 <p/>
 The fields that can be set on create, in either the fields parameter or the update parameter can be determined
 using the <b>/rest/api/2/issue/createmeta</b> resource.
 If a field is not configured to appear on the create screen, then it will not be in the createmeta, and a field
 validation error will occur if it is submitted.
 <p/>
 Creating a sub-task is similar to creating a regular issue, with two important differences:
 <ul>
 <li>the <code>issueType</code> field must correspond to a sub-task issue type (you can use
 <code>/issue/createmeta</code> to discover sub-task issue types), and</li>
 <li>you must provide a <code>parent</code> field in the issue create request containing the id or key of the
 parent issue.</li>
 </ul>"),
	)

	return models.Tool{
		Definition: tool,
		Handler:    CreateissueHandler(cfg),
	}
}
