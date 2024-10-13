/*
  sql for postgresql
*/

/* CREATE EXTENSION IF NOT EXISTS "uuid-ossp"; */

CREATE TABLE job
(
    id uuid NOT NULL DEFAULT uuid_generate_v4(),
    name text NOT NULL,
    description text NOT NULL,
    author text NOT NULL,
    members text[],
    created_at timestamp with time zone NOT NULL DEFAULT now(),
    PRIMARY KEY (id)
);

CREATE TABLE action
(
    id uuid NOT NULL DEFAULT uuid_generate_v4(),
    name text NOT NULL,
    payload json NOT NULL,
    created_at timestamp with time zone NOT NULL DEFAULT now(),
    job_id uuid NOT NULL,
    PRIMARY KEY (id),
    CONSTRAINT job_id FOREIGN KEY (job_id)
        REFERENCES job (id) MATCH SIMPLE
        ON UPDATE NO ACTION
        ON DELETE CASCADE
        NOT VALID
);

CREATE TABLE trigger
(
    id uuid NOT NULL DEFAULT uuid_generate_v4(),
    name text NOT NULL,
    payload json NOT NULL,
    created_at timestamp with time zone NOT NULL DEFAULT now(),
    job_id uuid NOT NULL,
    PRIMARY KEY (id),
        CONSTRAINT job_id FOREIGN KEY (job_id)
        REFERENCES job (id) MATCH SIMPLE
        ON UPDATE NO ACTION
        ON DELETE CASCADE
        NOT VALID
);
