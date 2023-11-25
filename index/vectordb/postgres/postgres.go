package postgres

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"github.com/google/uuid"
	"github.com/henomis/lingoose/index"
	"github.com/henomis/lingoose/index/option"
	"github.com/henomis/lingoose/types"
)

var _ index.VectorDB = &DB{}

type DB struct {
	db          *sql.DB
	table       string
	createIndex *CreateIndexOptions
}

type Distance string

const (
	DistanceCosine       Distance = "<=>"
	DistanceInnerProduct Distance = "<#>"
	DistanceEuclidean    Distance = "<->"
)

type CreateIndexOptions struct {
	Dimension uint64
	Distance  Distance
}

type Options struct {
	DB          *sql.DB
	Table       string
	CreateIndex *CreateIndexOptions
}

func New(options Options) *DB {
	return &DB{
		db:          options.DB,
		table:       options.Table,
		createIndex: options.CreateIndex,
	}
}

func (d *DB) IsEmpty(ctx context.Context) (bool, error) {
	err := d.createIndexIfRequired(ctx)
	if err != nil {
		return true, fmt.Errorf("%w: %w", index.ErrInternal, err)
	}

	var count int
	err = d.db.QueryRow(
		fmt.Sprintf("SELECT count(*) FROM %s", d.table),
	).Scan(&count)
	if err != nil {
		return true, err
	}

	return count == 0, nil
}

func (d *DB) Insert(ctx context.Context, datas []index.Data) error {
	err := d.createIndexIfRequired(ctx)
	if err != nil {
		return fmt.Errorf("%w: %w", index.ErrInternal, err)
	}

	var values []string
	for _, data := range datas {
		if data.ID == "" {
			id, errUUID := uuid.NewUUID()
			if errUUID != nil {
				return errUUID
			}
			data.ID = id.String()
		}

		jsonMetadata, marshalErr := json.Marshal(data.Metadata)
		if marshalErr != nil {
			return fmt.Errorf("%w: %w", index.ErrInternal, marshalErr)
		}

		values = append(
			values,
			fmt.Sprintf(
				"('%s','%s', '%s')",
				data.ID,
				floatToValues(data.Values),
				string(jsonMetadata),
			),
		)
	}

	_, err = d.db.ExecContext(
		ctx,
		fmt.Sprintf("INSERT INTO %s (id, embedding, metadata) VALUES %s",
			d.table,
			strings.Join(values, ","),
		),
	)
	if err != nil {
		return fmt.Errorf("%w: %w", index.ErrInternal, err)
	}

	return nil
}

func (d *DB) Search(ctx context.Context, values []float64, options *option.Options) (index.SearchResults, error) {
	return d.similaritySearch(ctx, values, options)
}

func (d *DB) Drop(ctx context.Context) error {
	_, err := d.db.ExecContext(ctx, fmt.Sprintf("DROP TABLE IF EXISTS %s", d.table))
	if err != nil {
		return fmt.Errorf("%w: %w", index.ErrInternal, err)
	}

	return nil
}

func (d *DB) Delete(ctx context.Context, ids []string) error {
	_, err := d.db.ExecContext(
		ctx,
		fmt.Sprintf("DELETE FROM %s WHERE id = ANY($1)", d.table),
		ids,
	)
	if err != nil {
		return fmt.Errorf("%w: %w", index.ErrInternal, err)
	}

	return nil
}

func (d *DB) similaritySearch(
	ctx context.Context,
	values []float64,
	opts *option.Options,
) (index.SearchResults, error) {
	if opts == nil {
		opts = index.GetDefaultOptions()
	}

	if opts.Filter == nil {
		opts.Filter = ""
	}

	queryVector := fmt.Sprintf("embedding %s '%s'", d.createIndex.Distance, floatToValues(values))
	//nolint:gosec
	query := fmt.Sprintf(
		"SELECT id, embedding, metadata, %s AS score FROM %s %s ORDER BY %s LIMIT %d",
		queryVector,
		d.table,
		queryVector,
		opts.Filter,
		opts.TopK,
	)

	rows, err := d.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", index.ErrInternal, err)
	}
	defer rows.Close()

	var results []index.SearchResult
	for rows.Next() {
		var id string
		var embedding string
		var jsonMetadata json.RawMessage
		var score float64
		err = rows.Scan(&id, &embedding, &jsonMetadata, &score)
		if err != nil {
			return nil, fmt.Errorf("%w: %w", index.ErrInternal, err)
		}

		metadata := make(types.Meta)
		err = json.Unmarshal(jsonMetadata, &metadata)
		if err != nil {
			return nil, fmt.Errorf("%w: %w", index.ErrInternal, err)
		}

		var embeddingValues []float64
		embeddingValues, err = valuesToFloats(embedding)
		if err != nil {
			return nil, fmt.Errorf("%w: %w", index.ErrInternal, err)
		}

		result := index.SearchResult{
			Data: index.Data{
				ID:       id,
				Metadata: metadata,
				Values:   embeddingValues,
			},
			Score: score,
		}
		results = append(results, result)
	}

	err = rows.Err()
	if err != nil {
		return nil, fmt.Errorf("%w: %w", index.ErrInternal, err)
	}

	return results, nil
}

func (d *DB) createIndexIfRequired(ctx context.Context) error {
	if d.createIndex == nil {
		return nil
	}

	_, err := d.db.Exec("CREATE EXTENSION IF NOT EXISTS vector")
	if err != nil {
		return fmt.Errorf("%w: %w", index.ErrInternal, err)
	}

	_, err = d.db.ExecContext(
		ctx,
		fmt.Sprintf("CREATE TABLE IF NOT EXISTS %s (id UUID PRIMARY KEY, metadata json, embedding vector(%d))",
			d.table, d.createIndex.Dimension),
	)
	if err != nil {
		return fmt.Errorf("%w: %w", index.ErrInternal, err)
	}

	return nil
}

func floatToValues(floats []float64) string {
	var b strings.Builder
	b.WriteString("[")
	for i, f := range floats {
		if i > 0 {
			b.WriteString(",")
		}
		b.WriteString(strconv.FormatFloat(f, 'f', -1, 64))
	}
	b.WriteString("]")
	return b.String()
}

func valuesToFloats(s string) ([]float64, error) {
	s = strings.Trim(s, "[]")
	parts := strings.Split(s, ",")
	floats := make([]float64, len(parts))
	for i, p := range parts {
		f, err := strconv.ParseFloat(p, 64)
		if err != nil {
			return nil, err
		}
		floats[i] = f
	}
	return floats, nil
}
