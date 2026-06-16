-- 注文済み商品を sold_out に更新
-- pending / completed の注文がある商品のみ（cancelled は on_sale に戻る）
UPDATE products SET status = 'sold_out' WHERE id IN (
  '10000000-0000-0000-0000-000000000001', -- iPhone 14 Pro        → pending 注文あり
  '10000000-0000-0000-0000-000000000003'  -- Sony WH-1000XM5      → completed 注文あり
);

-- 注文
-- order 1: pending   鈴木花子 が 田中太郎 の iPhone を購入
-- order 2: completed 田中太郎 が 佐藤次郎 の Sony ヘッドホンを購入・受取済み
-- order 3: cancelled 佐藤次郎 が 鈴木花子 の Nintendo Switch を購入→キャンセル
INSERT INTO orders (id, product_id, buyer_id, seller_id, price, status) VALUES
  (
    '30000000-0000-0000-0000-000000000001',
    '10000000-0000-0000-0000-000000000001',
    'seed_user_002',
    'seed_user_001',
    89000,
    'pending'
  ),
  (
    '30000000-0000-0000-0000-000000000002',
    '10000000-0000-0000-0000-000000000003',
    'seed_user_001',
    'seed_user_003',
    32000,
    'completed'
  ),
  (
    '30000000-0000-0000-0000-000000000003',
    '10000000-0000-0000-0000-000000000011',
    'seed_user_003',
    'seed_user_002',
    28000,
    'cancelled'
  )
ON CONFLICT DO NOTHING;

-- メッセージルーム（各注文に1部屋）
INSERT INTO message_rooms (id, order_id, buyer_id, seller_id) VALUES
  (
    '40000000-0000-0000-0000-000000000001',
    '30000000-0000-0000-0000-000000000001',
    'seed_user_002',
    'seed_user_001'
  ),
  (
    '40000000-0000-0000-0000-000000000002',
    '30000000-0000-0000-0000-000000000002',
    'seed_user_001',
    'seed_user_003'
  ),
  (
    '40000000-0000-0000-0000-000000000003',
    '30000000-0000-0000-0000-000000000003',
    'seed_user_003',
    'seed_user_002'
  )
ON CONFLICT DO NOTHING;
