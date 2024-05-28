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

-- Sequence and defined type
CREATE SEQUENCE IF NOT EXISTS id_release_note;

-- Table Definition
CREATE TABLE IF NOT EXISTS "public"."release_notes"
(
    "id"           int4        NOT NULL DEFAULT nextval('id_release_note'::regclass),
    "release_note" text        NOT NULL,
    "is_active"    bool        NOT NULL,
    "created_on"   timestamptz NOT NULL,
    "updated_on"   timestamptz,
    PRIMARY KEY ("id")
);

--> create unique index where for release note is_active is true
CREATE UNIQUE INDEX IF NOT EXISTS only_one_row_with_active_release_note ON release_notes (is_active) WHERE (is_active);

