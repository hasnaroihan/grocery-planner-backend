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

func CreateRandomRecipe(t *testing.T) Recipe {
	user := CreateRandomUser(t);

	arg:= CreateRecipeParams{
		Name: util.RandomString(25),
		Author: user.ID,
		Portion: int32(util.RandomInt(0, 100)),
		Steps: sql.NullString{
			String: util.RandomString(400),
			Valid: true,
		},
	}

	recipe, err := testQueries.CreateRecipe(
		context.Background(),
		arg,
	)
	require.NoError(t, err)
	require.NotEmpty(t, recipe)

	require.NotZero(t, recipe.ID)
	require.Equal(t, user.ID, recipe.Author)
	require.Equal(t, arg.Name, recipe.Name)
	require.Equal(t, arg.Portion, recipe.Portion)
	require.Equal(t, arg.Steps, recipe.Steps)
	require.NotZero(t, recipe.CreatedAt)

	return recipe
}

func CreateRandomRecipeIngredient(t *testing.T) (Recipe, []GetRecipeIngredientsRow) {
	recipe := CreateRandomRecipe(t)
	ingredient := CreateRandomIngredient(t)
	unit := CreateRandomUnit(t)

	arg := CreateRecipeIngredientParams {
		IngredientID: ingredient.ID,
		RecipeID: recipe.ID,
		Amount: float32(util.RandomInt(0, 100)),
		UnitID: unit.ID,
	}

	recipeIngredient, err := testQueries.CreateRecipeIngredient(
		context.Background(),
		arg,
	)
	require.NoError(t, err)
	require.NotEmpty(t, recipeIngredient)

	require.Equal(t, arg.IngredientID, recipeIngredient.IngredientID)
	require.Equal(t, arg.RecipeID, recipeIngredient.RecipeID)
	require.Equal(t, arg.Amount, recipeIngredient.Amount)
	require.Equal(t, arg.UnitID, recipeIngredient.UnitID)

	recipeIngredientList, err := testQueries.GetRecipeIngredients(
		context.Background(),
		recipe.ID,
	)
	require.NoError(t, err)
	require.NotEmpty(t, recipeIngredientList)
	require.Len(t, recipeIngredientList, 1)

	return recipe, recipeIngredientList
}

func TestCreateRecipe(t *testing.T) {
	user := CreateRandomUser(t);

	arg:= CreateRecipeParams{
		Name: util.RandomString(25),
		Author: user.ID,
		Portion: int32(util.RandomInt(0, 100)),
		Steps: sql.NullString{
			String: util.RandomString(400),
			Valid: true,
		},
	}

	recipe, err := testQueries.CreateRecipe(
		context.Background(),
		arg,
	)
	require.NoError(t, err)
	require.NotEmpty(t, recipe)

	require.NotZero(t, recipe.ID)
	require.Equal(t, user.ID, recipe.Author)
	require.Equal(t, arg.Name, recipe.Name)
	require.Equal(t, arg.Portion, recipe.Portion)
	require.Equal(t, arg.Steps, recipe.Steps)
	require.NotZero(t, recipe.CreatedAt)
}

func TestCreateRecipeIngredient(t *testing.T) {
	recipe := CreateRandomRecipe(t)
	ingredient := CreateRandomIngredient(t)
	unit := CreateRandomUnit(t)

	arg := CreateRecipeIngredientParams {
		IngredientID: ingredient.ID,
		RecipeID: recipe.ID,
		Amount: float32(util.RandomInt(0, 100)),
		UnitID: unit.ID,
	}

	recipeIngredient, err := testQueries.CreateRecipeIngredient(
		context.Background(),
		arg,
	)
	require.NoError(t, err)
	require.NotEmpty(t, recipeIngredient)

	require.Equal(t, arg.IngredientID, recipeIngredient.IngredientID)
	require.Equal(t, arg.RecipeID, recipeIngredient.RecipeID)
	require.Equal(t, arg.Amount, recipeIngredient.Amount)
	require.Equal(t, arg.UnitID, recipeIngredient.UnitID)
}

func TestDeleteRecipe(t *testing.T) {
	recipeNew := CreateRandomRecipe(t)

	err := testQueries.DeleteRecipe(
		context.Background(),
		recipeNew.ID,
	)
	require.NoError(t, err)

	// Test read deleted user
	recipe, err := testQueries.GetRecipe(
		context.Background(),
		recipeNew.ID,
	)
	require.Error(t, err)
	require.EqualError(t, err, sql.ErrNoRows.Error())
	require.Empty(t, recipe)
}

func TestDeleteRecipeIngredient(t *testing.T) {
	_, recipeIngredient := CreateRandomRecipeIngredient(t)

	arg := DeleteRecipeIngredientParams {
		IngredientID: recipeIngredient[0].IngredientID,
		RecipeID: recipeIngredient[0].RecipeID,
	}

	err := testQueries.DeleteRecipeIngredient(
		context.Background(),
		arg,
	)
	require.NoError(t, err)
}

