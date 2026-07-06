
INSERT INTO digmark.vertical (id, name) VALUES (-1, 'None');
INSERT INTO digmark.vertical (name) VALUES
  ('Dating'),
  ('E-commerce'),
  ('Education'),
  ('Health & Wellness'),
  ('Home Services'),
  ('Housing'),
  ('Legal Services'),
  ('Mobile Apps'),
  ('Personal Finance'),
  ('Travel')
;


-- digmark.campaign
DO
$do$
DECLARE
  v_country_fk bigint; v_iso2 char(2);
  v_vertical_fk bigint; v_vertical text;
BEGIN

FOR i IN 1..200 LOOP
  SELECT (ceil(random() * 249))::bigint INTO v_country_fk;
  SELECT iso2 INTO v_iso2 FROM geo.country WHERE id = v_country_fk;
  
  SELECT (ceil(random() * 10))::bigint INTO v_vertical_fk;
  SELECT name INTO v_vertical FROM digmark.vertical WHERE id = v_vertical_fk;

  INSERT INTO digmark.campaign (country_fk, daily_budget_eur, is_active, manager, name, vertical_fk)
  
  SELECT 
    v_country_fk AS country_fk,
    (floor(random() * 1000))::numeric AS daily_budget_eur,
    CASE WHEN random() < 0.3 THEN false ELSE true END AS is_active,
    CASE WHEN random() < 0.33 THEN 'Anna'::digmark.manager WHEN random() < 0.5 THEN 'Michael'::digmark.manager ELSE 'Paul'::digmark.manager END AS manager,
    v_iso2 || ' - ' || v_vertical AS name,
    v_vertical_fk AS vertical_fk
  FROM generate_series(1,200)
  ON CONFLICT DO NOTHING;
  
END LOOP;
END
$do$;


-- digmark.campaign_performance
DO
$do$
DECLARE 
  v_camp_id bigint;
  v_num_days int;
  v_impressions int; v_clicks int; v_conversions int; v_revenue_eur numeric; v_spend_eur numeric;
BEGIN

  FOR v_camp_id IN SELECT id FROM digmark.campaign LOOP

    -- simulate between 1 and 95 days of performance data for each campaign, starting from today and working backwards
    SELECT (ceil(random() * 95))::int INTO v_num_days;

    FOR i IN 0..v_num_days -1 LOOP

      -- simulate funnel
      SELECT floor(random() * 10000)::int INTO v_impressions;
      SELECT floor(random() * v_impressions)::int INTO v_clicks;
      SELECT floor(random() * v_clicks)::int INTO v_conversions;
      SELECT (random() * (50 - 5) + 5) / 10 * v_conversions INTO v_revenue_eur; -- 0.5 to 5 EUR per conversion
      SELECT (random() * (12 - 6) + 6) / 10 * v_revenue_eur INTO v_spend_eur; -- 60% to 120% of revenue
  
      INSERT INTO digmark.campaign_performance (campaign_fk, clicks, conversions, day_cet, impressions, revenue_eur, spend_eur)
      SELECT 
        v_camp_id AS campaign_fk,
        v_clicks AS clicks,
        v_conversions AS conversions,
        current_date - i AS day_cet,
        v_impressions AS impressions,
        v_revenue_eur AS revenue_eur,
        v_spend_eur AS spend_eur;
    END LOOP;
  END LOOP;
END
$do$;


INSERT INTO digmark.launcher_fb (name, manager, fan_page, daily_budget_eur, status, step, message) VALUES 
  ('FB - Acc1 - BR - Daating', 'Anna', 'Page1', 10, 'Invalid', 0, 'invalid vertical: Daating'),
  ('FB - Acc1 - BS - E-commerce', 'Michael', 'Page2', 15, 'Ready', 0, ''),
  ('FB - Acc2 - BT - Education', 'Paul', 'Page3', 20, 'Processing', 1, ''),
  ('FB - Acc1 - BV - Health & Wellness', 'Anna', 'Page1', 10, 'Completed', 2, ''),
  ('FB - Acc1 - BW - Home Services', 'Michael', 'Page2', 15, 'Failed', 0, 'Facebook API error');


INSERT INTO digmark.launcher_gads (name, manager, daily_budget_eur, status, step, message) VALUES 
  ('GAds - Acc2 - XX - Housing', 'Paul', 20, 'Invalid', 0, 'invalid country: XX'),
  ('GAds - Acc1 - BZ - Legal Services', 'Anna', 10, 'Ready', 0, ''),
  ('GAds - Acc1 - CA - Mobile Apps', 'Michael', 15, 'Processing', 1, ''),
  ('GAds - Acc2 - CC - Personal Finance', 'Paul', 20, 'Processing', 2, ''),
  ('GAds - Acc1 - CD - Travel', 'Anna', 10, 'Completed', 3, '');