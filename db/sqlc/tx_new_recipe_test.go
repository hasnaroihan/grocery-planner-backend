package db

import (
	"context"
	"database/sql"
	"testing"

	"github.com/hasnaroihan/grocery-planner/util"
	"github.com/stretchr/testify/require"
)

func TestNewRecipeTx(t *testing.T) {
	storage := NewStorage(testDB)
	var args []NewRecipeParams

	for i := 0; i < 3; i++ {
		var arg NewRecipeParams
		author := CreateRandomUser(t)
		arg = NewRecipeParams{
			Name:    util.RandomString(10),
			Author:  author.ID,
			Portion: int32(util.RandomInt(1, 4)),
			Steps: sql.NullString{
				String: util.RandomString(100),
				Valid:  true,
			},
		}
		for i := 0; i < 2; i++ {
			ingredient := CreateRandomIngredient(t)
			arg.ListIngredients = append(arg.ListIngredients,
				ListIngredientParam{
					ID: sql.NullInt32{
						Int32: ingredient.ID,
						Valid: true,
					},
					Name:   ingredient.Name,
					Amount: float32(util.RandomInt(50, 175)),
					UnitID: ingredient.DefaultUnit.Int32,
				},
			)
		}

		// New random ingredient
		arg.ListIngredients = append(arg.ListIngredients,
			ListIngredientParam{
				ID:     sql.NullInt32{},
				Name:   util.RandomIngredient(),
				Amount: float32(util.RandomInt(1, 3)),
				UnitID: CreateRandomUnit(t).ID,
			},
		)

		// New dedicated ingredient for concurrency test
		arg.ListIngredients = append(arg.ListIngredients,
			ListIngredientParam{
				ID:     sql.NullInt32{},
				Name:   "garam",
				Amount: float32(util.RandomInt(1, 3)),
				UnitID: CreateRandomUnit(t).ID,
			})

		args = append(args, arg)
	}

	errs := make(chan error)
	results := make(chan RecipeResult)

	for idx := range args {
		go func(arg *NewRecipeParams) {
			result, err := storage.NewRecipeTx(context.Background(), *arg)

			errs <- err
			results <- result
		}(&args[idx])
	}

	for i := 0; i < 3; i++ {
		err := <-errs
		result := <-results
		require.NoError(t, err)
		require.NotEmpty(t, result)
		require.NotZero(t, result.Recipe.ID)
		require.NotZero(t, result.Recipe.Name)
		require.Len(t, result.Ingredients, 4)
	}
}