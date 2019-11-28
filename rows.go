package sqldblogger

import (
	"context"
	"database/sql/driver"
	"io"
	"reflect"
	"time"
)

// rows is a wrapper which implements:
// - driver.Rows
// - driver.RowsNextResultSet
// - driver.RowsColumnTypeScanType
// - driver.RowsColumnTypeDatabaseTypeName
// - driver.RowsColumnTypeLength
// - driver.RowsColumnTypeNullable
// - driver.RowsColumnTypePrecisionScale
type rows struct {
	driver.Rows
	logger *logger
}

// Columns implement driver.Rows
func (r *rows) Columns() []string {
	return r.Rows.Columns()
}

// Close implement driver.Rows
func (r *rows) Close() error {
	start := time.Now()
	err := r.Rows.Close()

	if err != nil {
		r.logger.log(context.Background(), LevelError, "RowsClose", start, err)
	}

	return err
}

// Next implement driver.Rows
func (r *rows) Next(dest []driver.Value) error {
	start := time.Now()
	err := r.Rows.Next(dest)

	if err != nil && err != io.EOF {
		r.logger.log(context.Background(), LevelError, "RowsNext", start, err)
	}

	return err
}

// HasNextResultSet implement driver.RowsNextResultSet
func (r *rows) HasNextResultSet() bool {
	if rs, ok := r.Rows.(driver.RowsNextResultSet); ok {
		return rs.HasNextResultSet()
	}

	return false
}

// NextResultSet implement driver.RowsNextResultSet
func (r *rows) NextResultSet() error {
	rs, ok := r.Rows.(driver.RowsNextResultSet)
	if !ok {
		return io.EOF
	}

	start := time.Now()
	err := rs.NextResultSet()

	if err != nil && err != io.EOF {
		r.logger.log(context.Background(), LevelError, "RowsNextResultSet", start, err)
	}

	return err
}

// ColumnTypeScanType implement driver.RowsColumnTypeScanType
func (r *rows) ColumnTypeScanType(index int) reflect.Type {
	if rs, ok := r.Rows.(driver.RowsColumnTypeScanType); ok {
		return rs.ColumnTypeScanType(index)
	}

	return reflect.SliceOf(reflect.TypeOf(""))
}

// ColumnTypeDatabaseTypeName driver.RowsColumnTypeDatabaseTypeName
func (r *rows) ColumnTypeDatabaseTypeName(index int) string {
	if rs, ok := r.Rows.(driver.RowsColumnTypeDatabaseTypeName); ok {
		return rs.ColumnTypeDatabaseTypeName(index)
	}

	return ""
}

// ColumnTypeLength implement driver.RowsColumnTypeLength
func (r *rows) ColumnTypeLength(index int) (length int64, ok bool) {
	if rs, ok := r.Rows.(driver.RowsColumnTypeLength); ok {
		return rs.ColumnTypeLength(index)
	}

	return 0, false
}

// ColumnTypeNullable implement driver.RowsColumnTypeNullable
func (r *rows) ColumnTypeNullable(index int) (nullable, ok bool) {
	if rs, ok := r.Rows.(driver.RowsColumnTypeNullable); ok {
		return rs.ColumnTypeNullable(index)
	}

	return false, false
}

// ColumnTypePrecisionScale implement driver.RowsColumnTypePrecisionScale
func (r *rows) ColumnTypePrecisionScale(index int) (precision, scale int64, ok bool) {
	if rs, ok := r.Rows.(driver.RowsColumnTypePrecisionScale); ok {
		return rs.ColumnTypePrecisionScale(index)
	}

	return 0, 0, false
}
