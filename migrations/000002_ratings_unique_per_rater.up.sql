ALTER TABLE ratings DROP CONSTRAINT ratings_order_id_key;
ALTER TABLE ratings ADD CONSTRAINT ratings_order_id_rater_id_key UNIQUE (order_id, rater_id);
