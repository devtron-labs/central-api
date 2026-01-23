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

-- Rollback migration for pgvector documentation tables

-- Drop view
DROP VIEW IF EXISTS "public"."document_stats";

-- Drop trigger
DROP TRIGGER IF EXISTS update_documents_updated_at ON "public"."documents";

-- Drop function
DROP FUNCTION IF EXISTS update_updated_at_column();

-- Drop indexes
DROP INDEX IF EXISTS "public"."documents_embedding_idx";
DROP INDEX IF EXISTS "public"."documents_source_idx";
DROP INDEX IF EXISTS "public"."documents_title_idx";

-- Drop tables
DROP TABLE IF EXISTS "public"."documents";
DROP TABLE IF EXISTS "public"."schema_migrations";

-- Drop extension (optional - comment out if other tables use it)
-- DROP EXTENSION IF EXISTS vector;

