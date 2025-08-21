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

func GetcreateissuemetaHandler(cfg *config.APIConfig) func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		args, ok := request.Params.Arguments.(map[string]any)
		if !ok {
			return mcp.NewToolResultError("Invalid arguments object"), nil
		}
		queryParams := make([]string, 0)
		if val, ok := args["projectIds"]; ok {
			queryParams = append(queryParams, fmt.Sprintf("projectIds=%v", val))
		}
		if val, ok := args["projectKeys"]; ok {
			queryParams = append(queryParams, fmt.Sprintf("projectKeys=%v", val))
		}
		if val, ok := args["issuetypeIds"]; ok {
			queryParams = append(queryParams, fmt.Sprintf("issuetypeIds=%v", val))
		}
		if val, ok := args["issuetypeNames"]; ok {
			queryParams = append(queryParams, fmt.Sprintf("issuetypeNames=%v", val))
		}
		queryString := ""
		if len(queryParams) > 0 {
			queryString = "?" + strings.Join(queryParams, "&")
		}
		url := fmt.Sprintf("%s/api/2/issue/createmeta%s", cfg.BaseURL, queryString)
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

func CreateGetcreateissuemetaTool(cfg *config.APIConfig) models.Tool {
	tool := mcp.NewTool("get_api_2_issue_createmeta",
		mcp.WithDescription("Returns the meta data for creating issues. This includes the available projects, issue types and fields,
 including field types and whether or not those fields are required.
 Projects will not be returned if the user does not have permission to create issues in that project.
 <p/>
 The fields in the createmeta correspond to the fields in the create screen for the project/issuetype.
 Fields not in the screen will not be in the createmeta.
 <p/>
 Fields will only be returned if <code>expand=projects.issuetypes.fields</code>.
 <p/>
 The results can be filtered by project and/or issue type, given by the query params."),
		mcp.WithString("projectIds", mcp.Description("combined with the projectKeys param, lists the projects with which to filter the results. If absent, all projects are returned.\n                       This parameter can be specified multiple times, and/or be a comma-separated list.\n                       Specifiying a project that does not exist (or that you cannot create issues in) is not an error, but it will not be in the results.")),
		mcp.WithString("projectKeys", mcp.Description("combined with the projectIds param, lists the projects with which to filter the results. If null, all projects are returned.\n                       This parameter can be specified multiple times, and/or be a comma-separated list.\n                       Specifiying a project that does not exist (or that you cannot create issues in) is not an error, but it will not be in the results.")),
		mcp.WithString("issuetypeIds", mcp.Description("combinded with issuetypeNames, lists the issue types with which to filter the results. If null, all issue types are returned.\n                       This parameter can be specified multiple times, and/or be a comma-separated list.\n                       Specifiying an issue type that does not exist is not an error.")),
		mcp.WithString("issuetypeNames", mcp.Description("combinded with issuetypeIds, lists the issue types with which to filter the results. If null, all issue types are returned.\n                       This parameter can be specified multiple times, but is NOT interpreted as a comma-separated list.\n                       Specifiying an issue type that does not exist is not an error.")),
	)

	return models.Tool{
		Definition: tool,
		Handler:    GetcreateissuemetaHandler(cfg),
	}
}
