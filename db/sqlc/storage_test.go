package db

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/hasnaroihan/grocery-planner/util"
	"github.com/stretchr/testify/require"
	"golang.org/x/exp/slices"
)

func TestNewRecipeTx(t *testing.T) {
	storage := NewStorage(testDB)
	var args []NewRecipeParams

	for i := 0; i < 3; i++ {
		var arg NewRecipeParams
		author := CreateRandomUser(t)
		arg = NewRecipeParams{
			Name: util.RandomString(10),
			Author: author.ID,
			Portion: int32(util.RandomInt(1, 4)),
			Steps: sql.NullString{
				String: util.RandomString(100),
				Valid: true,
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
				Name: ingredient.Name,
				Amount: float32(util.RandomInt(50,175)),
				UnitID: ingredient.DefaultUnit.Int32,
			},
			)
		}
	
		// New random ingredient
		arg.ListIngredients = append(arg.ListIngredients, 
		ListIngredientParam{
			ID: sql.NullInt32{},
			Name: util.RandomIngredient(),
			Amount: float32(util.RandomInt(1, 3)),
			UnitID: CreateRandomUnit(t).ID,
		},
		)
	
		// New dedicated ingredient for concurrency test
		arg.ListIngredients = append(arg.ListIngredients, 
		ListIngredientParam{
			ID: sql.NullInt32{},
			Name: "garam",
			Amount: float32(util.RandomInt(1, 3)),
			UnitID: CreateRandomUnit(t).ID,
		})

		args = append(args, arg)
	}
	

	errs := make(chan error)
	results := make(chan NewRecipeResult)

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

func TestGetRecipeTx(t *testing.T) {
	storage := NewStorage(testDB)
	recipeNew, recipeIngredientsNew := CreateRandomRecipeIngredient(t)

	errs := make(chan error)
	results := make(chan NewRecipeResult)

	go func() {
		result, err := storage.GetRecipeTx(
			context.Background(), recipeNew.ID,
		)
		errs <- err
		results <- result
	}()

	err := <- errs
	result := <- results
	require.NoError(t, err)
	require.NotEmpty(t, result)
	require.Len(t, recipeIngredientsNew, 1)

	require.Equal(t, recipeNew.ID, result.Recipe.ID)
	require.Equal(t, recipeNew.Name, result.Recipe.Name)
	require.Equal(t, recipeNew.Author, result.Recipe.Author)
	require.Equal(t, recipeNew.Portion, result.Recipe.Portion)
	require.Equal(t, recipeNew.Steps, result.Recipe.Steps)
	require.WithinDuration(t, recipeNew.CreatedAt, result.Recipe.CreatedAt, time.Second)

	for _,row := range recipeIngredientsNew {
		require.Equal(t, recipeIngredientsNew[0].IngredientID, row.IngredientID)
		require.Equal(t, recipeIngredientsNew[0].RecipeID, row.RecipeID)
		require.Equal(t, recipeIngredientsNew[0].Name, row.Name)
		require.Equal(t, recipeIngredientsNew[0].Amount, row.Amount)
		require.Equal(t, recipeIngredientsNew[0].UnitID, row.UnitID)
	}
}

func TestGenerateGroceries(t *testing.T) {
	var recipes []Recipe
	var recipeIngredients []GetRecipeIngredientsRow

	storage := NewStorage(testDB)

	author := CreateRandomUser(t)
	arg := GenerateGroceriesParam{
		Author: uuid.NullUUID{
			UUID: author.ID,
			Valid: true,
		},
	}
	for i := 0; i < 5; i++ {
		recipe, recipeIngredient := CreateRandomRecipeIngredient(t)
		recipes = append(recipes, recipe)
		recipeIngredients = append(recipeIngredients, recipeIngredient...)

		arg.Recipes = append(arg.Recipes, scheduleRecipePortion{
			RecipeID: int64(recipe.ID),
			Portion: int32(util.RandomInt(1,5)),
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

	err := <- errs
	result := <- results
	require.NoError(t, err)
	require.NotEmpty(t, result)

	require.Equal(t, author.ID, result.Schedule.Author.UUID)
	require.NotZero(t, result.Schedule.ID)
	require.NotZero(t, result.Schedule.CreatedAt)

	for _,row := range result.Recipes {
		idx := slices.IndexFunc(recipes, func (r Recipe) bool {return r.ID == row.RecipeID})
		idx2 := slices.IndexFunc(arg.Recipes, func (r scheduleRecipePortion) bool {return r.RecipeID == row.RecipeID})
		require.NotNil(t, idx)
		require.Equal(t, recipes[idx].Name, row.Name)
		require.Equal(t, arg.Recipes[idx2].Portion, row.Portion)
	}

	for _,row := range result.Groceries {
		idx := slices.IndexFunc(recipeIngredients, func (i GetRecipeIngredientsRow) bool {return int32(i.RecipeID) == row.ID})
		require.NotNil(t, idx)
	}
}