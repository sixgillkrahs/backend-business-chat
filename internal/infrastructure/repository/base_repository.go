package repository

import (
	"context"
	"fmt"
	"reflect"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/sixgillkrahs/backend-business-chat/internal/infrastructure/database"
)

type baseRepository[T any] struct {
	db *database.PostgresDB
}

func newBaseRepository[T any](db *database.PostgresDB) baseRepository[T] {
	return baseRepository[T]{
		db: db,
	}
}

// getTableName returns the table name based on the struct type T.
func getTableName(val interface{}) string {
	t := reflect.TypeOf(val)
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	name := strings.ToLower(t.Name())

	if strings.HasSuffix(name, "y") {
		return name[:len(name)-1] + "ies"
	}
	if strings.HasSuffix(name, "s") {
		return name
	}
	return name + "s"
}

// GetAll returns all records of type T from the database.
func (r *baseRepository[T]) GetAll(ctx context.Context, model T) ([]T, error) {
	if r.db == nil {
		return nil, fmt.Errorf("database connection is not initialized")
	}

	tableName := getTableName(model)
	t := reflect.TypeOf(model)
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}

	var columns []string
	var fieldIndexes []int
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		if field.PkgPath != "" {
			continue
		}
		dbTag := field.Tag.Get("db")
		if dbTag == "-" {
			continue
		}
		if dbTag == "" {
			dbTag = field.Tag.Get("json")
		}
		if dbTag == "" || dbTag == "-" {
			dbTag = strings.ToLower(field.Name)
		} else {
			dbTag = strings.Split(dbTag, ",")[0]
		}
		columns = append(columns, dbTag)
		fieldIndexes = append(fieldIndexes, i)
	}

	query := fmt.Sprintf("SELECT %s FROM %s", strings.Join(columns, ", "), tableName)

	rows, err := r.db.Pool.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to execute query: %w", err)
	}
	defer rows.Close()

	var results []T
	for rows.Next() {
		var item T
		val := reflect.ValueOf(&item).Elem()

		scanDest := make([]interface{}, len(columns))
		for idx, fieldIdx := range fieldIndexes {
			scanDest[idx] = val.Field(fieldIdx).Addr().Interface()
		}

		if err := rows.Scan(scanDest...); err != nil {
			return nil, fmt.Errorf("failed to scan row: %w", err)
		}
		results = append(results, item)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows error: %w", err)
	}

	return results, nil
}

// Create inserts a new record of type T into the database.
func (r *baseRepository[T]) Create(ctx context.Context, item *T) error {
	if r.db == nil {
		return fmt.Errorf("database connection is not initialized")
	}

	tableName := getTableName(*item)
	val := reflect.ValueOf(item).Elem()
	t := val.Type()

	var columns []string
	var values []interface{}
	var placeholders []string
	placeholderCount := 1

	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		if field.PkgPath != "" {
			continue
		}
		dbTag := field.Tag.Get("db")
		if dbTag == "" {
			dbTag = field.Tag.Get("json")
		}
		dbTag = strings.Split(dbTag, ",")[0]

		if dbTag == "-" {
			continue
		}
		if dbTag == "id" {
			continue
		}
		if dbTag == "" {
			dbTag = strings.ToLower(field.Name)
		}

		columns = append(columns, dbTag)
		values = append(values, val.Field(i).Interface())
		placeholders = append(placeholders, fmt.Sprintf("$%d", placeholderCount))
		placeholderCount++
	}

	query := fmt.Sprintf("INSERT INTO %s (%s) VALUES (%s) RETURNING id", tableName, strings.Join(columns, ", "), strings.Join(placeholders, ", "))

	var id int64
	err := r.db.Pool.QueryRow(ctx, query, values...).Scan(&id)
	if err != nil {
		return fmt.Errorf("failed to insert record: %w", err)
	}

	idField := val.FieldByName("ID")
	if idField.IsValid() && idField.CanSet() {
		idField.Set(reflect.ValueOf(int(id)).Convert(idField.Type()))
	}

	return nil
}

