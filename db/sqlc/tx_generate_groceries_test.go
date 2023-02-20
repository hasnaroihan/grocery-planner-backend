package db

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/hasnaroihan/grocery-planner/util"
	"github.com/stretchr/testify/require"
	"golang.org/x/exp/slices"
)

func TestGenerateGroceries(t *testing.T) {
	var recipes []Recipe
	var recipeIngredients []GetRecipeIngredientsRow

	storage := NewStorage(testDB)

	author := CreateRandomUser(t)
	arg := GenerateGroceriesParam{
		Author: uuid.NullUUID{
			UUID:  author.ID,
			Valid: true,
		},
	}
	for i := 0; i < 5; i++ {
		recipe, recipeIngredient := CreateRandomRecipeIngredient(t)
		recipes = append(recipes, recipe)
		recipeIngredients = append(recipeIngredients, recipeIngredient...)

		arg.Recipes = append(arg.Recipes, scheduleRecipePortion{
			RecipeID: int64(recipe.ID),
			Portion:  int32(util.RandomInt(1, 5)),
		})
	}

	errs := make(chan error)
	results := make(chan GenerateGroceriesResult)

	go func() {
		result, err := storage.GenerateGroceries(
			context.Background(),
			arg,
		)

		errs <- err
		results <- result
	}()

	err := <-errs
	result := <-results
	require.NoError(t, err)
	require.NotEmpty(t, result)

	require.Equal(t, author.ID, result.Schedule.Author.UUID)
	require.NotZero(t, result.Schedule.ID)
	require.NotZero(t, result.Schedule.CreatedAt)

	for _, row := range result.Recipes {
		idx := slices.IndexFunc(recipes, func(r Recipe) bool { return r.ID == row.RecipeID })
		idx2 := slices.IndexFunc(arg.Recipes, func(r scheduleRecipePortion) bool { return r.RecipeID == row.RecipeID })
		require.NotNil(t, idx)
		require.Equal(t, recipes[idx].Name, row.Name)
		require.Equal(t, arg.Recipes[idx2].Portion, row.Portion)
	}

	for _, row := range result.Groceries {
		idx := slices.IndexFunc(recipeIngredients, func(i GetRecipeIngredientsRow) bool { return int32(i.RecipeID) == row.ID })
		require.NotNil(t, idx)
	}
}