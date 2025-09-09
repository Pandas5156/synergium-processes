-- Удаление всех ролей
DELETE FROM roles WHERE name IN (
  'admin',
  'supervisor',
  'curator',
  'operator',
  'decision_maker'
);