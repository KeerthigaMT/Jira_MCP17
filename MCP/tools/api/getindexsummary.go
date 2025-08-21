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

func GetindexsummaryHandler(cfg *config.APIConfig) func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		url := fmt.Sprintf("%s/api/2/index/summary", cfg.BaseURL)
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

func CreateGetindexsummaryTool(cfg *config.APIConfig) models.Tool {
	tool := mcp.NewTool("get_api_2_index_summary",
		mcp.WithDescription("Summarizes index condition of current node.
 <p/>
 Returned data consists of:
 <ul>
 <li><code>nodeId</code> - Node identifier.</li>
 <li><code>reportTime</code> - Time of this report creation.</li>
 <li><code>issueIndex</code> - Summary of issue index status.</li>
 <li><code>replicationQueues</code> - Map of index replication queues, where
 keys represent nodes from which replication operations came from.</li>
 </ul>
 <p/>
 <code>issueIndex</code> can contain:
 <ul>
 <li><code>indexReadable</code> - If <code>false</code> the end point failed to read data from issue index
 (check JIRA logs for detailed stack trace), otherwise <code>true</code>.
 When <code>false</code> other fields of <code>issueIndex</code> can be not visible.</li>
 <li><code>countInDatabase</code> - Count of issues found in database.</li>
 <li><code>countInIndex</code> - Count of issues found while querying index.</li>
 <li><code>lastUpdatedInDatabase</code> - Time of last update of issue found in database.</li>
 <li><code>lastUpdatedInIndex</code> - Time of last update of issue found while querying index.</li>
 </ul>
 <p/>
 <code>replicationQueues</code>'s map values can contain:
 <ul>
 <li><code>lastConsumedOperation</code> - Last executed index replication operation by current node from sending node's queue.</li>
 <li><code>lastConsumedOperation.id</code> - Identifier of the operation.</li>
 <li><code>lastConsumedOperation.replicationTime</code> - Time when the operation was sent to other nodes.</li>
 <li><code>lastOperationInQueue</code> - Last index replication operation in sending node's queue.</li>
 <li><code>lastOperationInQueue.id</code> - Identifier of the operation.</li>
 <li><code>lastOperationInQueue.replicationTime</code> - Time when the operation was sent to other nodes.</li>
 <li><code>queueSize</code> - Number of operations in queue from sending node to current node.</li>
 </ul>"),
	)

	return models.Tool{
		Definition: tool,
		Handler:    GetindexsummaryHandler(cfg),
	}
}
