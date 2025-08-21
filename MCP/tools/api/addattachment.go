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

func AddattachmentHandler(cfg *config.APIConfig) func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		url := fmt.Sprintf("%s/api/2/issue/%s/attachments", cfg.BaseURL)
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

func CreateAddattachmentTool(cfg *config.APIConfig) models.Tool {
	tool := mcp.NewTool("post_api_2_issue_issueIdOrKey_attachments",
		mcp.WithDescription("Add one or more attachments to an issue.
 <p>
 This resource expects a multipart post. The media-type multipart/form-data is defined in RFC 1867. Most client
 libraries have classes that make dealing with multipart posts simple. For instance, in Java the Apache HTTP Components
 library provides a
 <a href="http://hc.apache.org/httpcomponents-client-ga/httpmime/apidocs/org/apache/http/entity/mime/MultipartEntity.html">MultiPartEntity</a>
 that makes it simple to submit a multipart POST.
 <p>
 In order to protect against XSRF attacks, because this method accepts multipart/form-data, it has XSRF protection
 on it.  This means you must submit a header of X-Atlassian-Token: no-check with the request, otherwise it will be
 blocked.
 <p>
 The name of the multipart/form-data parameter that contains attachments must be "file"
 <p>
 A simple example to upload a file called "myfile.txt" to issue REST-123:
 <pre>curl -D- -u admin:admin -X POST -H "X-Atlassian-Token: no-check" -F "file=@myfile.txt" http://myhost/rest/api/2/issue/TEST-123/attachments</pre>"),
	)

	return models.Tool{
		Definition: tool,
		Handler:    AddattachmentHandler(cfg),
	}
}
