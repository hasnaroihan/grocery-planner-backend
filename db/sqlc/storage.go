package db

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/google/uuid"
)

type Storage struct {
	*Queries
	db *sql.DB
}

func NewStorage(db *sql.DB) *Storage {
	return &Storage{
		db: db,
		Queries: New(db),
	}
}

func (s *Storage) execTx(ctx context.Context, fn func(*Queries) error) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}

	q := New(tx)
	err = fn(q)
	if err != nil {
		if rbErr := tx.Rollback(); rbErr != nil {
			return fmt.Errorf("transaction error: %v, rollback error: %v", err, rbErr)
		}

		return err
	}

	return tx.Commit()
}

type ListIngredientParam struct {
	ID sql.NullInt32 `json:"id"`
	Name string		`json:"name"`
	Amount float32 	`json:"amount"`
	UnitID int32   	`json:"unitID"`
}
type NewRecipeParams struct {
	Name    string         					`json:"name"`
	Author  uuid.UUID      					`json:"author"`
	Portion int32          					`json:"portion"`
	Steps   sql.NullString 					`json:"steps"`
	ListIngredients []ListIngredientParam 	`json:"ingredients"`
}

type IngredientResult struct {
	Ingredient Ingredient `json:"ingredient"`
	Amount float32 	`json:"amount"`
	UnitID int32   	`json:"unitID"`
}

type NewRecipeResult struct {
	Recipe Recipe 					`json:"recipe"`
	Ingredients []IngredientResult 	`json:"ingredients"`
}

// Create recipe, create new ingredients, create recipe-ingredients
func (s *Storage) NewRecipeTx(ctx context.Context, arg NewRecipeParams) (NewRecipeResult, error) {
	var result NewRecipeResult

	err := s.execTx(ctx, func(q *Queries) error {
		var err error

		result.Recipe, err = q.CreateRecipe(
			ctx,
			CreateRecipeParams{
				Name: arg.Name,
				Author: arg.Author,
				Portion: arg.Portion,
				Steps: arg.Steps,
			},
		)
		if err != nil {
			return err
		}

		// Create ingredients for every new ingredients, and create recipe ingredients
		for _,item := range arg.ListIngredients {
			var ingredient Ingredient
			var recipesIngredient RecipesIngredient

			if item.ID.Valid { // Get ingredient
				ingredient, err = q.GetIngredient(ctx, item.ID.Int32)
				if err != nil {
					return err
				}
			} else { // Create ingredient
				ingredient, err = q.CreateIngredient(
					ctx,
					CreateIngredientParams{
						Name: item.Name,
						DefaultUnit: sql.NullInt32{
							Int32: item.UnitID,
							Valid: true,
						},
					},
				)
				if err != nil {
					return err
				}
			}

			// Create recipes ingredient
			recipesIngredient, err = q.CreateRecipeIngredient(
				ctx,
				CreateRecipeIngredientParams{
					RecipeID: result.Recipe.ID,
					IngredientID: ingredient.ID,
					Amount: item.Amount,
					UnitID: item.UnitID,
				},
			)
			if err != nil {
				return err
			}

			// Append ingredients to result
			result.Ingredients = append(result.Ingredients, IngredientResult{
				ingredient,
				recipesIngredient.Amount,
				recipesIngredient.UnitID,
			})
		}

		return nil
	})
	
	return result, err
}

type scheduleRecipePortion struct {
	RecipeID int32	`json:"recipe_id"`
	Portion int32 `json:"portion"`
}

type GenerateGroceriesParam struct {
	Author uuid.NullUUID				`json:"author"`
	Recipes []scheduleRecipePortion		`json:"recipes"`
}

type GenerateGroceriesResult struct {
	Schedule Schedule 					`json:"schedule"`
	Recipes []SchedulesRecipe		`json:"recipes"`
	Groceries []ListGroceriesRow		`json:"groceries"`
}

// Create schedule, create schedule recipes, return list of groceries
func (s *Storage) GenerateGroceries (ctx context.Context, arg GenerateGroceriesParam) (GenerateGroceriesResult, error) {
	var result GenerateGroceriesResult

	err := s.execTx(ctx, func(q *Queries) error {
		var err error

		result.Schedule, err = q.CreateSchedule(ctx, arg.Author)
		if err != nil {
			return err
		}

		for _,recipe := range arg.Recipes {
			var scheduleRecipe SchedulesRecipe

			scheduleRecipe, err = q.CreateScheduleRecipe(
				ctx,
				CreateScheduleRecipeParams{
					ScheduleID: result.Schedule.ID,
					RecipeID: int64(recipe.RecipeID),
					Portion: recipe.Portion,
				},
			)
			if err != nil {
				return err
			}

			result.Recipes = append(result.Recipes, scheduleRecipe)
		}

		result.Groceries, err = q.ListGroceries(ctx, result.Schedule.ID)
		if err != nil {
			return err
		}

		return nil
	})

	return result, err
}