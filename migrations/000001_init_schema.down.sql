DROP TABLE IF EXISTS feedback_embeddings;
DROP TABLE IF EXISTS viewing_history;
DROP TABLE IF EXISTS likes;
DROP TABLE IF EXISTS comments;
DROP TABLE IF EXISTS messages;
DROP TABLE IF EXISTS message_rooms;
DROP TABLE IF EXISTS orders;
DROP TABLE IF EXISTS damage_reports;
DROP TABLE IF EXISTS damages;
DROP TABLE IF EXISTS product_models;
DROP TABLE IF EXISTS product_images;
DROP TABLE IF EXISTS products;
DROP TABLE IF EXISTS damage_detection_summaries;
DROP TABLE IF EXISTS categories;
DROP TABLE IF EXISTS users;

DROP TYPE IF EXISTS order_status;
DROP TYPE IF EXISTS model_status;
DROP TYPE IF EXISTS damage_type;
DROP TYPE IF EXISTS image_angle;
DROP TYPE IF EXISTS product_status;
DROP TYPE IF EXISTS product_condition;

DROP EXTENSION IF EXISTS vector;
