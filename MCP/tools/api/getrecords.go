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

func GetrecordsHandler(cfg *config.APIConfig) func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		args, ok := request.Params.Arguments.(map[string]any)
		if !ok {
			return mcp.NewToolResultError("Invalid arguments object"), nil
		}
		queryParams := make([]string, 0)
		if val, ok := args["offset"]; ok {
			queryParams = append(queryParams, fmt.Sprintf("offset=%v", val))
		}
		if val, ok := args["limit"]; ok {
			queryParams = append(queryParams, fmt.Sprintf("limit=%v", val))
		}
		if val, ok := args["filter"]; ok {
			queryParams = append(queryParams, fmt.Sprintf("filter=%v", val))
		}
		if val, ok := args["from"]; ok {
			queryParams = append(queryParams, fmt.Sprintf("from=%v", val))
		}
		if val, ok := args["to"]; ok {
			queryParams = append(queryParams, fmt.Sprintf("to=%v", val))
		}
		if val, ok := args["projectIds"]; ok {
			queryParams = append(queryParams, fmt.Sprintf("projectIds=%v", val))
		}
		if val, ok := args["userIds"]; ok {
			queryParams = append(queryParams, fmt.Sprintf("userIds=%v", val))
		}
		queryString := ""
		if len(queryParams) > 0 {
			queryString = "?" + strings.Join(queryParams, "&")
		}
		url := fmt.Sprintf("%s/api/2/auditing/record%s", cfg.BaseURL, queryString)
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

func CreateGetrecordsTool(cfg *config.APIConfig) models.Tool {
	tool := mcp.NewTool("get_api_2_auditing_record",
		mcp.WithDescription("Returns auditing records filtered using provided parameters"),
		mcp.WithString("offset", mcp.Description("- the number of record from which search starts")),
		mcp.WithString("limit", mcp.Description("- maximum number of returned results (if is limit is <= 0 or > 1000, it will be set do default value: 1000)")),
		mcp.WithString("filter", mcp.Description("- text query; each record that will be returned must contain the provided text in one of its fields")),
		mcp.WithString("from", mcp.Description("- timestamp in past; 'from' must be less or equal 'to', otherwise the result set will be empty\n               only records that where created in the same moment or after the 'from' timestamp will be provided in response")),
		mcp.WithString("to", mcp.Description("- timestamp in past; 'from' must be less or equal 'to', otherwise the result set will be empty\n               only records that where created in the same moment or earlier than the 'to' timestamp will be provided in response")),
		mcp.WithString("projectIds", mcp.Description("- list of project ids to look for")),
		mcp.WithString("userIds", mcp.Description("- list of user ids to look for")),
	)

	return models.Tool{
		Definition: tool,
		Handler:    GetrecordsHandler(cfg),
	}
}
