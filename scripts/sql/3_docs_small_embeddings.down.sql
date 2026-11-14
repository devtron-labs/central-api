/*
 * Copyright (c) 2024. Devtron Inc.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

-- Rollback: back to BAAI/bge-large-en-v1.5 vectors. The table is emptied again.

TRUNCATE TABLE "public"."documents";

ALTER TABLE "public"."documents" DROP COLUMN IF EXISTS "content_hash";
ALTER TABLE "public"."documents" ALTER COLUMN "embedding" TYPE vector(1024);

CREATE INDEX IF NOT EXISTS documents_embedding_idx
    ON "public"."documents" USING ivfflat (embedding vector_cosine_ops)
    WITH (lists = 100);

DELETE FROM "public"."schema_migrations" WHERE version = '3';
