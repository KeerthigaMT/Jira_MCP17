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

func LinkissuesHandler(cfg *config.APIConfig) func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		url := fmt.Sprintf("%s/api/2/issueLink", cfg.BaseURL)
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

func CreateLinkissuesTool(cfg *config.APIConfig) models.Tool {
	tool := mcp.NewTool("post_api_2_issueLink",
		mcp.WithDescription("Creates an issue link between two issues.
 The user requires the link issue permission for the issue which will be linked to another issue.
 The specified link type in the request is used to create the link and will create a link from the first issue
 to the second issue using the outward description. It also create a link from the second issue to the first issue using the
 inward description of the issue link type.
 It will add the supplied comment to the first issue. The comment can have a restriction who can view it.
 If group is specified, only users of this group can view this comment, if roleLevel is specified only users who have the specified role can view this comment.
 The user who creates the issue link needs to belong to the specified group or have the specified role."),
	)

	return models.Tool{
		Definition: tool,
		Handler:    LinkissuesHandler(cfg),
	}
}
