CREATE EXTENSION IF NOT EXISTS vector;

CREATE TYPE product_condition AS ENUM ('good', 'fair', 'poor');

CREATE TYPE product_status AS ENUM ('on_sale', 'sold_out');

CREATE TYPE image_angle AS ENUM ('front', 'back', 'right', 'left', 'top');

CREATE TYPE damage_type AS ENUM ('scratch', 'dirt', 'wear');

CREATE TYPE model_status AS ENUM ('pending', 'processing', 'done', 'failed');

CREATE TYPE order_status AS ENUM ('pending', 'completed', 'cancelled');

CREATE TABLE users (
	id VARCHAR(255) PRIMARY KEY,
	display_name VARCHAR(255) NOT NULL,
	avatar_url VARCHAR(500),
	created_at TIMESTAMP NOT NULL DEFAULT NOW(),
	updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
	deleted_at TIMESTAMP
);

CREATE TABLE categories (
	id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
	parent_id UUID REFERENCES categories(id),
	name VARCHAR(255) NOT NULL,
	created_at TIMESTAMP NOT NULL DEFAULT NOW(),
	updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE TABLE damage_detection_summaries (
	id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
	user_id VARCHAR(255) NOT NULL REFERENCES users(id),
	condition product_condition,
	condition_note TEXT,
	status TEXT NOT NULL DEFAULT 'processing',
	created_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE TABLE products (
	id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
	user_id VARCHAR(255) NOT NULL REFERENCES users(id),
	category_id UUID NOT NULL REFERENCES categories(id),
	title VARCHAR(255) NOT NULL,
	description TEXT,
	price INT NOT NULL,
	condition product_condition NOT NULL,
	condition_note TEXT,
	status product_status NOT NULL,
	created_at TIMESTAMP NOT NULL DEFAULT NOW(),
	updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
	deleted_at TIMESTAMP
);

CREATE TABLE product_images (
	id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
	product_id UUID REFERENCES products(id),
	summary_id UUID REFERENCES damage_detection_summaries(id),
	url VARCHAR(500) NOT NULL,
	angle image_angle,
	created_at TIMESTAMP NOT NULL DEFAULT NOW(),
	deleted_at TIMESTAMP
);

CREATE UNIQUE INDEX product_images_product_id_angle_unique
    ON product_images (product_id, angle)
    WHERE deleted_at IS NULL;

CREATE TABLE product_models (
	id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
	product_id UUID NOT NULL REFERENCES products(id),
	glb_url VARCHAR(500),
	job_id VARCHAR(255),
	status model_status NOT NULL,
	created_at TIMESTAMP NOT NULL DEFAULT NOW(),
	updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
	deleted_at TIMESTAMP
);

CREATE TABLE damages (
	id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
	image_id UUID NOT NULL REFERENCES product_images(id),
	damage_type damage_type NOT NULL,
	bbox_x1 INT,
	bbox_y1 INT,
	bbox_x2 INT,
	bbox_y2 INT,
	model_x FLOAT,
	model_y FLOAT,
	model_z FLOAT,
	description TEXT,
	created_at TIMESTAMP NOT NULL DEFAULT NOW(),
	deleted_at TIMESTAMP
);

CREATE TABLE damage_reports (
	id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
	product_id UUID NOT NULL REFERENCES products(id),
	user_id VARCHAR(255) NOT NULL REFERENCES users(id),
	image_id UUID REFERENCES product_images(id),
	damage_type damage_type NOT NULL,
	bbox_x1 INT,
	bbox_y1 INT,
	bbox_x2 INT,
	bbox_y2 INT,
	model_x FLOAT,
	model_y FLOAT,
	model_z FLOAT,
	description TEXT,
	created_at TIMESTAMP NOT NULL DEFAULT NOW(),
	deleted_at TIMESTAMP
);

CREATE TABLE orders (
	id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
	product_id UUID NOT NULL REFERENCES products(id),
	buyer_id VARCHAR(255) NOT NULL REFERENCES users(id),
	seller_id VARCHAR(255) NOT NULL REFERENCES users(id),
	price INT NOT NULL,
	status order_status NOT NULL,
	created_at TIMESTAMP NOT NULL DEFAULT NOW(),
	updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE TABLE ratings (
	id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
	order_id UUID NOT NULL REFERENCES orders(id) UNIQUE,
	rater_id VARCHAR(255) NOT NULL REFERENCES users(id),
	ratee_id VARCHAR(255) NOT NULL REFERENCES users(id),
	score INT NOT NULL CHECK (score BETWEEN 1 AND 5),
	created_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE TABLE message_rooms (
	id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
	order_id UUID NOT NULL REFERENCES orders(id),
	buyer_id VARCHAR(255) NOT NULL REFERENCES users(id),
	seller_id VARCHAR(255) NOT NULL REFERENCES users(id),
	created_at TIMESTAMP NOT NULL DEFAULT NOW(),
	deleted_at TIMESTAMP
);

CREATE TABLE messages (
	id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
	room_id UUID NOT NULL REFERENCES message_rooms(id),
	sender_id VARCHAR(255) NOT NULL REFERENCES users(id),
	content TEXT NOT NULL,
	created_at TIMESTAMP NOT NULL DEFAULT NOW(),
	deleted_at TIMESTAMP
);

CREATE TABLE comments (
	id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
	product_id UUID NOT NULL REFERENCES products(id),
	user_id VARCHAR(255) NOT NULL REFERENCES users(id),
	content TEXT NOT NULL,
	created_at TIMESTAMP NOT NULL DEFAULT NOW(),
	deleted_at TIMESTAMP
);

CREATE TABLE likes (
	id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
	user_id VARCHAR(255) NOT NULL REFERENCES users(id),
	product_id UUID NOT NULL REFERENCES products(id),
	created_at TIMESTAMP NOT NULL DEFAULT NOW(),
	UNIQUE (user_id, product_id)
);

CREATE TABLE viewing_history (
	id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
	user_id VARCHAR(255) NOT NULL REFERENCES users(id),
	product_id UUID NOT NULL REFERENCES products(id),
	viewed_at TIMESTAMP NOT NULL DEFAULT NOW(),
	UNIQUE (user_id, product_id)
);

CREATE TABLE feedback_embeddings (
	id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
	damage_report_id UUID NOT NULL REFERENCES damage_reports(id),
	category_id UUID NOT NULL REFERENCES categories(id),
	embedding vector(3072) NOT NULL,
	created_at TIMESTAMP NOT NULL DEFAULT NOW()
);
