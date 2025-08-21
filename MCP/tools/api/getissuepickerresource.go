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

func GetissuepickerresourceHandler(cfg *config.APIConfig) func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		args, ok := request.Params.Arguments.(map[string]any)
		if !ok {
			return mcp.NewToolResultError("Invalid arguments object"), nil
		}
		queryParams := make([]string, 0)
		if val, ok := args["query"]; ok {
			queryParams = append(queryParams, fmt.Sprintf("query=%v", val))
		}
		if val, ok := args["currentJQL"]; ok {
			queryParams = append(queryParams, fmt.Sprintf("currentJQL=%v", val))
		}
		if val, ok := args["currentIssueKey"]; ok {
			queryParams = append(queryParams, fmt.Sprintf("currentIssueKey=%v", val))
		}
		if val, ok := args["currentProjectId"]; ok {
			queryParams = append(queryParams, fmt.Sprintf("currentProjectId=%v", val))
		}
		if val, ok := args["showSubTasks"]; ok {
			queryParams = append(queryParams, fmt.Sprintf("showSubTasks=%v", val))
		}
		if val, ok := args["showSubTaskParent"]; ok {
			queryParams = append(queryParams, fmt.Sprintf("showSubTaskParent=%v", val))
		}
		queryString := ""
		if len(queryParams) > 0 {
			queryString = "?" + strings.Join(queryParams, "&")
		}
		url := fmt.Sprintf("%s/api/2/issue/picker%s", cfg.BaseURL, queryString)
		req, err := http.NewRequest("GET", url, nil)
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

func CreateGetissuepickerresourceTool(cfg *config.APIConfig) models.Tool {
	tool := mcp.NewTool("get_api_2_issue_picker",
		mcp.WithDescription("Returns suggested issues which match the auto-completion query for the user which executes this request. This REST
 method will check the user's history and the user's browsing context and select this issues, which match the query."),
		mcp.WithString("query", mcp.Description("the query.")),
		mcp.WithString("currentJQL", mcp.Description("the JQL in context of which the request is executed. Only issues which match this JQL query will be included in results.")),
		mcp.WithString("currentIssueKey", mcp.Description("the key of the issue in context of which the request is executed. The issue which is in context will not be included in the auto-completion result, even if it matches the query.")),
		mcp.WithString("currentProjectId", mcp.Description("the id of the project in context of which the request is executed. Suggested issues will be only from this project.")),
		mcp.WithString("showSubTasks", mcp.Description("if set to false, subtasks will not be included in the list.")),
		mcp.WithString("showSubTaskParent", mcp.Description("if set to false and request is executed in context of a subtask, the parent issue will not be included in the auto-completion result, even if it matches the query.")),
	)

	return models.Tool{
		Definition: tool,
		Handler:    GetissuepickerresourceHandler(cfg),
	}
}
