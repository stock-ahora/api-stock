CREATE TABLE if not exists "client_account"
(
    "id"         uuid PRIMARY KEY,
    "name"       varchar   NOT NULL,
    "created_at" timestamp NOT NULL,
    "updated_ad" timestamp
);

CREATE TABLE if not exists "user_account"
(
    "id"                uuid PRIMARY KEY,
    "name"              varchar   NOT NULL,
    "surname"           varchar,
    "email"             varchar   NOT NULL,
    "client_account_id" uuid,
    "admin"             bool      NOT NULL,
    "created_at"        timestamp NOT NULL,
    "updated_ad"        timestamp
);

CREATE TABLE if not exists "product"
(
    "id"             UUID PRIMARY KEY,
    "referencial_id" uuid    NOT NULL,
    "name"           varchar NOT NULL,
    "description"    varchar NOT NULL,
    "stock"          integer NOT NULL,
    "status"         varchar,
    "clientAccount"  uuid,
    "created_at"     timestamp,
    "update_at"      timestamp
);

CREATE TABLE if not exists "sku"
(
    "id"         uuid PRIMARY KEY,
    "name_sku"   varchar NOT NULL,
    "status"     bool,
    "product_id" uuid,
    "created_at" timestamp,
    "updated_at" timestamp
);

CREATE TABLE if not exists "movement"
(
    "id"               uuid PRIMARY KEY,
    "count"            integer,
    "product_id"       uuid,
    "date_limit"       time,
    "request_id"       uuid,
    "movement_type_id" serial,
    "create_at"        timestamp,
    "updated_at"       timestamp
);

CREATE TABLE if not exists "movements_type"
(
    "id"          serial PRIMARY KEY,
    "name"        varchar,
    "description" varchar
);

CREATE TABLE if not exists "request_per_product"
(
    "id"          uuid PRIMARY KEY,
    "product_id"  uuid,
    "movement_id" uuid,
    "request_id"  uuid
);

CREATE TABLE if not exists "request"
(
    "id"                uuid PRIMARY KEY,
    "client_account_id" uuid,
    "status"            varchar,
    "create_at"         timestamp,
    "updated_at"        timestamp
);

CREATE TABLE if not exists "documents"
(
    "id"         uuid PRIMARY KEY,
    "s3_path"    varchar,
    "request_id" uuid,
    "created_at" timestamp,
    "update_at"  timestamp
);


-- Estas líneas son redundantes ya que las restricciones ya fueron añadidas arriba.
-- Si realmente necesitas verificación condicional, usa bloques DO:

DO $$
BEGIN
  IF NOT EXISTS (SELECT 1 FROM pg_constraint WHERE conname = 'clienta_users') THEN
    EXECUTE 'ALTER TABLE "user_account" ADD CONSTRAINT "clienta_users" FOREIGN KEY ("client_account_id") REFERENCES "client_account" ("id")';
  END IF;
END $$;

DO $$
BEGIN
  IF NOT EXISTS (SELECT 1 FROM pg_constraint WHERE conname = 'product_client') THEN
    EXECUTE 'ALTER TABLE "product" ADD CONSTRAINT "product_client" FOREIGN KEY ("clientAccount") REFERENCES "client_account" ("id")';
  END IF;
END $$;

DO $$
BEGIN
  IF NOT EXISTS (SELECT 1 FROM pg_constraint WHERE conname = 'sku_products') THEN
    EXECUTE 'ALTER TABLE "sku" ADD CONSTRAINT "sku_products" FOREIGN KEY ("product_id") REFERENCES "product" ("id")';
  END IF;
END $$;

DO $$
BEGIN
  IF NOT EXISTS (SELECT 1 FROM pg_constraint WHERE conname = 'movement_products') THEN
    EXECUTE 'ALTER TABLE "movement" ADD CONSTRAINT "movement_products" FOREIGN KEY ("product_id") REFERENCES "product" ("id")';
  END IF;
END $$;

DO $$
BEGIN
  IF NOT EXISTS (SELECT 1 FROM pg_constraint WHERE conname = 'document_request') THEN
    EXECUTE 'ALTER TABLE "documents" ADD CONSTRAINT "document_request" FOREIGN KEY ("request_id") REFERENCES "request" ("id")';
  END IF;
END $$;

DO $$
BEGIN
  IF NOT EXISTS (SELECT 1 FROM pg_constraint WHERE conname = 'requets_client') THEN
    EXECUTE 'ALTER TABLE "request" ADD CONSTRAINT "requets_client" FOREIGN KEY ("client_account_id") REFERENCES "client_account" ("id")';
  END IF;
END $$;

DO $$
BEGIN
  IF NOT EXISTS (SELECT 1 FROM pg_constraint WHERE conname = 'movement_request') THEN
    EXECUTE 'ALTER TABLE "movement" ADD CONSTRAINT "movement_request" FOREIGN KEY ("request_id") REFERENCES "request" ("id")';
  END IF;
END $$;

DO $$
BEGIN
  IF NOT EXISTS (SELECT 1 FROM pg_constraint WHERE conname = 'request_movement') THEN
    EXECUTE 'ALTER TABLE "request_per_product" ADD CONSTRAINT "request_movement" FOREIGN KEY ("request_id") REFERENCES "request" ("id")';
  END IF;
END $$;

DO $$
BEGIN
  IF NOT EXISTS (SELECT 1 FROM pg_constraint WHERE conname = 'product_request') THEN
    EXECUTE 'ALTER TABLE "request_per_product" ADD CONSTRAINT "product_request" FOREIGN KEY ("product_id") REFERENCES "product" ("id")';
  END IF;
END $$;

DO $$
BEGIN
  IF NOT EXISTS (SELECT 1 FROM pg_constraint WHERE conname = 'mov_request') THEN
    EXECUTE 'ALTER TABLE "request_per_product" ADD CONSTRAINT "mov_request" FOREIGN KEY ("movement_id") REFERENCES "movement" ("id")';
  END IF;
END $$;

DO $$
BEGIN
  IF NOT EXISTS (SELECT 1 FROM pg_constraint WHERE conname = 'type_movement') THEN
    EXECUTE 'ALTER TABLE "movement" ADD CONSTRAINT "type_movement" FOREIGN KEY ("movement_type_id") REFERENCES "movements_type" ("id")';
  END IF;
END $$;