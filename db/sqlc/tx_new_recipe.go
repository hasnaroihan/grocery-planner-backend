package db

import (
	"context"
	"database/sql"

	"github.com/google/uuid"
)

type ListIngredientParam struct {
	ID     sql.NullInt32 `json:"id"`
	Name   string        `json:"name"`
	Amount float32       `json:"amount"`
	UnitID int32         `json:"unitID"`
}
type NewRecipeParams struct {
	Name            string                `json:"name"`
	Author          uuid.UUID             `json:"author"`
	Portion         int32                 `json:"portion"`
	Steps           sql.NullString        `json:"steps"`
	ListIngredients []ListIngredientParam `json:"ingredients"`
}

type IngredientResult struct {
	Ingredient Ingredient `json:"ingredient"`
	Amount     float32    `json:"amount"`
	UnitID     int32      `json:"unitID"`
}

type RecipeResult struct {
	Recipe      Recipe                    `json:"recipe"`
	Ingredients []GetRecipeIngredientsRow `json:"ingredients"`
}

// Create recipe, create new ingredients, create recipe-ingredients
func (s *SQLStorage) NewRecipeTx(ctx context.Context, arg NewRecipeParams) (RecipeResult, error) {
	var result RecipeResult

	err := s.execTx(ctx, func(q *Queries) error {
		var err error

		result.Recipe, err = q.CreateRecipe(
			ctx,
			CreateRecipeParams{
				Name:    arg.Name,
				Author:  arg.Author,
				Portion: arg.Portion,
				Steps:   arg.Steps,
			},
		)
		if err != nil {
			return err
		}

		// Create ingredients for every new ingredients, and create recipe ingredients
		for _, item := range arg.ListIngredients {
			var create CreateRecipeIngredientParams

			if item.ID.Valid { // Get ingredient
				create.RecipeID = result.Recipe.ID
				create.IngredientID = item.ID.Int32
				create.Amount = item.Amount
				create.UnitID = item.UnitID
			} else { // Create ingredient
				ingredient, err := q.CreateIngredient(
					ctx,
					CreateIngredientParams{
						Name: item.Name,
						DefaultUnit: sql.NullInt32{
							Int32: item.UnitID,
							Valid: true,
						},
					},
				)
				if err != nil && err != sql.ErrNoRows {
					return err
				}
				if err == sql.ErrNoRows {
					ingredient, err = q.SearchIngredientName(ctx, item.Name)
					if err != nil {
						return err
					}
				}

				create.RecipeID = result.Recipe.ID
				create.IngredientID = ingredient.ID
				create.Amount = item.Amount
				create.UnitID = item.UnitID
			}

			// Create recipes ingredient
			_, err = q.CreateRecipeIngredient(ctx, create)
			if err != nil {
				return err
			}			
		}

		result.Ingredients, err = q.GetRecipeIngredients(ctx, result.Recipe.ID)
		if err != nil {
			return err
		}

		return nil
	})

	return result, err
}