-- テストユーザー
INSERT INTO users (id, display_name, avatar_url) VALUES
  ('seed_user_001', '田中 太郎',   'https://i.pravatar.cc/150?img=1'),
  ('seed_user_002', '鈴木 花子',   'https://i.pravatar.cc/150?img=2'),
  ('seed_user_003', '佐藤 次郎',   'https://i.pravatar.cc/150?img=3')
ON CONFLICT DO NOTHING;
