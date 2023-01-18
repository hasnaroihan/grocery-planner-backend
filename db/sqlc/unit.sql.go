// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.16.0
// source: unit.sql

package db

import (
	"context"
)

const createUnit = `-- name: CreateUnit :one
INSERT INTO units (
    name
) VALUES (
    $1
) RETURNING id, name
`

func (q *Queries) CreateUnit(ctx context.Context, name string) (Unit, error) {
	row := q.db.QueryRowContext(ctx, createUnit, name)
	var i Unit
	err := row.Scan(&i.ID, &i.Name)
	return i, err
}

const deleteUnit = `-- name: DeleteUnit :exec
DELETE FROM units
WHERE id = $1
`

func (q *Queries) DeleteUnit(ctx context.Context, id int32) error {
	_, err := q.db.ExecContext(ctx, deleteUnit, id)
	return err
}

const listUnit = `-- name: ListUnit :many
SELECT id, name from units
`

func (q *Queries) ListUnit(ctx context.Context) ([]Unit, error) {
	rows, err := q.db.QueryContext(ctx, listUnit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []Unit
	for rows.Next() {
		var i Unit
		if err := rows.Scan(&i.ID, &i.Name); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Close(); err != nil {
		return nil, err
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const updateUnit = `-- name: UpdateUnit :exec
UPDATE units
    set name = $2
WHERE id = $1
`

type UpdateUnitParams struct {
	ID   int32  `json:"id"`
	Name string `json:"name"`
}

func (q *Queries) UpdateUnit(ctx context.Context, arg UpdateUnitParams) error {
	_, err := q.db.ExecContext(ctx, updateUnit, arg.ID, arg.Name)
	return err
}