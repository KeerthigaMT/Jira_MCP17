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

func SearchHandler(cfg *config.APIConfig) func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		args, ok := request.Params.Arguments.(map[string]any)
		if !ok {
			return mcp.NewToolResultError("Invalid arguments object"), nil
		}
		queryParams := make([]string, 0)
		if val, ok := args["jql"]; ok {
			queryParams = append(queryParams, fmt.Sprintf("jql=%v", val))
		}
		if val, ok := args["startAt"]; ok {
			queryParams = append(queryParams, fmt.Sprintf("startAt=%v", val))
		}
		if val, ok := args["maxResults"]; ok {
			queryParams = append(queryParams, fmt.Sprintf("maxResults=%v", val))
		}
		if val, ok := args["validateQuery"]; ok {
			queryParams = append(queryParams, fmt.Sprintf("validateQuery=%v", val))
		}
		if val, ok := args["fields"]; ok {
			queryParams = append(queryParams, fmt.Sprintf("fields=%v", val))
		}
		if val, ok := args["expand"]; ok {
			queryParams = append(queryParams, fmt.Sprintf("expand=%v", val))
		}
		queryString := ""
		if len(queryParams) > 0 {
			queryString = "?" + strings.Join(queryParams, "&")
		}
		url := fmt.Sprintf("%s/api/2/search%s", cfg.BaseURL, queryString)
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

func CreateSearchTool(cfg *config.APIConfig) models.Tool {
	tool := mcp.NewTool("get_api_2_search",
		mcp.WithDescription("Searches for issues using JQL.
 <p>
 <b>Sorting</b>
 the <code>jql</code> parameter is a full <a href="http://confluence.atlassian.com/display/JIRA/Advanced+Searching">JQL</a>
 expression, and includes an <code>ORDER BY</code> clause.
 </p>
 <p>
 The <code>fields</code> param (which can be specified multiple times) gives a comma-separated list of fields
 to include in the response. This can be used to retrieve a subset of fields.
 A particular field can be excluded by prefixing it with a minus.
 <p>
 By default, only navigable (<code>*navigable</code>) fields are returned in this search resource. Note: the default is different
 in the get-issue resource -- the default there all fields (<code>*all</code>).
 <ul>
 <li><code>*all</code> - include all fields</li>
 <li><code>*navigable</code> - include just navigable fields</li>
 <li><code>summary,comment</code> - include just the summary and comments</li>
 <li><code>-description</code> - include navigable fields except the description (the default is <code>*navigable</code> for search)</li>
 <li><code>*all,-comment</code> - include everything except comments</li>
 </ul>
 <p>
 </p>
 <p><b>GET vs POST:</b>
 If the JQL query is too large to be encoded as a query param you should instead
 POST to this resource.
 </p>
 <p>
 <b>Expanding Issues in the Search Result:</b>
 It is possible to expand the issues returned by directly specifying the expansion on the expand parameter passed
 in to this resources.
 </p>
 <p>
 For instance, to expand the &quot;changelog&quot; for all the issues on the search result, it is neccesary to
 specify &quot;changelog&quot; as one of the values to expand.
 </p>"),
		mcp.WithString("jql", mcp.Description("a JQL query string")),
		mcp.WithString("startAt", mcp.Description("the index of the first issue to return (0-based)")),
		mcp.WithString("maxResults", mcp.Description("the maximum number of issues to return (defaults to 50). The maximum allowable value is\n                      dictated by the JIRA property 'jira.search.views.default.max'. If you specify a value that is higher than this\n                      number, your search results will be truncated.")),
		mcp.WithString("validateQuery", mcp.Description("whether to validate the JQL query")),
		mcp.WithString("fields", mcp.Description("the list of fields to return for each issue. By default, all navigable fields are returned.")),
		mcp.WithString("expand", mcp.Description("A comma-separated list of the parameters to expand.")),
	)

	return models.Tool{
		Definition: tool,
		Handler:    SearchHandler(cfg),
	}
}
