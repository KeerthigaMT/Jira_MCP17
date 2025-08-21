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

func PolicycheckcreateuserHandler(cfg *config.APIConfig) func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		url := fmt.Sprintf("%s/api/2/password/policy/createUser", cfg.BaseURL)
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

func CreatePolicycheckcreateuserTool(cfg *config.APIConfig) models.Tool {
	tool := mcp.NewTool("post_api_2_password_policy_createUser",
		mcp.WithDescription("Returns a list of statements explaining why the password policy would disallow a proposed password for a new user.
 <p>
 You can use this method to test the password policy validation. This could be done prior to an action 
 where a new user and related password are created, using methods like the ones in 
 <a href="https://docs.atlassian.com/jira/latest/com/atlassian/jira/bc/user/UserService.html">UserService</a>.      
 For example, you could use this to validate a password in a create user form in the user interface, as the user enters it.<br/>
 The username and new password must be not empty to perform the validation.<br/>
 Note, this method will help you validate against the policy only. It won't check any other validations that might be performed 
 when creating a new user, e.g. checking whether a user with the same name already exists.
 </p>"),
	)

	return models.Tool{
		Definition: tool,
		Handler:    PolicycheckcreateuserHandler(cfg),
	}
}
