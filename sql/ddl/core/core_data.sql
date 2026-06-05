
-- core.array_type: between 0 and 2 values for each column
DO
$do$
BEGIN
  FOR i IN 1..100 LOOP
  
    INSERT INTO core.array_type (c_bool, c_date, c_enum, c_int, c_numeric, c_text)
    
    WITH lens AS (
      SELECT
        floor(random() * 3)::int AS c_bool_len,
        floor(random() * 3)::int AS c_date_len,
        floor(random() * 3)::int AS c_enum_len,
        floor(random() * 3)::int AS c_int_len,
        floor(random() * 3)::int AS c_numeric_len,
        floor(random() * 3)::int AS c_text_len
    )
    SELECT 
    
    CASE WHEN lens.c_bool_len = 0 THEN ARRAY[]::boolean[] ELSE 
      ARRAY(SELECT CASE WHEN random() < 0.5 THEN false ELSE true END
        FROM generate_series(1,lens.c_bool_len)
      ) END AS c_bool,

    CASE WHEN lens.c_date_len = 0 THEN ARRAY[]::date[] ELSE 
      ARRAY(SELECT (now() - floor(random() * 100 * 365.25) * INTERVAL '1 day')::date
        FROM generate_series(1,lens.c_date_len)
      ) END AS c_date,

    CASE WHEN lens.c_enum_len = 0 THEN ARRAY[]::core.mandatory_enum[] ELSE 
      ARRAY(SELECT (CASE WHEN random() < 0.5 THEN 'A' ELSE 'B' END)::core.mandatory_enum
        FROM generate_series(1,lens.c_enum_len)
      ) END AS c_enum,

    CASE WHEN lens.c_int_len = 0 THEN ARRAY[]::int[] ELSE 
      ARRAY(SELECT (floor(random() * 10000) - 5000)::int
        FROM generate_series(1,lens.c_int_len)
      ) END AS c_int,

    CASE WHEN lens.c_numeric_len = 0 THEN ARRAY[]::numeric[] ELSE 
      ARRAY(SELECT (random() * 1000 - 500)::numeric
        FROM generate_series(1,lens.c_numeric_len)
      ) END AS c_numeric,
      
    CASE WHEN lens.c_text_len = 0 THEN ARRAY[]::text[] ELSE 
      ARRAY(SELECT SUBSTRING(gen_random_uuid()::text FROM 1 FOR 8)
        FROM generate_series(1,lens.c_text_len)
      ) END AS c_text
      
    FROM lens;
    
 END LOOP;
END
$do$;


INSERT INTO core.default_value (c_default_text, c_suggested_text)
SELECT 
  CASE WHEN random() < 0.5 THEN 'Default text' ELSE SUBSTRING(gen_random_uuid()::text FROM 1 FOR 8) END AS c_default_text,
  CASE WHEN random() < 0.5 THEN 'Suggested text' ELSE SUBSTRING(gen_random_uuid()::text FROM 1 FOR 8) END AS c_suggested_text
FROM generate_series(1,100);


INSERT INTO core.mandatory_value (c_bool, c_date_cet, c_enum, c_int, c_numeric, c_table_fk, c_text, c_time)
SELECT 
  CASE WHEN random() < 0.5 THEN false ELSE true END AS c_bool,
  (now() - floor(random() * 100 * 365.25) * INTERVAL '1 day')::date AS c_date,
  (CASE WHEN random() < 0.5 THEN 'A' ELSE 'B' END)::core.mandatory_enum AS c_enum,
  (floor(random() * 10000) - 5000)::int AS c_int,
  (random() * 1000 - 500)::numeric AS c_numeric,
  (ceil(random() * 5))::bigint AS c_table_fk,
  SUBSTRING(gen_random_uuid()::text FROM 1 FOR 8) AS c_text,
  (current_time - floor(random() * 60 * 24) * INTERVAL '1 minute')::time AS c_time
FROM generate_series(1,100);


INSERT INTO core.optional_value (c_bool, c_date_cet, c_enum, c_int, c_numeric, c_table_fk, c_text, c_time)
SELECT 
  CASE WHEN random() < 0.5 THEN false ELSE true END AS c_bool,
  CASE WHEN random() < 0.5 THEN (now() - floor(random() * 100 * 365.25) * INTERVAL '1 day')::date
    ELSE '0001-01-01' END AS c_date_cet,
  CASE WHEN random() < 0.5 THEN (CASE WHEN random() < 0.5 THEN 'A' ELSE 'B' END)::core.optional_enum 
    ELSE ''::core.optional_enum END AS c_enum,
  CASE WHEN random() < 0.5 THEN (floor(random() * 1000) - 500)::int 
    ELSE 0 END AS c_int,
  CASE WHEN random() < 0.5 THEN (random() * 10000 - 500)::numeric 
    ELSE 0.0 END AS c_numeric,
  CASE WHEN random() < 0.5 THEN (ceil(random() * 249))::bigint
    ELSE -1 END AS c_table_fk,
  CASE WHEN random() < 0.5 THEN SUBSTRING(gen_random_uuid()::text FROM 1 FOR 8) 
    ELSE '' END AS c_text,
  CASE WHEN random() < 0.5 THEN (current_time - floor(random() * 60 * 24) * INTERVAL '1 minute')::time 
    ELSE '00:00:00'::time END AS c_time
FROM generate_series(1,100);


-- core.variant_type
DO
$do$
BEGIN
  FOR i IN 1..100 LOOP

    INSERT INTO core.variant_type (c_constrained_text, c_ip, c_long_text, c_money_amount, c_percent)
    SELECT 

    array_to_string(
      ARRAY(SELECT substr('ABCDEFGHIJKLMNOPQRSTUVWXYZ',((random()*(26-1)+1)::integer),1) from generate_series(1,6)),''
    ) AS c_constrained_text,

    CASE WHEN random() < 0.5 THEN -- v4
      concat(trunc(random() * 250 + 2), '.' , trunc(random() * 250 + 2), '.', trunc(random() * 250 + 2), '.', trunc(random() * 250 + 2))::inet
    ELSE -- v6
      concat(
        lpad(to_hex(floor(random() * 65536)::int), 4, '0'), ':', lpad(to_hex(floor(random() * 65536)::int), 4, '0'), ':',
        lpad(to_hex(floor(random() * 65536)::int), 4, '0'), ':', lpad(to_hex(floor(random() * 65536)::int), 4, '0'), ':',
        lpad(to_hex(floor(random() * 65536)::int), 4, '0'), ':', lpad(to_hex(floor(random() * 65536)::int), 4, '0'), ':',
        lpad(to_hex(floor(random() * 65536)::int), 4, '0'), ':', lpad(to_hex(floor(random() * 65536)::int), 4, '0'))::inet 
    END AS c_ip,

    array_to_string(
      ARRAY(SELECT substr('ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz+-#!%&                    ',((random()*(78-1)+1)::integer),1) from generate_series(1,1000)),''
    ) AS c_long_text,
    
    (random() * 100000 - 50000)::numeric AS c_money_amount,
    random()::numeric AS c_percent
    ;

  END LOOP;
END
$do$;