// FindByID retrieves a record by its integer ID.
func (r *baseRepository[T]) FindByID(ctx context.Context, id int64) (*T, error) {
	if r.db == nil {
		return nil, fmt.Errorf("database connection is not initialized")
	}

	var item T
	tableName := getTableName(item)
	val := reflect.ValueOf(&item).Elem()
	t := val.Type()

	var columns []string
	var fieldIndexes []int
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		if field.PkgPath != "" {
			continue
		}
		dbTag := field.Tag.Get("db")
		if dbTag == "" {
			dbTag = field.Tag.Get("json")
		}
		dbTag = strings.Split(dbTag, ",")[0]
		if dbTag == "-" {
			continue
		}
		if dbTag == "" {
			dbTag = strings.ToLower(field.Name)
		}
		columns = append(columns, dbTag)
		fieldIndexes = append(fieldIndexes, i)
	}

	query := fmt.Sprintf("SELECT %s FROM %s WHERE id = $1", strings.Join(columns, ", "), tableName)

	scanDest := make([]interface{}, len(columns))
	for idx, fieldIdx := range fieldIndexes {
		scanDest[idx] = val.Field(fieldIdx).Addr().Interface()
	}

	err := r.db.Pool.QueryRow(ctx, query, id).Scan(scanDest...)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, pgx.ErrNoRows
		}
		return nil, fmt.Errorf("failed to retrieve record by id: %w", err)
	}

	return &item, nil
}

// FindOneByField retrieves a record matching a specific field value.
func (r *baseRepository[T]) FindOneByField(ctx context.Context, fieldName string, fieldValue interface{}) (*T, error) {
	if r.db == nil {
		return nil, fmt.Errorf("database connection is not initialized")
	}

	var item T
	tableName := getTableName(item)
	val := reflect.ValueOf(&item).Elem()
	t := val.Type()

	var columns []string
	var fieldIndexes []int
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		if field.PkgPath != "" {
			continue
		}
		dbTag := field.Tag.Get("db")
		if dbTag == "" {
			dbTag = field.Tag.Get("json")
		}
		dbTag = strings.Split(dbTag, ",")[0]
		if dbTag == "-" {
			continue
		}
		if dbTag == "" {
			dbTag = strings.ToLower(field.Name)
		}
		columns = append(columns, dbTag)
		fieldIndexes = append(fieldIndexes, i)
	}

	query := fmt.Sprintf("SELECT %s FROM %s WHERE %s = $1", strings.Join(columns, ", "), tableName, fieldName)

	scanDest := make([]interface{}, len(columns))
	for idx, fieldIdx := range fieldIndexes {
		scanDest[idx] = val.Field(fieldIdx).Addr().Interface()
	}

	err := r.db.Pool.QueryRow(ctx, query, fieldValue).Scan(scanDest...)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, pgx.ErrNoRows
		}
		return nil, fmt.Errorf("failed to retrieve record by %s: %w", fieldName, err)
	}

	return &item, nil
}

// Update updates an existing record in the database.
func (r *baseRepository[T]) Update(ctx context.Context, item *T) error {
	if r.db == nil {
		return fmt.Errorf("database connection is not initialized")
	}

	tableName := getTableName(*item)
	val := reflect.ValueOf(item).Elem()
	t := val.Type()

	var updates []string
	var values []interface{}
	placeholderCount := 1
	var idVal interface{}

	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		if field.PkgPath != "" {
			continue
		}
		dbTag := field.Tag.Get("db")
		if dbTag == "" {
			dbTag = field.Tag.Get("json")
		}
		dbTag = strings.Split(dbTag, ",")[0]
		if dbTag == "-" {
			continue
		}
		if dbTag == "" {
			dbTag = strings.ToLower(field.Name)
		}

		if dbTag == "id" {
			idVal = val.Field(i).Interface()
			continue
		}

		updates = append(updates, fmt.Sprintf("%s = $%d", dbTag, placeholderCount))
		values = append(values, val.Field(i).Interface())
		placeholderCount++
	}

	if idVal == nil {
		return fmt.Errorf("missing id field for update")
	}

	query := fmt.Sprintf("UPDATE %s SET %s WHERE id = $%d", tableName, strings.Join(updates, ", "), placeholderCount)
	values = append(values, idVal)

	_, err := r.db.Pool.Exec(ctx, query, values...)
	if err != nil {
		return fmt.Errorf("failed to update record: %w", err)
	}

	return nil
}

// Delete deletes a record by its integer ID.
func (r *baseRepository[T]) Delete(ctx context.Context, id int64) error {
	if r.db == nil {
		return fmt.Errorf("database connection is not initialized")
	}

	var item T
	tableName := getTableName(item)
	query := fmt.Sprintf("DELETE FROM %s WHERE id = $1", tableName)

	_, err := r.db.Pool.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete record: %w", err)
	}

	return nil
}
