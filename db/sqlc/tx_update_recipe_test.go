package db

import (
	"context"
	"database/sql"
	"testing"

	"github.com/hasnaroihan/grocery-planner/util"
	"github.com/stretchr/testify/require"
)

func TestUpdateRecipeTx(t *testing.T) {
	storage := NewStorage(testDB)
	var args []TxUpdateRecipeParams

	for i := 0; i < 3; i++ {
		author := CreateRandomUser(t)
		recipeNew := NewRecipeParams{
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
			recipeNew.ListIngredients = append(recipeNew.ListIngredients,
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
		
		recipe,_ := storage.NewRecipeTx(context.Background(), recipeNew)
		args = append(args, TxUpdateRecipeParams{
			Recipe: UpdateRecipeParams{
				ID: recipe.Recipe.ID,
				Name: util.RandomString(15),
				Portion: int32(util.RandomInt(2,10)),
				Steps: sql.NullString{
					String: util.RandomString(50),
					Valid: true,
				},
			},
			ListIngredients: []ListIngredientParam{
				{
					ID: sql.NullInt32{
						Int32: recipe.Ingredients[0].IngredientID,
						Valid: true,
					},
					Name: recipe.Ingredients[0].Name,
					Amount: float32(util.RandomInt(300,500)),
					UnitID: recipe.Ingredients[0].UnitID,
				},
				{
					ID:     sql.NullInt32{},
					Name:   util.RandomIngredient(),
					Amount: float32(util.RandomInt(1, 3)),
					UnitID: CreateRandomUnit(t).ID,
				},
			},
		})
	}

	errs := make(chan error)
	results := make(chan RecipeResult)

	for idx := range args {
		go func(arg *TxUpdateRecipeParams, i int) {
			result, err := storage.UpdateRecipeTx(context.Background(), *arg)
			// log.Print(i)
			errs <- err
			results <- result
		}(&args[idx], idx)
	}

	for i := 0; i < 3; i++ {
		err := <-errs
		result := <-results
		require.NoError(t, err)
		require.NotEmpty(t, result)
		require.NotZero(t, result.Recipe.ID)
		require.NotZero(t, result.Recipe.Name)
		require.Len(t, result.Ingredients, 3)
	}
}