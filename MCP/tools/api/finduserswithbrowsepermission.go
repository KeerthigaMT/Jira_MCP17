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

func FinduserswithbrowsepermissionHandler(cfg *config.APIConfig) func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		args, ok := request.Params.Arguments.(map[string]any)
		if !ok {
			return mcp.NewToolResultError("Invalid arguments object"), nil
		}
		queryParams := make([]string, 0)
		if val, ok := args["username"]; ok {
			queryParams = append(queryParams, fmt.Sprintf("username=%v", val))
		}
		if val, ok := args["issueKey"]; ok {
			queryParams = append(queryParams, fmt.Sprintf("issueKey=%v", val))
		}
		if val, ok := args["projectKey"]; ok {
			queryParams = append(queryParams, fmt.Sprintf("projectKey=%v", val))
		}
		if val, ok := args["startAt"]; ok {
			queryParams = append(queryParams, fmt.Sprintf("startAt=%v", val))
		}
		if val, ok := args["maxResults"]; ok {
			queryParams = append(queryParams, fmt.Sprintf("maxResults=%v", val))
		}
		queryString := ""
		if len(queryParams) > 0 {
			queryString = "?" + strings.Join(queryParams, "&")
		}
		url := fmt.Sprintf("%s/api/2/user/viewissue/search%s", cfg.BaseURL, queryString)
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

func CreateFinduserswithbrowsepermissionTool(cfg *config.APIConfig) models.Tool {
	tool := mcp.NewTool("get_api_2_user_viewissue_search",
		mcp.WithDescription("Returns a list of active users that match the search string. This resource cannot be accessed anonymously 
 and requires the Browse Users global permission.
 Given an issue key this resource will provide a list of users that match the search string and have
 the browse issue permission for the issue provided."),
		mcp.WithString("username", mcp.Description("the username filter, no users returned if left blank")),
		mcp.WithString("issueKey", mcp.Description("the issue key for the issue being edited we need to find viewable users for.")),
		mcp.WithString("projectKey", mcp.Description("the optional project key to search for users with if no issueKey is supplied.")),
		mcp.WithString("startAt", mcp.Description("the index of the first user to return (0-based)")),
		mcp.WithString("maxResults", mcp.Description("the maximum number of users to return (defaults to 50). The maximum allowed value is 1000.\n                   If you specify a value that is higher than this number, your search results will be truncated.")),
	)

	return models.Tool{
		Definition: tool,
		Handler:    FinduserswithbrowsepermissionHandler(cfg),
	}
}
