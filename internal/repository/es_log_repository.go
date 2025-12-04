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
	defer res.Body.Close()

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
	defer res.Body.Close()

	if res.IsError() {
		return fmt.Errorf("bulk indexing failed: %s", res.String())
	}

	return nil
}
