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

func Post_api_2_user_avatar_temporaryHandler(cfg *config.APIConfig) func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		args, ok := request.Params.Arguments.(map[string]any)
		if !ok {
			return mcp.NewToolResultError("Invalid arguments object"), nil
		}
		queryParams := make([]string, 0)
		if val, ok := args["username"]; ok {
			queryParams = append(queryParams, fmt.Sprintf("username=%v", val))
		}
		queryString := ""
		if len(queryParams) > 0 {
			queryString = "?" + strings.Join(queryParams, "&")
		}
		url := fmt.Sprintf("%s/api/2/user/avatar/temporary%s", cfg.BaseURL, queryString)
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

func CreatePost_api_2_user_avatar_temporaryTool(cfg *config.APIConfig) models.Tool {
	tool := mcp.NewTool("post_api_2_user_avatar_temporary",
		mcp.WithDescription("Creates temporary avatar using multipart. The response is sent back as JSON stored in a textarea. This is because
 the client uses remote iframing to submit avatars using multipart. So we must send them a valid HTML page back from
 which the client parses the JSON from.
 <p>
 Creating a temporary avatar is part of a 3-step process in uploading a new
 avatar for a user: upload, crop, confirm. This endpoint allows you to use a multipart upload
 instead of sending the image directly as the request body.
 </p>
 <p>
 You *must* use "avatar" as the name of the upload parameter:</p>
 <p/>
 <pre>
 curl -c cookiejar.txt -X POST -u admin:admin -H "X-Atlassian-Token: no-check" \
   -F "avatar=@mynewavatar.png;type=image/png" \
   'http://localhost:8090/jira/rest/api/2/user/avatar/temporary?username=admin'
 </pre>"),
		mcp.WithString("username", mcp.Description("Username")),
	)

	return models.Tool{
		Definition: tool,
		Handler:    Post_api_2_user_avatar_temporaryHandler(cfg),
	}
}
