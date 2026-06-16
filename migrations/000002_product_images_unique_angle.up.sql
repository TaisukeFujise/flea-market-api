CREATE UNIQUE INDEX product_images_product_id_angle_unique
    ON product_images (product_id, angle)
    WHERE deleted_at IS NULL;
