CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

CREATE TABLE IF NOT EXISTS public.ingredients
(
    id integer NOT NULL GENERATED ALWAYS AS IDENTITY ( INCREMENT 1 START 1 MINVALUE 1 MAXVALUE 9999 ),
    name character varying(100) NOT NULL,
    created_at timestamp with time zone NOT NULL DEFAULT (now()),
    default_unit integer,
    PRIMARY KEY (id)
);

CREATE TABLE IF NOT EXISTS public.unit
(
    id integer NOT NULL GENERATED ALWAYS AS IDENTITY ( INCREMENT 1 START 1 MINVALUE 1 ),
    name character varying(25) NOT NULL,
    PRIMARY KEY (id)
);

CREATE TABLE IF NOT EXISTS public.recipes
(
    id bigint NOT NULL GENERATED ALWAYS AS IDENTITY ( INCREMENT 1 START 1 ),
    name character varying(255) NOT NULL DEFAULT 'unknown',
    author uuid NOT NULL,
	portion integer NOT NULL DEFAULT 1,
    created_at timestamp with time zone NOT NULL DEFAULT (now()),
    modified_at date NOT NULL,
    PRIMARY KEY (id)
);

CREATE TABLE IF NOT EXISTS public.users
(
    id uuid NOT NULL DEFAULT uuid_generate_v4(),
    username character varying(255) NOT NULL,
    email character varying NOT NULL,
    password character varying NOT NULL,
    created_at timestamp with time zone NOT NULL DEFAULT (now()),
    role character varying(25) NOT NULL DEFAULT 'common',
    verified_at timestamp with time zone DEFAULT null,
    PRIMARY KEY (id),
    UNIQUE (id, username)
);

CREATE TABLE IF NOT EXISTS public.recipes_ingredients
(
    ingredients_id integer NOT NULL,
    recipes_id bigint NOT NULL,
    amount real NOT NULL,
    unit_id integer NOT NULL,
    PRIMARY KEY (ingredients_id, recipes_id)
);

CREATE TABLE IF NOT EXISTS public.schedules
(
    id bigint NOT NULL GENERATED ALWAYS AS IDENTITY ( INCREMENT 1 START 1 MINVALUE 1 ),
    created_at timestamp with time zone NOT NULL DEFAULT (now()),
    PRIMARY KEY (id)
);

CREATE TABLE IF NOT EXISTS public.schedules_recipes
(
    schedule_id bigint NOT NULL,
    recipe_id bigint NOT NULL,
	portion integer NOT NULL DEFAULT 1,
    PRIMARY KEY (schedule_id, recipe_id)
);

CREATE TABLE IF NOT EXISTS public.recipes_steps
(
    recipe_id bigint NOT NULL,
    "number" integer NOT NULL,
    steps text NOT NULL,
    PRIMARY KEY (recipe_id, "number")
);

ALTER TABLE IF EXISTS public.ingredients
    ADD CONSTRAINT default_unit FOREIGN KEY (default_unit)
    REFERENCES public.unit (id) MATCH SIMPLE
    ON UPDATE CASCADE
    ON DELETE SET NULL
    NOT VALID;


ALTER TABLE IF EXISTS public.recipes
    ADD CONSTRAINT fk_recipes FOREIGN KEY (author)
    REFERENCES public.users (id) MATCH SIMPLE
    ON UPDATE RESTRICT
    ON DELETE CASCADE;


ALTER TABLE IF EXISTS public.recipes_ingredients
    ADD CONSTRAINT fk_recipe_ingredients FOREIGN KEY (ingredients_id)
    REFERENCES public.ingredients (id) MATCH SIMPLE
    ON UPDATE CASCADE
    ON DELETE CASCADE
    NOT VALID;


ALTER TABLE IF EXISTS public.recipes_ingredients
    ADD CONSTRAINT fk_recipe_recipe FOREIGN KEY (recipes_id)
    REFERENCES public.recipes (id) MATCH SIMPLE
    ON UPDATE RESTRICT
    ON DELETE CASCADE
    NOT VALID;


ALTER TABLE IF EXISTS public.recipes_ingredients
    ADD CONSTRAINT fk_recipe_unit FOREIGN KEY (unit_id)
    REFERENCES public.unit (id) MATCH SIMPLE
    ON UPDATE CASCADE
    ON DELETE SET NULL
    NOT VALID;


ALTER TABLE IF EXISTS public.schedules_recipes
    ADD CONSTRAINT fk_schedule_schedule FOREIGN KEY (schedule_id)
    REFERENCES public.schedules (id) MATCH SIMPLE
    ON UPDATE RESTRICT
    ON DELETE CASCADE
    NOT VALID;


ALTER TABLE IF EXISTS public.schedules_recipes
    ADD CONSTRAINT fk_schedule_recipe FOREIGN KEY (recipe_id)
    REFERENCES public.recipes (id) MATCH SIMPLE
    ON UPDATE RESTRICT
    ON DELETE CASCADE
    NOT VALID;


ALTER TABLE IF EXISTS public.recipes_steps
    ADD CONSTRAINT fk_recipes_steps FOREIGN KEY (recipe_id)
    REFERENCES public.recipes (id) MATCH SIMPLE
    ON UPDATE RESTRICT
    ON DELETE CASCADE
    NOT VALID;

CREATE INDEX idx_ingredients on public.ingredients (id, name);
CREATE INDEX idx_recipes on public.recipes (id, name);
CREATE INDEX idx_users on public.users (id, username);