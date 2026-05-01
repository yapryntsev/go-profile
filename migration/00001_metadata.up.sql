create type processing_status as enum ('pending', 'active', 'completed');

create table metadata (
    id                uuid primary key                  default pg_catalog.gen_random_uuid(),
    user_id           text                     not null,
    file_name         text                     not null,
    mime_type         text                     not null,
    width             int                      not null,
    height            int                      not null,
    size_bytes        bigint                   not null,
    s3_key            text                     not null,
    thumbnail_s3_keys jsonb,
    processing_status processing_status                 default 'pending',
    created_at        timestamp with time zone          default now(),
    updated_at        timestamp with time zone          default now(),
    deleted_at        timestamp with time zone
);

create index idx_metadata_user_id on metadata(user_id) where deleted_at is null;