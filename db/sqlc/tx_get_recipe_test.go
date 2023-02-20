package db

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestGetRecipeTx(t *testing.T) {
	storage := NewStorage(testDB)
	recipeNew, recipeIngredientsNew := CreateRandomRecipeIngredient(t)

	errs := make(chan error)
	results := make(chan RecipeResult)

	go func() {
		result, err := storage.GetRecipeTx(
			context.Background(), recipeNew.ID,
		)
		errs <- err
		results <- result
	}()

	err := <-errs
	result := <-results
	require.NoError(t, err)
	require.NotEmpty(t, result)
	require.Len(t, recipeIngredientsNew, 1)

	require.Equal(t, recipeNew.ID, result.Recipe.ID)
	require.Equal(t, recipeNew.Name, result.Recipe.Name)
	require.Equal(t, recipeNew.Author, result.Recipe.Author)
	require.Equal(t, recipeNew.Portion, result.Recipe.Portion)
	require.Equal(t, recipeNew.Steps, result.Recipe.Steps)
	require.WithinDuration(t, recipeNew.CreatedAt, result.Recipe.CreatedAt, time.Second)

	for _, row := range recipeIngredientsNew {
		require.Equal(t, recipeIngredientsNew[0].IngredientID, row.IngredientID)
		require.Equal(t, recipeIngredientsNew[0].RecipeID, row.RecipeID)
		require.Equal(t, recipeIngredientsNew[0].Name, row.Name)
		require.Equal(t, recipeIngredientsNew[0].Amount, row.Amount)
		require.Equal(t, recipeIngredientsNew[0].UnitID, row.UnitID)
	}
}