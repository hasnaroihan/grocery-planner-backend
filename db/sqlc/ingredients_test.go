package db

import (
	"context"
	"database/sql"
	"regexp"
	"testing"
	"time"

	"github.com/hasnaroihan/grocery-planner/util"
	"github.com/stretchr/testify/require"
)

func CreateRandomIngredient(t *testing.T) Ingredient {
	unit := CreateRandomUnit(t)
	arg := CreateIngredientParams {
		Name: util.RandomIngredient(),
		DefaultUnit: sql.NullInt32{
			Int32: unit.ID,
			Valid: true,
		},
	}

	ingredient, err := testQueries.CreateIngredient(
		context.Background(),
		arg,
	)
	require.NoError(t, err)
	require.NotEmpty(t, ingredient)

	require.NotZero(t, ingredient.ID)
	require.Equal(t, arg.Name, ingredient.Name)
	require.NotZero(t, ingredient.CreatedAt)
	require.Equal(t, arg.DefaultUnit.Int32, ingredient.DefaultUnit.Int32)

	return ingredient
}

func TestCreateIngredient(t *testing.T){
	unit := CreateRandomUnit(t)
	arg := CreateIngredientParams {
		Name: util.RandomIngredient(),
		DefaultUnit: sql.NullInt32{
			Int32: unit.ID,
			Valid: true,
		},
	}

	ingredient, err := testQueries.CreateIngredient(
		context.Background(),
		arg,
	)
	require.NoError(t, err)
	require.NotEmpty(t, ingredient)

	require.NotZero(t, ingredient.ID)
	require.Equal(t, arg.Name, ingredient.Name)
	require.NotZero(t, ingredient.CreatedAt)
	require.Equal(t, arg.DefaultUnit.Int32, ingredient.DefaultUnit.Int32)
}

func TestDeleteIngredient(t *testing.T) {
	ingredientNew := CreateRandomIngredient(t)
	err := testQueries.DeleteIngredient(
		context.Background(),
		ingredientNew.ID,
	)
	require.NoError(t, err)

	// Test read deleted user
	ingredient, err := testQueries.GetIngredient(
		context.Background(),
		ingredientNew.ID,
	)
	require.Error(t, err)
	require.EqualError(t, err, sql.ErrNoRows.Error())
	require.Empty(t, ingredient)
}

func TestGetIngredient(t *testing.T) {
	ingredientNew := CreateRandomIngredient(t)

	ingredient, err := testQueries.GetIngredient(
		context.Background(),
		ingredientNew.ID,
	)
	require.NoError(t, err)
	require.NotEmpty(t, ingredient)

	require.Equal(t, ingredientNew.ID, ingredient.ID)
	require.Equal(t, ingredientNew.Name, ingredient.Name)
	require.WithinDuration(t, ingredientNew.CreatedAt, ingredient.CreatedAt, time.Second)
	require.Equal(t, ingredientNew.DefaultUnit, ingredient.DefaultUnit)
}

func TestListIngredients(t *testing.T) {
	for i := 0; i < 5; i++ {
		CreateRandomIngredient(t)
	}
	ingredients, err := testQueries.ListIngredients(
		context.Background(),
	)
	require.NoError(t, err)
	require.NotEmpty(t, ingredients)

	for _,row := range ingredients {
		require.NotEmpty(t, row)
	}
}

func TestSearchIngredients(t *testing.T) {
	unit := CreateRandomUnit(t)

	// First Data
	testQueries.CreateIngredient(
		context.Background(),
		CreateIngredientParams{
			Name: "daging",
			DefaultUnit: sql.NullInt32{
				Int32: unit.ID,
				Valid: true,
			},
		},
	)

	// Second Data
	testQueries.CreateIngredient(
		context.Background(),
		CreateIngredientParams{
			Name: "daging giling",
			DefaultUnit: sql.NullInt32{
				Int32: unit.ID,
				Valid: true,
			},
		},
	)

	// Third Data
	testQueries.CreateIngredient(
		context.Background(),
		CreateIngredientParams{
			Name: "mix daging",
			DefaultUnit: sql.NullInt32{
				Int32: unit.ID,
				Valid: true,
			},
		},
	)

	ingredients, err := testQueries.SearchIngredients(
		context.Background(),
		"%daging%",
	)
	require.NoError(t, err)
	require.NotEmpty(t, ingredients)

	for _,row := range ingredients {
		require.NotEmpty(t, row)
		require.Regexp(t, regexp.MustCompile("daging"), row)
	}
}

func TestUpdateIngredient(t *testing.T) {
	ingredientNew := CreateRandomIngredient(t)

	arg := UpdateIngredientParams {
		ID: ingredientNew.ID,
		Name: util.RandomIngredient(),
		DefaultUnit: ingredientNew.DefaultUnit,
	}

	ingredient, err := testQueries.UpdateIngredient(
		context.Background(),
		arg,
	)
	require.NoError(t, err)
	require.NotEmpty(t, ingredient)

	require.Equal(t, arg.ID, ingredient.ID)
	require.Equal(t, arg.Name, ingredient.Name)
	require.WithinDuration(t, ingredientNew.CreatedAt, ingredient.CreatedAt, time.Second)
	require.Equal(t, arg.DefaultUnit, ingredient.DefaultUnit)
}