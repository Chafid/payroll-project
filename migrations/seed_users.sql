-- Example insert for admin
INSERT INTO users (id, username, password, role, level_id, created_at, created_by)
VALUES (
  gen_random_uuid(),
  'admin',
  '$2a$10$MWTyPjrQV.TopF4.oU/y7./yVoY45gz4ZDRdu85HhaRHcRdTkbdeu[', -- hashed 'admin123'
  'admin',
  NULL,
  NOW(),
  NULL
);

INSERT INTO employee_levels (id, name, base_salary) VALUES
  (gen_random_uuid(), 'Junior', 5000000),
  (gen_random_uuid(), 'Mid', 8000000),
  (gen_random_uuid(), 'Senior', 12000000);

-- Seed 100 employees with random level_id
DO $$
DECLARE
  level_ids UUID[];
  level_count INT;
  rand_index INT;
BEGIN
  -- Get all level_ids into an array
  SELECT array_agg(id) INTO level_ids FROM employee_levels;
  level_count := array_length(level_ids, 1);

  FOR i IN 1..100 LOOP
    rand_index := trunc(random() * level_count + 1);
    INSERT INTO users (id, username, password, role, level_id, created_at, updated_at)
    VALUES (
      gen_random_uuid(),
      'employee' || lpad(i::text, 3, '0'),
      crypt('employee' || lpad(i::text, 3, '0'), gen_salt('bf')),
      'employee',
      level_ids[rand_index],
      NOW(),
      NOW()
    );
  END LOOP;
END $$;