package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/jira-7-6-1/mcp-server/config"
	"github.com/jira-7-6-1/mcp-server/models"
	"github.com/mark3labs/mcp-go/mcp"
)

func DeleteworklogHandler(cfg *config.APIConfig) func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		args, ok := request.Params.Arguments.(map[string]any)
		if !ok {
			return mcp.NewToolResultError("Invalid arguments object"), nil
		}
		queryParams := make([]string, 0)
		if val, ok := args["adjustEstimate"]; ok {
			queryParams = append(queryParams, fmt.Sprintf("adjustEstimate=%v", val))
		}
		if val, ok := args["newEstimate"]; ok {
			queryParams = append(queryParams, fmt.Sprintf("newEstimate=%v", val))
		}
		if val, ok := args["increaseBy"]; ok {
			queryParams = append(queryParams, fmt.Sprintf("increaseBy=%v", val))
		}
		queryString := ""
		if len(queryParams) > 0 {
			queryString = "?" + strings.Join(queryParams, "&")
		}
		url := fmt.Sprintf("%s/api/2/issue/%s/worklog/%s%s", cfg.BaseURL, queryString)
		req, err := http.NewRequest("DELETE", url, nil)
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

func CreateDeleteworklogTool(cfg *config.APIConfig) models.Tool {
	tool := mcp.NewTool("delete_api_2_issue_issueIdOrKey_worklog_id",
		mcp.WithDescription("Deletes an existing worklog entry."),
		mcp.WithString("adjustEstimate", mcp.Description("(optional) allows you to provide specific instructions to update the remaining time estimate of the issue.  Valid values are\n                       <ul>\n                       <li>\"new\" - sets the estimate to a specific value</li>\n                       <li>\"leave\"- leaves the estimate as is</li>\n                       <li>\"manual\" - specify a specific amount to increase remaining estimate by</li>\n                       <li>\"auto\"- Default option.  Will automatically adjust the value based on the new timeSpent specified on the worklog</li> </ul>")),
		mcp.WithString("newEstimate", mcp.Description("(required when \"new\" is selected for adjustEstimate) the new value for the remaining estimate field. e.g. \"2d\"")),
		mcp.WithString("increaseBy", mcp.Description("(required when \"manual\" is selected for adjustEstimate) the amount to increase the remaining estimate by e.g. \"2d\"")),
	)

	return models.Tool{
		Definition: tool,
		Handler:    DeleteworklogHandler(cfg),
	}
}
