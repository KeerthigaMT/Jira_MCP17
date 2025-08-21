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

func ReindexHandler(cfg *config.APIConfig) func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		args, ok := request.Params.Arguments.(map[string]any)
		if !ok {
			return mcp.NewToolResultError("Invalid arguments object"), nil
		}
		queryParams := make([]string, 0)
		if val, ok := args["type"]; ok {
			queryParams = append(queryParams, fmt.Sprintf("type=%v", val))
		}
		if val, ok := args["indexComments"]; ok {
			queryParams = append(queryParams, fmt.Sprintf("indexComments=%v", val))
		}
		if val, ok := args["indexChangeHistory"]; ok {
			queryParams = append(queryParams, fmt.Sprintf("indexChangeHistory=%v", val))
		}
		if val, ok := args["indexWorklogs"]; ok {
			queryParams = append(queryParams, fmt.Sprintf("indexWorklogs=%v", val))
		}
		queryString := ""
		if len(queryParams) > 0 {
			queryString = "?" + strings.Join(queryParams, "&")
		}
		url := fmt.Sprintf("%s/api/2/reindex%s", cfg.BaseURL, queryString)
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

func CreateReindexTool(cfg *config.APIConfig) models.Tool {
	tool := mcp.NewTool("post_api_2_reindex",
		mcp.WithDescription("Kicks off a reindex.  Need Admin permissions to perform this reindex."),
		mcp.WithString("type", mcp.Description("Case insensitive String indicating type of reindex.  If omitted, then defaults to BACKGROUND_PREFERRED.")),
		mcp.WithString("indexComments", mcp.Description("Indicates that comments should also be reindexed. Not relevant for foreground reindex, where comments are always reindexed.")),
		mcp.WithString("indexChangeHistory", mcp.Description("Indicates that changeHistory should also be reindexed. Not relevant for foreground reindex, where changeHistory is always reindexed.")),
		mcp.WithString("indexWorklogs", mcp.Description("Indicates that changeHistory should also be reindexed. Not relevant for foreground reindex, where changeHistory is always reindexed.")),
	)

	return models.Tool{
		Definition: tool,
		Handler:    ReindexHandler(cfg),
	}
}
