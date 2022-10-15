-- +goose Up
-- +goose StatementBegin
CREATE TABLE slots
(
    slot_id          uuid DEFAULT gen_random_uuid() PRIMARY KEY,
    slot_description text not null
);

CREATE TABLE banners
(
    banner_id          uuid DEFAULT gen_random_uuid() PRIMARY KEY,
    banner_description text not null
);

CREATE TABLE social_groups
(
    group_id          uuid DEFAULT gen_random_uuid() PRIMARY KEY,
    group_description text not null
);

CREATE TABLE slot_banners
(
    slot_id   uuid REFERENCES slots (slot_id) ON DELETE CASCADE,
    banner_id uuid REFERENCES banners (banner_id) ON DELETE CASCADE,
    CONSTRAINT slot_banner_pkey PRIMARY KEY (slot_id, banner_id)
);

CREATE TABLE banner_stats
(
    slot_id uuid REFERENCES slots (slot_id) ON DELETE SET NULL,
    group_id uuid REFERENCES social_groups (group_id) ON DELETE SET NULL,
    banner_id uuid REFERENCES banners (banner_id) ON DELETE SET NULL,
    clicks_amount int NOT NULL,
    shows_amount int NOT NULL,
    CONSTRAINT banner_stats_pkey PRIMARY KEY (slot_id, group_id, banner_id)
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
drop table if exists banner_stats;
drop table if exists slot_banners;
drop table if exists slots;
drop table if exists banners;
drop table if exists social_groups;
-- +goose StatementEnd
