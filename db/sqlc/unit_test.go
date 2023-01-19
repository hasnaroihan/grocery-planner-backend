package db

import (
	"context"
	"database/sql"
	"testing"

	"github.com/hasnaroihan/grocery-planner/util"
	"github.com/stretchr/testify/require"
)

func CreateRandomUnit(t *testing.T) Unit{
	name := util.RandomUnit()
	unit, err := testQueries.CreateUnit(
		context.Background(),
		name,
	)
	require.NoError(t, err)
	require.NotEmpty(t, unit)

	require.NotZero(t, unit.ID)
	require.Equal(t, name, unit.Name)

	return unit
}

func TestCreateUnit(t *testing.T) {
	name := util.RandomUnit()
	unit, err := testQueries.CreateUnit(
		context.Background(),
		name,
	)
	require.NoError(t, err)
	require.NotEmpty(t, unit)

	require.NotZero(t, unit.ID)
	require.Equal(t, name, unit.Name)
}

func TestDeleteUnit(t *testing.T) {
	unitNew := CreateRandomUnit(t)
	err := testQueries.DeleteUnit(
		context.Background(),
		unitNew.ID,
	)
	require.NoError(t, err)

	// Test read deleted user
	unit, err := testQueries.GetUnit(
		context.Background(),
		unitNew.ID,
	)
	require.Error(t, err)
	require.EqualError(t, err, sql.ErrNoRows.Error())
	require.Empty(t, unit)
}

func TestGetUnit(t *testing.T) {
	unitNew := CreateRandomUnit(t)

	unit, err := testQueries.GetUnit(
		context.Background(),
		unitNew.ID,
	)
	require.NoError(t, err)
	require.NotEmpty(t, unit)

	require.Equal(t, unitNew.ID, unit.ID)
	require.Equal(t, unitNew.Name, unit.Name)
}

func TestListUnits(t *testing.T) {
	for i := 0; i < 5; i++ {
		CreateRandomUnit(t)
	}
	units, err := testQueries.ListUnits(
		context.Background(),
	)
	require.NoError(t, err)
	require.NotEmpty(t, units)

	for _,row := range units {
		require.NotEmpty(t, row)
	}
}

func TestUpdateUnit(t *testing.T) {
	unitNew := CreateRandomUnit(t)

	arg := UpdateUnitParams {
		ID: unitNew.ID,
		Name: util.RandomUnit(),
	}

	unit, err := testQueries.UpdateUnit(
		context.Background(),
		arg,
	)
	require.NoError(t, err)
	require.NotEmpty(t, unit)

	require.Equal(t, arg.ID, unit.ID)
	require.Equal(t, arg.Name, unit.Name)
}