package opensearch

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/opensearch-project/opensearch-go/v2"
	"github.com/opensearch-project/opensearch-go/v2/opensearchapi"
)

// Client is a wrapper around the official OpenSearch client
type Client struct {
	os *opensearch.Client
}

// NewClient initializes a connection to OpenSearch
func NewClient(url, username, password string) (*Client, error) {
	// Initialize the client
	client, err := opensearch.NewClient(opensearch.Config{
		Addresses: []string{url},
		Username:  username,
		Password:  password,
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true}, // Allow self-signed certs for local dev
		},
	})

	if err != nil {
		return nil, fmt.Errorf("failed to create OpenSearch client: %w", err)
	}

	// Ping cluster to verify connection
	info, err := client.Info()
	if err != nil {
		return nil, fmt.Errorf("failed to connect to OpenSearch cluster: %w", err)
	}
	defer info.Body.Close()

	if info.IsError() {
		return nil, fmt.Errorf("opensearch error: %s", info.String())
	}

	return &Client{os: client}, nil
}

// IndexDocument indexes a single JSON document into a specific index
func (c *Client) IndexDocument(ctx context.Context, index string, documentID string, body interface{}) error {
	data, err := json.Marshal(body)
	if err != nil {
		return fmt.Errorf("failed to marshal body: %w", err)
	}

	req := opensearchapi.IndexRequest{
		Index:      index,
		DocumentID: documentID,
		Body:       strings.NewReader(string(data)),
		Refresh:    "true", // Make it instantly searchable (can be tuned for performance)
	}

	res, err := req.Do(ctx, c.os)
	if err != nil {
		return fmt.Errorf("index request failed: %w", err)
	}
	defer res.Body.Close()

	if res.IsError() {
		return fmt.Errorf("opensearch index error: %s", res.String())
	}

	return nil
}

// Search executes a search query against an index
func (c *Client) Search(ctx context.Context, index string, query map[string]interface{}) ([]map[string]interface{}, int64, error) {
	queryBody, err := json.Marshal(query)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to marshal query: %w", err)
	}

	req := opensearchapi.SearchRequest{
		Index:   []string{index},
		Body:    strings.NewReader(string(queryBody)),
		Timeout: 5 * time.Second,
	}

	res, err := req.Do(ctx, c.os)
	if err != nil {
		return nil, 0, fmt.Errorf("search request failed: %w", err)
	}
	defer res.Body.Close()

	if res.IsError() {
		return nil, 0, fmt.Errorf("opensearch search error: %s", res.String())
	}

	var responseBody map[string]interface{}
	if err := json.NewDecoder(res.Body).Decode(&responseBody); err != nil {
		return nil, 0, fmt.Errorf("failed to decode response: %w", err)
	}

	// Extract hits
	hitsObj, ok := responseBody["hits"].(map[string]interface{})
	if !ok {
		return nil, 0, nil // No hits
	}

	var total int64
	if totalObj, ok := hitsObj["total"].(map[string]interface{}); ok {
		if val, ok := totalObj["value"].(float64); ok {
			total = int64(val)
		}
	}

	hitsList, ok := hitsObj["hits"].([]interface{})
	if !ok {
		return nil, total, nil
	}

	var results []map[string]interface{}
	for _, hitIntf := range hitsList {
		if hit, ok := hitIntf.(map[string]interface{}); ok {
			if source, ok := hit["_source"].(map[string]interface{}); ok {
				results = append(results, source)
			}
		}
	}

	return results, total, nil
}
