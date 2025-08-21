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

func GetpropertiesHandler(cfg *config.APIConfig) func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		args, ok := request.Params.Arguments.(map[string]any)
		if !ok {
			return mcp.NewToolResultError("Invalid arguments object"), nil
		}
		queryParams := make([]string, 0)
		if val, ok := args["includeReservedKeys"]; ok {
			queryParams = append(queryParams, fmt.Sprintf("includeReservedKeys=%v", val))
		}
		if val, ok := args["key"]; ok {
			queryParams = append(queryParams, fmt.Sprintf("key=%v", val))
		}
		if val, ok := args["workflowName"]; ok {
			queryParams = append(queryParams, fmt.Sprintf("workflowName=%v", val))
		}
		if val, ok := args["workflowMode"]; ok {
			queryParams = append(queryParams, fmt.Sprintf("workflowMode=%v", val))
		}
		queryString := ""
		if len(queryParams) > 0 {
			queryString = "?" + strings.Join(queryParams, "&")
		}
		url := fmt.Sprintf("%s/api/2/workflow/api/2/transitions/%s/properties%s", cfg.BaseURL, queryString)
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

func CreateGetpropertiesTool(cfg *config.APIConfig) models.Tool {
	tool := mcp.NewTool("get_api_2_workflow_api_2_transitions_id_properties",
		mcp.WithDescription("Return the property or properties associated with a transition."),
		mcp.WithString("includeReservedKeys", mcp.Description("some keys under the \"jira.\" prefix are editable, some are not. Set this to true\n                            in order to include the non-editable keys in the response.")),
		mcp.WithString("key", mcp.Description("the name of the property key to query. Can be left off the query to return all properties.")),
		mcp.WithString("workflowName", mcp.Description("the name of the workflow to use.")),
		mcp.WithString("workflowMode", mcp.Description("the type of workflow to use. Can either be \"live\" or \"draft\".")),
	)

	return models.Tool{
		Definition: tool,
		Handler:    GetpropertiesHandler(cfg),
	}
}
