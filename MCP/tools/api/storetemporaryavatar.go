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

func StoretemporaryavatarHandler(cfg *config.APIConfig) func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		args, ok := request.Params.Arguments.(map[string]any)
		if !ok {
			return mcp.NewToolResultError("Invalid arguments object"), nil
		}
		queryParams := make([]string, 0)
		if val, ok := args["filename"]; ok {
			queryParams = append(queryParams, fmt.Sprintf("filename=%v", val))
		}
		if val, ok := args["size"]; ok {
			queryParams = append(queryParams, fmt.Sprintf("size=%v", val))
		}
		queryString := ""
		if len(queryParams) > 0 {
			queryString = "?" + strings.Join(queryParams, "&")
		}
		url := fmt.Sprintf("%s/api/2/avatar/%s/temporary%s", cfg.BaseURL, queryString)
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

func CreateStoretemporaryavatarTool(cfg *config.APIConfig) models.Tool {
	tool := mcp.NewTool("post_api_2_avatar_type_temporary",
		mcp.WithDescription("Creates temporary avatar"),
		mcp.WithString("filename", mcp.Description("name of file being uploaded")),
		mcp.WithString("size", mcp.Description("size of file")),
	)

	return models.Tool{
		Definition: tool,
		Handler:    StoretemporaryavatarHandler(cfg),
	}
}
