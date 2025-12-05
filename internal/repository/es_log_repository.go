package repository

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/Yupoer/logpulse/internal/domain"
	"github.com/elastic/go-elasticsearch/v8"
	"github.com/elastic/go-elasticsearch/v8/esapi"
)

type esLogRepository struct {
	client *elasticsearch.Client
}

func NewESLogRepository(address string) (domain.LogSearchRepository, error) {
	cfg := elasticsearch.Config{
		Addresses: []string{address},
	}
	client, err := elasticsearch.NewClient(cfg)
	if err != nil {
		return nil, err
	}

	// Fail Fast
	res, err := client.Info()
	if err != nil {
		return nil, err
	}
	defer func() { _ = res.Body.Close() }()

	return &esLogRepository{client: client}, nil
}

func (r *esLogRepository) BulkIndex(ctx context.Context, entries []*domain.LogEntry) error {
	if len(entries) == 0 {
		return nil
	}

	var buf bytes.Buffer
	// Bulk API needs special NDJSON format:
	// Action: { "index" : { "_index" : "logs" } } \n
	// Data:   { "field1" : "value1" } \n
	for _, entry := range entries {
		// 1. Action Line (Metadata)
		meta := []byte(fmt.Sprintf(`{ "index" : { "_index" : "logs" } }%s`, "\n"))
		buf.Write(meta)

		// 2. Data Line (Content)
		data, err := json.Marshal(entry)
		if err != nil {
			log.Printf("Failed to marshal log entry for ES: %v", err)
			continue
		}
		buf.Write(data)
		buf.WriteByte('\n') // every line must be separated by a newline
	}

	// 3. Send Request
	req := esapi.BulkRequest{
		Body: bytes.NewReader(buf.Bytes()),
	}

	res, err := req.Do(ctx, r.client)
	if err != nil {
		return err
	}
	defer func() { _ = res.Body.Close() }()

	if res.IsError() {
		return fmt.Errorf("bulk indexing failed: %s", res.String())
	}

	return nil
}

func (r *esLogRepository) Search(ctx context.Context, query string) ([]*domain.LogEntry, error) {
	var buf bytes.Buffer

	// 建構 ES Query DSL (Domain Specific Language)
	// 這裡我們用 multi_match 同時搜尋 message 和 service_name 欄位
	queryJSON := map[string]interface{}{
		"query": map[string]interface{}{
			"multi_match": map[string]interface{}{
				"query":  query,
				"fields": []string{"message", "service_name", "level"},
			},
		},
	}

	if err := json.NewEncoder(&buf).Encode(queryJSON); err != nil {
		return nil, err
	}

	// 執行搜尋
	res, err := r.client.Search(
		r.client.Search.WithContext(ctx),
		r.client.Search.WithIndex("logs"),
		r.client.Search.WithBody(&buf),
		r.client.Search.WithTrackTotalHits(true),
	)
	if err != nil {
		return nil, err
	}
	defer func() { _ = res.Body.Close() }()

	if res.IsError() {
		return nil, fmt.Errorf("search request failed: %s", res.Status())
	}

	// 解析回應 (這是最醜的部分，但必須做)
	var result map[string]interface{}
	if err := json.NewDecoder(res.Body).Decode(&result); err != nil {
		return nil, err
	}

	// 導航到 hits.hits
	hits, ok := result["hits"].(map[string]interface{})["hits"].([]interface{})
	if !ok {
		return []*domain.LogEntry{}, nil // 沒搜到東西
	}

	logs := make([]*domain.LogEntry, 0, len(hits))
	for _, hit := range hits {
		hitMap := hit.(map[string]interface{})
		source := hitMap["_source"]

		// 將 _source (map) 轉回 LogEntry (struct)
		// 這裡為了簡化，我們先轉成 JSON bytes 再轉回 Struct (雖然有效能損耗但程式碼最乾淨)
		tmpBytes, _ := json.Marshal(source)
		var entry domain.LogEntry
		if err := json.Unmarshal(tmpBytes, &entry); err == nil {
			logs = append(logs, &entry)
		}
	}

	return logs, nil
}
