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

func FindassignableusersHandler(cfg *config.APIConfig) func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		args, ok := request.Params.Arguments.(map[string]any)
		if !ok {
			return mcp.NewToolResultError("Invalid arguments object"), nil
		}
		queryParams := make([]string, 0)
		if val, ok := args["username"]; ok {
			queryParams = append(queryParams, fmt.Sprintf("username=%v", val))
		}
		if val, ok := args["project"]; ok {
			queryParams = append(queryParams, fmt.Sprintf("project=%v", val))
		}
		if val, ok := args["issueKey"]; ok {
			queryParams = append(queryParams, fmt.Sprintf("issueKey=%v", val))
		}
		if val, ok := args["startAt"]; ok {
			queryParams = append(queryParams, fmt.Sprintf("startAt=%v", val))
		}
		if val, ok := args["maxResults"]; ok {
			queryParams = append(queryParams, fmt.Sprintf("maxResults=%v", val))
		}
		if val, ok := args["actionDescriptorId"]; ok {
			queryParams = append(queryParams, fmt.Sprintf("actionDescriptorId=%v", val))
		}
		queryString := ""
		if len(queryParams) > 0 {
			queryString = "?" + strings.Join(queryParams, "&")
		}
		url := fmt.Sprintf("%s/api/2/user/assignable/search%s", cfg.BaseURL, queryString)
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

func CreateFindassignableusersTool(cfg *config.APIConfig) models.Tool {
	tool := mcp.NewTool("get_api_2_user_assignable_search",
		mcp.WithDescription("Returns a list of users that match the search string. This resource cannot be accessed anonymously.
 Please note that this resource should be called with an issue key when a list of assignable users is retrieved
 for editing.  For create only a project key should be supplied.  The list of assignable users may be incorrect
 if it's called with the project key for editing."),
		mcp.WithString("username", mcp.Description("the username")),
		mcp.WithString("project", mcp.Description("the key of the project we are finding assignable users for")),
		mcp.WithString("issueKey", mcp.Description("the issue key for the issue being edited we need to find assignable users for.")),
		mcp.WithString("startAt", mcp.Description("the index of the first user to return (0-based)")),
		mcp.WithString("maxResults", mcp.Description("the maximum number of users to return (defaults to 50). The maximum allowed value is 1000.\n                   If you specify a value that is higher than this number, your search results will be truncated.")),
		mcp.WithString("actionDescriptorId", mcp.Description("")),
	)

	return models.Tool{
		Definition: tool,
		Handler:    FindassignableusersHandler(cfg),
	}
}
