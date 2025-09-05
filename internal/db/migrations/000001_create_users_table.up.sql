CREATE TABLE "client_account" (
                                  "id" uuid PRIMARY KEY,
                                  "name" varchar NOT NULL,
                                  "created_at" timestamp NOT NULL,
                                  "updated_ad" timestamp
);

CREATE TABLE "user_account" (
                                "id" uuid PRIMARY KEY,
                                "name" varchar NOT NULL,
                                "surname" varchar,
                                "email" varchar NOT NULL,
                                "client_account_id" uuid,
                                "admin" bool NOT NULL,
                                "created_at" timestamp NOT NULL,
                                "updated_ad" timestamp
);

CREATE TABLE "product" (
                           "id" UUID PRIMARY KEY,
                           "referencial_id" uuid NOT NULL,
                           "name" varchar NOT NULL,
                           "description" varchar NOT NULL,
                           "stock" integer NOT NULL,
                           "status" varchar,
                           "clientAccount" uuid,
                           "created_at" timestamp,
                           "update_at" timestamp
);

CREATE TABLE "sku" (
                       "id" uuid PRIMARY KEY,
                       "name_sku" varchar NOT NULL,
                       "status" bool,
                       "product_id" uuid,
                       "created_at" timestamp,
                       "updated_at" timestamp
);

CREATE TABLE "movement" (
                            "id" uuid PRIMARY KEY,
                            "count" integer,
                            "product_id" uuid,
                            "date_limit" time,
                            "request_id" uuid,
                            "create_at" timestamp,
                            "updated_at" timestamp
);

CREATE TABLE "request_per_product" (
                                       "id" uuid PRIMARY KEY,
                                       "request_id" uuid,
                                       "product_id" uuid
);

CREATE TABLE "request" (
                           "id" uuid PRIMARY KEY,
                           "client_account_id" uuid,
                           "status" varchar,
                           "create_at" timestamp,
                           "updated_at" timestamp
);

CREATE TABLE "documents" (
                             "id" uuid PRIMARY KEY,
                             "s3_path" varchar,
                             "request_id" uuid,
                             "created_at" timestamp,
                             "update_at" timestamp
);

ALTER TABLE "user_account" ADD CONSTRAINT "clienta_users" FOREIGN KEY ("client_account_id") REFERENCES "client_account" ("id");

ALTER TABLE "product" ADD CONSTRAINT "product_client" FOREIGN KEY ("clientAccount") REFERENCES "client_account" ("id");

ALTER TABLE "sku" ADD CONSTRAINT "sku_products" FOREIGN KEY ("product_id") REFERENCES "product" ("id");

ALTER TABLE "movement" ADD CONSTRAINT "movement_products" FOREIGN KEY ("product_id") REFERENCES "product" ("id");

ALTER TABLE "documents" ADD CONSTRAINT "document_request" FOREIGN KEY ("request_id") REFERENCES "request" ("id");

ALTER TABLE "request" ADD CONSTRAINT "requets_client" FOREIGN KEY ("client_account_id") REFERENCES "client_account" ("id");

ALTER TABLE "movement" ADD CONSTRAINT "movement_request" FOREIGN KEY ("request_id") REFERENCES "request" ("id");

ALTER TABLE "request_per_product" ADD CONSTRAINT "request_movement" FOREIGN KEY ("request_id") REFERENCES "request" ("id");

ALTER TABLE "request_per_product" ADD CONSTRAINT "product_request" FOREIGN KEY ("product_id") REFERENCES "movement" ("id");
