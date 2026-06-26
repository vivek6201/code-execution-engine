-- Create "job_records" table
CREATE TABLE "job_records" (
  "id" uuid NOT NULL,
  "user_id" uuid NOT NULL,
  "api_key_id" uuid NOT NULL,
  "language" text NOT NULL,
  "code" text NOT NULL,
  "status" text NULL,
  "output" text NULL,
  "error" text NULL,
  "fatal_error" text NULL,
  "time_ms" bigint NULL,
  "memory_kb" bigint NULL,
  "total" bigint NULL,
  "passed" bigint NULL,
  "test_cases" text NULL,
  "time_limit_ms" bigint NULL,
  "memory_limit_kb" bigint NULL,
  "callback_url" text NULL,
  "created_at" timestamptz NULL,
  "updated_at" timestamptz NULL,
  PRIMARY KEY ("id")
);
-- Create index "idx_job_records_api_key_id" to table: "job_records"
CREATE INDEX "idx_job_records_api_key_id" ON "job_records" ("api_key_id");
-- Create index "idx_job_records_user_id" to table: "job_records"
CREATE INDEX "idx_job_records_user_id" ON "job_records" ("user_id");
-- Create "users" table
CREATE TABLE "users" (
  "id" uuid NOT NULL DEFAULT gen_random_uuid(),
  "email" text NOT NULL,
  "password_hash" text NOT NULL,
  "name" text NULL,
  "plan" character varying(20) NULL DEFAULT 'basic',
  "created_at" timestamptz NULL,
  "updated_at" timestamptz NULL,
  PRIMARY KEY ("id")
);
-- Create index "idx_users_email" to table: "users"
CREATE UNIQUE INDEX "idx_users_email" ON "users" ("email");
-- Create "api_keys" table
CREATE TABLE "api_keys" (
  "id" uuid NOT NULL DEFAULT gen_random_uuid(),
  "user_id" uuid NOT NULL,
  "name" text NULL,
  "key_hash" text NOT NULL,
  "prefix" text NOT NULL,
  "last_used_at" timestamptz NULL,
  "expires_at" timestamptz NULL,
  "revoked" boolean NULL DEFAULT false,
  "created_at" timestamptz NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "fk_users_api_keys" FOREIGN KEY ("user_id") REFERENCES "users" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION
);
-- Create index "idx_api_keys_key_hash" to table: "api_keys"
CREATE UNIQUE INDEX "idx_api_keys_key_hash" ON "api_keys" ("key_hash");
-- Create index "idx_api_keys_user_id" to table: "api_keys"
CREATE INDEX "idx_api_keys_user_id" ON "api_keys" ("user_id");
