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

func EditissueHandler(cfg *config.APIConfig) func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		args, ok := request.Params.Arguments.(map[string]any)
		if !ok {
			return mcp.NewToolResultError("Invalid arguments object"), nil
		}
		queryParams := make([]string, 0)
		if val, ok := args["notifyUsers"]; ok {
			queryParams = append(queryParams, fmt.Sprintf("notifyUsers=%v", val))
		}
		queryString := ""
		if len(queryParams) > 0 {
			queryString = "?" + strings.Join(queryParams, "&")
		}
		url := fmt.Sprintf("%s/api/2/issue/%s%s", cfg.BaseURL, queryString)
		req, err := http.NewRequest("PUT", url, nil)
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

func CreateEditissueTool(cfg *config.APIConfig) models.Tool {
	tool := mcp.NewTool("put_api_2_issue_issueIdOrKey",
		mcp.WithDescription("Edits an issue from a JSON representation.
 <p/>
 The issue can either be updated by setting explicit the field value(s)
 or by using an operation to change the field value.
 <p/>
 The fields that can be updated, in either the fields parameter or the update parameter, can be determined
 using the <b>/rest/api/2/issue/{issueIdOrKey}/editmeta</b> resource.<br>
 If a field is not configured to appear on the edit screen, then it will not be in the editmeta, and a field
 validation error will occur if it is submitted.
 <p/>
 Specifying a "field_id": field_value in the "fields" is a shorthand for a "set" operation in the "update" section.<br>
 Field should appear either in "fields" or "update", not in both."),
		mcp.WithString("notifyUsers", mcp.Description("send the email with notification that the issue was updated to users that watch it.\n                    Admin or project admin permissions are required to disable the notification.")),
	)

	return models.Tool{
		Definition: tool,
		Handler:    EditissueHandler(cfg),
	}
}
