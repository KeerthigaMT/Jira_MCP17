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

func Get_api_2_notificationscheme_idHandler(cfg *config.APIConfig) func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		args, ok := request.Params.Arguments.(map[string]any)
		if !ok {
			return mcp.NewToolResultError("Invalid arguments object"), nil
		}
		queryParams := make([]string, 0)
		if val, ok := args["expand"]; ok {
			queryParams = append(queryParams, fmt.Sprintf("expand=%v", val))
		}
		queryString := ""
		if len(queryParams) > 0 {
			queryString = "?" + strings.Join(queryParams, "&")
		}
		url := fmt.Sprintf("%s/api/2/notificationscheme/%s%s", cfg.BaseURL, queryString)
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

func CreateGet_api_2_notificationscheme_idTool(cfg *config.APIConfig) models.Tool {
	tool := mcp.NewTool("get_api_2_notificationscheme_id",
		mcp.WithDescription("Returns a full representation of the notification scheme for the given id. This resource will return a
 notification scheme containing a list of events and recipient configured to receive notifications for these events. Consumer
 should allow events without recipients to appear in response. User accessing
 the data is required to have permissions to administer at least one project associated with the requested notification scheme.
 <p>
 Notification recipients can be:
 <ul>
 <li>current assignee - the value of the notificationType is CurrentAssignee</li>
 <li>issue reporter - the value of the notificationType is Reporter</li>
 <li>current user - the value of the notificationType is CurrentUser</li>
 <li>project lead - the value of the notificationType is ProjectLead</li>
 <li>component lead - the value of the notificationType is ComponentLead</li>
 <li>all watchers - the value of the notification type is AllWatchers</li>
 <li>configured user - the value of the notification type is User. Parameter will contain key of the user. Information about the user will be provided
 if <b>user</b> expand parameter is used. </li>
 <li>configured group - the value of the notification type is Group. Parameter will contain name of the group. Information about the group will be provided
 if <b>group</b> expand parameter is used. </li>
 <li>configured email address - the value of the notification type is EmailAddress, additionally information about the email will be provided.</li>
 <li>users or users in groups in the configured custom fields - the value of the notification type is UserCustomField or GroupCustomField. Parameter
 will contain id of the custom field. Information about the field will be provided if <b>field</b> expand parameter is used. </li>
 <li>configured project role - the value of the notification type is ProjectRole. Parameter will contain project role id. Information about the project role
 will be provided if <b>projectRole</b> expand parameter is used. </li>
 </ul>
 Please see the example for reference.
 </p>
 The events can be JIRA system events or events configured by administrator. In case of the system events, data about theirs
 ids, names and descriptions is provided. In case of custom events, the template event is included as well."),
		mcp.WithString("expand", mcp.Description("")),
	)

	return models.Tool{
		Definition: tool,
		Handler:    Get_api_2_notificationscheme_idHandler(cfg),
	}
}
