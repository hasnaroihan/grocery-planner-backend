ALTER TABLE IF EXISTS public.ingredients
    DROP CONSTRAINT IF EXISTS unique_id_ig;

ALTER TABLE IF EXISTS public.ingredients
    DROP CONSTRAINT IF EXISTS unique_name_ig;

ALTER TABLE IF EXISTS public.units
    DROP CONSTRAINT IF EXISTS unique_id_units;

ALTER TABLE IF EXISTS public.units
    DROP CONSTRAINT IF EXISTS unique_name_units;

ALTER TABLE IF EXISTS public.users
    DROP CONSTRAINT IF EXISTS unique_id_users;

ALTER TABLE IF EXISTS public.users
    DROP CONSTRAINT IF EXISTS unique_name_users;

ALTER TABLE IF EXISTS public.users
    DROP CONSTRAINT IF EXISTS unique_email_users;

DROP TABLE IF EXISTS public.recipes_ingredients;
DROP TABLE IF EXISTS public.schedules_recipes;
DROP TABLE IF EXISTS public.recipes_steps;
DROP TABLE IF EXISTS public.schedules;
DROP TABLE IF EXISTS public.ingredients;
DROP TABLE IF EXISTS public.recipes;
DROP TABLE IF EXISTS public.unit;
DROP TABLE IF EXISTS public.users;
