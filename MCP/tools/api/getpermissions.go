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

func GetpermissionsHandler(cfg *config.APIConfig) func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		args, ok := request.Params.Arguments.(map[string]any)
		if !ok {
			return mcp.NewToolResultError("Invalid arguments object"), nil
		}
		queryParams := make([]string, 0)
		if val, ok := args["projectKey"]; ok {
			queryParams = append(queryParams, fmt.Sprintf("projectKey=%v", val))
		}
		if val, ok := args["projectId"]; ok {
			queryParams = append(queryParams, fmt.Sprintf("projectId=%v", val))
		}
		if val, ok := args["issueKey"]; ok {
			queryParams = append(queryParams, fmt.Sprintf("issueKey=%v", val))
		}
		if val, ok := args["issueId"]; ok {
			queryParams = append(queryParams, fmt.Sprintf("issueId=%v", val))
		}
		queryString := ""
		if len(queryParams) > 0 {
			queryString = "?" + strings.Join(queryParams, "&")
		}
		url := fmt.Sprintf("%s/api/2/mypermissions%s", cfg.BaseURL, queryString)
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

func CreateGetpermissionsTool(cfg *config.APIConfig) models.Tool {
	tool := mcp.NewTool("get_api_2_mypermissions",
		mcp.WithDescription("Returns all permissions in the system and whether the currently logged in user has them. You can optionally provide a specific context to get permissions for
 (projectKey OR projectId OR issueKey OR issueId)
 <ul>
 <li> When no context supplied the project related permissions will return true if the user has that permission in ANY project </li>
 <li> If a project context is provided, project related permissions will return true if the user has the permissions in the specified project.
 For permissions that are determined using issue data (e.g Current Assignee), true will be returned if the user meets the permission criteria in ANY issue in that project </li>
 <li> If an issue context is provided, it will return whether or not the user has each permission in that specific issue</li>
 </ul>
 <p>
 NB: The above means that for issue-level permissions (EDIT_ISSUE for example), hasPermission may be true when no context is provided, or when a project context is provided,
 <b>but</b> may be false for any given (or all) issues. This would occur (for example) if Reporters were given the EDIT_ISSUE permission. This is because
 any user could be a reporter, except in the context of a concrete issue, where the reporter is known.
 </p>
 <p>
 Global permissions will still be returned for all scopes.
 </p>
 <p>
 Prior to version 6.4 this service returned project permissions with keys corresponding to com.atlassian.jira.security.Permissions.Permission constants.
 Since 6.4 those keys are considered deprecated and this service returns system project permission keys corresponding to constants defined in com.atlassian.jira.permission.ProjectPermissions.
 Permissions with legacy keys are still also returned for backwards compatibility, they are marked with an attribute deprecatedKey=true.
 The attribute is missing for project permissions with the current keys.
 </p>"),
		mcp.WithString("projectKey", mcp.Description("- key of project to scope returned permissions for.")),
		mcp.WithString("projectId", mcp.Description("- id of project to scope returned permissions for.")),
		mcp.WithString("issueKey", mcp.Description("- key of the issue to scope returned permissions for.")),
		mcp.WithString("issueId", mcp.Description("- id of the issue to scope returned permissions for.")),
	)

	return models.Tool{
		Definition: tool,
		Handler:    GetpermissionsHandler(cfg),
	}
}