func TestGetRecipe(t *testing.T) {
	recipeNew := CreateRandomRecipe(t)

	recipe, err := testQueries.GetRecipe(
		context.Background(),
		recipeNew.ID,
	)
	require.NoError(t, err)
	require.NotEmpty(t, recipe)

	require.Equal(t, recipeNew.ID, recipe.ID)
	require.Equal(t, recipeNew.Name, recipe.Name)
	require.Equal(t, recipeNew.Author, recipe.Author)
	require.Equal(t, recipeNew.Portion, recipe.Portion)
	require.Equal(t, recipeNew.Steps, recipe.Steps)
	require.WithinDuration(t, recipeNew.CreatedAt, recipe.CreatedAt, time.Second)
}

func TestGetRecipeIngredients(t *testing.T) {
	recipeNew, recipeIngredientsNew := CreateRandomRecipeIngredient(t)

	recipeIngredients, err := testQueries.GetRecipeIngredients(
		context.Background(),
		recipeNew.ID,
	)
	require.NoError(t, err)
	require.NotEmpty(t, recipeIngredients)
	require.Len(t, recipeIngredients, 1)
	
	for _,row := range recipeIngredients {
		require.Equal(t, recipeIngredientsNew[0].IngredientID, row.IngredientID)
		require.Equal(t, recipeIngredientsNew[0].RecipeID, row.RecipeID)
		require.Equal(t, recipeIngredientsNew[0].Name, row.Name)
		require.Equal(t, recipeIngredientsNew[0].Amount, row.Amount)
		require.Equal(t, recipeIngredientsNew[0].UnitID, row.UnitID)
	}
}

func TestListRecipes(t *testing.T) {
	for i := 0; i < 4; i++ {
		CreateRandomRecipe(t)
	}

	arg := ListRecipesParams {
		Limit: 2,
		Offset: 2,
	}
	recipes, err := testQueries.ListRecipes(
		context.Background(),
		arg,
	)
	require.NoError(t, err)
	require.NotEmpty(t, recipes)
	require.Len(t, recipes, int(arg.Limit))

	for _,row := range recipes {
		require.NotEmpty(t, row)
	}
}

func TestSearchRecipe(t *testing.T) {
	user := CreateRandomUser(t)

	// First Data
	testQueries.CreateRecipe(
		context.Background(),
		CreateRecipeParams{
			Name: "bola daging",
			Author: user.ID,
			Portion: 1,
			Steps: sql.NullString{
				String: util.RandomString(400),
				Valid: true,
			},
		},
	)

	// Second Data
	testQueries.CreateRecipe(
		context.Background(),
		CreateRecipeParams{
			Name: "sup bola daging",
			Author: user.ID,
			Portion: 1,
			Steps: sql.NullString{
				String: util.RandomString(400),
				Valid: true,
			},
		},
	)

	// Third Data
	testQueries.CreateRecipe(
		context.Background(),
		CreateRecipeParams{
			Name: "bola daging bakar",
			Author: user.ID,
			Portion: 1,
			Steps: sql.NullString{
				String: util.RandomString(400),
				Valid: true,
			},
		},
	)

	arg := SearchRecipeParams {
		Name: "%bola daging%",
		Limit: 3,
		Offset: 0,
	}

	recipes, err := testQueries.SearchRecipe(
		context.Background(),
		arg,
	)
	require.NoError(t, err)
	require.NotEmpty(t, recipes)

	for _,row := range recipes {
		require.NotEmpty(t, row)
		require.Regexp(t, regexp.MustCompile("bola daging"), row.Name)
		require.NotZero(t, row.ID)
	}
}

func TestUpdateRecipe(t *testing.T) {
	recipeNew := CreateRandomRecipe(t)

	arg := UpdateRecipeParams {
		ID: recipeNew.ID,
		Name: "test update resep",
		Portion: 2,
		Steps: sql.NullString{},
	}

	recipe, err := testQueries.UpdateRecipe(
		context.Background(),
		arg,
	)
	require.NoError(t, err)
	require.NotEmpty(t, recipe)

	require.Equal(t, arg.ID, recipe.ID)
	require.Equal(t, arg.Name, recipe.Name)
	require.Equal(t, arg.Portion, recipe.Portion)
	require.Zero(t, recipe.Steps)
	require.WithinDuration(t, time.Now(), recipe.ModifiedAt, time.Second)
}

func TestUpdateRecipeIngredient(t *testing.T) {
	recipeNew, recipeIngredientNew := CreateRandomRecipeIngredient(t)

	arg := UpdateRecipeIngredientParams {
		RecipeID: recipeNew.ID,
		IngredientID: recipeIngredientNew[0].IngredientID,
		Amount: 3,
		UnitID: recipeIngredientNew[0].UnitID,
	}

	recipeIngredient, err := testQueries.UpdateRecipeIngredient(
		context.Background(),
		arg,
	)
	require.NoError(t, err)
	require.NotEmpty(t, recipeIngredient)

	require.Equal(t, arg.IngredientID, recipeIngredient.IngredientID)
	require.Equal(t, arg.RecipeID, recipeIngredient.RecipeID)
	require.Equal(t, arg.Amount, recipeIngredient.Amount)
	require.Equal(t, arg.UnitID, recipeIngredient.UnitID)
}