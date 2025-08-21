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

func ListHandler(cfg *config.APIConfig) func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		args, ok := request.Params.Arguments.(map[string]any)
		if !ok {
			return mcp.NewToolResultError("Invalid arguments object"), nil
		}
		queryParams := make([]string, 0)
		if val, ok := args["filter"]; ok {
			queryParams = append(queryParams, fmt.Sprintf("filter=%v", val))
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
		url := fmt.Sprintf("%s/api/2/dashboard%s", cfg.BaseURL, queryString)
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

func CreateListTool(cfg *config.APIConfig) models.Tool {
	tool := mcp.NewTool("get_api_2_dashboard",
		mcp.WithDescription("Returns a list of all dashboards, optionally filtering them."),
		mcp.WithString("filter", mcp.Description("an optional filter that is applied to the list of dashboards. Valid values include\n                        <code>\"favourite\"</code> for returning only favourite dashboards, and <code>\"my\"</code> for returning\n                        dashboards that are owned by the calling user.")),
		mcp.WithString("startAt", mcp.Description("the index of the first dashboard to return (0-based). must be 0 or a multiple of\n                        <code>maxResults</code>")),
		mcp.WithString("maxResults", mcp.Description("a hint as to the the maximum number of dashboards to return in each call. Note that the\n                        JIRA server reserves the right to impose a <code>maxResults</code> limit that is lower than the value that a\n                        client provides, dues to lack or resources or any other condition. When this happens, your results will be\n                        truncated. Callers should always check the returned <code>maxResults</code> to determine the value that is\n                        effectively being used.")),
	)

	return models.Tool{
		Definition: tool,
		Handler:    ListHandler(cfg),
	}
}
