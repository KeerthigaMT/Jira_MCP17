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

func DeleteactorHandler(cfg *config.APIConfig) func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		args, ok := request.Params.Arguments.(map[string]any)
		if !ok {
			return mcp.NewToolResultError("Invalid arguments object"), nil
		}
		queryParams := make([]string, 0)
		if val, ok := args["user"]; ok {
			queryParams = append(queryParams, fmt.Sprintf("user=%v", val))
		}
		if val, ok := args["group"]; ok {
			queryParams = append(queryParams, fmt.Sprintf("group=%v", val))
		}
		queryString := ""
		if len(queryParams) > 0 {
			queryString = "?" + strings.Join(queryParams, "&")
		}
		url := fmt.Sprintf("%s/api/2/project/%s/role/%s%s", cfg.BaseURL, queryString)
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

func CreateDeleteactorTool(cfg *config.APIConfig) models.Tool {
	tool := mcp.NewTool("delete_api_2_project_projectIdOrKey_role_id",
		mcp.WithDescription("Deletes actors (users or groups) from a project role.
 <p>
 <ul>
 <li>Delete a user from the role: <code>/rest/api/2/project/{projectIdOrKey}/role/{roleId}?user={username}</code></li>
 <li>Delete a group from the role: <code>/rest/api/2/project/{projectIdOrKey}/role/{roleId}?group={groupname}</code></li>
 </ul>"),
		mcp.WithString("user", mcp.Description("the username to remove from the project role")),
		mcp.WithString("group", mcp.Description("the groupname to remove from the project role")),
	)

	return models.Tool{
		Definition: tool,
		Handler:    DeleteactorHandler(cfg),
	}
}
