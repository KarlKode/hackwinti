/*
 * CALENDAR MODELS
 */
DROP TABLE IF EXISTS calendar_models_tmp;
CREATE TABLE calendar_models_tmp (
  calender INTEGER NOT NULL,
  description TEXT NOT NULL,
  monday VARCHAR(5) NOT NULL,
  tuesday VARCHAR(5) NOT NULL,
  wednesday VARCHAR(5) NOT NULL,
  thursday VARCHAR(5) NOT NULL,
  friday VARCHAR(5) NOT NULL,
  saturday VARCHAR(5) NOT NULL,
  sunday VARCHAR(5) NOT NULL,
  total VARCHAR(5) NOT NULL,
  day_overtime VARCHAR(5) NOT NULL,
  week_overtime VARCHAR(5) NOT NULL,
  overtime VARCHAR(5) NOT NULL
);

COPY calendar_models_tmp FROM '/Users/m/Code/hw/b/calendar_models.csv' DELIMITER ',' CSV HEADER;

DROP TABLE IF EXISTS calendar_models;
CREATE TABLE calendar_models (
  calender INTEGER NOT NULL,
  description TEXT NOT NULL,
  monday NUMERIC(4,1) NOT NULL,
  tuesday NUMERIC(4,1) NOT NULL,
  wednesday NUMERIC(4,1) NOT NULL,
  thursday NUMERIC(4,1) NOT NULL,
  friday NUMERIC(4,1) NOT NULL,
  saturday NUMERIC(4,1) NOT NULL,
  sunday NUMERIC(4,1) NOT NULL,
  total NUMERIC(4,1) NOT NULL,
  day_overtime NUMERIC(4,1) NOT NULL,
  week_overtime NUMERIC(4,1) NOT NULL,
  overtime NUMERIC(4,1) NOT NULL,
  PRIMARY KEY (calender)
);

INSERT INTO calendar_models (calender, description, monday, tuesday, wednesday, thursday, friday, saturday, sunday, total, day_overtime, week_overtime, overtime)
SELECT
  calender, description, COALESCE(nullif(monday, '')::NUMERIC, '0'), COALESCE(nullif(tuesday, '')::NUMERIC, '0'), COALESCE(nullif(wednesday, '')::NUMERIC, '0'), COALESCE(nullif(thursday, ''), '0')::NUMERIC, COALESCE(nullif(friday, '')::NUMERIC, '0'), COALESCE(nullif(saturday, '')::NUMERIC, '0'), COALESCE(nullif(sunday, '')::NUMERIC, '0'), COALESCE(nullif(total, '')::NUMERIC, '0'), COALESCE(nullif(day_overtime, '')::NUMERIC, '0'), COALESCE(nullif(week_overtime, '')::NUMERIC, '0'), COALESCE(regexp_replace(nullif(overtime, ''), '%', ''), '0')::NUMERIC
FROM calendar_models_tmp;

DROP TABLE IF EXISTS calendar_models_tmp;

/*
 * WORKCENTERS
 */
DROP TABLE IF EXISTS workcenters_tmp;
CREATE TABLE workcenters_tmp (
  empty CHAR(1) NOT NULL,
  workcenter VARCHAR(10) NOT NULL,
  description TEXT DEFAULT '',
  rate INTEGER DEFAULT -1,
  multipliable VARCHAR(3) DEFAULT 'no',
  combineable VARCHAR(6) DEFAULT 'costly',
  machines NUMERIC(4,1) DEFAULT -1,
  calender INTEGER DEFAULT -1,
  team_factor NUMERIC(6,3) DEFAULT -1,
  min_lead_time INTEGER DEFAULT -1,
  capacity NUMERIC(6,3) DEFAULT -1
);

COPY workcenters_tmp FROM '/Users/m/Code/hw/b/workcenters.csv' DELIMITER ',' CSV HEADER;

DROP TABLE IF EXISTS workcenters;
CREATE TABLE workcenters (
  workcenter VARCHAR(10) NOT NULL,
  description TEXT DEFAULT '',
  rate INTEGER DEFAULT -1,
  multipliable BOOLEAN DEFAULT FALSE,
  combineable VARCHAR(6) DEFAULT 'costly',
  machines NUMERIC(4,1) DEFAULT -1,
  calender INTEGER DEFAULT -1,
  team_factor NUMERIC(6,3) DEFAULT -1,
  min_lead_time INTEGER DEFAULT -1,
  capacity NUMERIC(6,3) DEFAULT -1
);

INSERT INTO workcenters (workcenter, description, rate, multipliable, combineable, machines, calender, team_factor, min_lead_time, capacity) SELECT
  workcenter, description, rate, multipliable = 'yes', combineable, machines, calender, team_factor, min_lead_time, capacity
FROM workcenters_tmp;

DROP TABLE IF EXISTS workcenters_tmp;

/*
 * MATERIALS
 */
DROP TABLE IF EXISTS materials_tmp;
CREATE TABLE materials_tmp (
  material VARCHAR(20) NOT NULL,
  description TEXT DEFAULT '',
  product_hierarchy varchar(20) NOT NULL,
  division VARCHAR(5) NOT NULL,
  grp varchar(20) NOT NULL,
  basic_material VARCHAR(20) NOT NULL,
  dimensions VARCHAR(30) NOT NULL,
  min_lot VARCHAR(15) NOT NULL,
  weight_net VARCHAR(15) NOT NULL,
  weight_net_unit VARCHAR(5) NOT NULL,
  weight_gross VARCHAR(15) NOT NULL,
  weight_gross_unit VARCHAR(5) NOT NULL
);

COPY materials_tmp FROM '/Users/m/Code/hw/b/material_master.csv' DELIMITER ',' CSV HEADER;

DROP TABLE IF EXISTS materials;
CREATE TABLE materials (
  material VARCHAR(20) NOT NULL,
  description TEXT DEFAULT '',
  product_hierarchy VARCHAR(20) NOT NULL,
  division INTEGER DEFAULT NULL,
  grp varchar(20) NOT NULL,
  basic_material VARCHAR(20) DEFAULT NULL,
  dimensions VARCHAR(30) DEFAULT NULL,
  min_lot NUMERIC(13,3) DEFAULT NULL,
  weight_net NUMERIC(13,3) DEFAULT NULL,
  weight_net_unit VARCHAR(5) DEFAULT NULL,
  weight_gross NUMERIC(13,3) DEFAULT NULL,
  weight_gross_unit VARCHAR(5) DEFAULT NULL
);

CREATE INDEX materials_ix ON materials (material);

INSERT INTO materials (material, description, product_hierarchy, division, grp, basic_material, dimensions, min_lot, weight_net, weight_net_unit, weight_gross, weight_gross_unit)
SELECT
  material, description, product_hierarchy, nullif(division, '')::INTEGER, grp, nullif(basic_material, ''), nullif(dimensions, ''), nullif(min_lot, '')::NUMERIC, regexp_replace(nullif(weight_net, ''), ',', '')::NUMERIC, weight_net_unit, regexp_replace(nullif(weight_gross, ''), ',', '')::NUMERIC, weight_gross_unit
FROM materials_tmp;

DROP TABLE IF EXISTS materials_tmp;

/*
 * ORDERS
 */
DROP TABLE IF EXISTS orders_tmp;
CREATE TABLE orders_tmp (
  order_id VARCHAR(20) NOT NULL,
  material VARCHAR(20) NOT NULL,
  material_description TEXT NOT NULL,
  amount DECIMAL(8,3) NOT NULL,
  start_time VARCHAR(20),
  end_time VARCHAR(20)
);

COPY orders_tmp FROM '/Users/m/Code/hw/b/production_order_headers.csv' DELIMITER ',' CSV HEADER;

DROP TABLE IF EXISTS orders;
CREATE TABLE orders (
  order_id VARCHAR(20) NOT NULL,
  material VARCHAR(20) NOT NULL,
  material_description TEXT NOT NULL,
  amount DECIMAL(8,3) NOT NULL,
  start_time TIMESTAMP,
  end_time TIMESTAMP,
  PRIMARY KEY (order_id)
);

CREATE INDEX orders_material_ix ON orders (material);

INSERT INTO orders (order_id, material, material_description, amount, start_time, end_time)
SELECT order_id, material, material_description, amount, to_timestamp(nullif(start_time, ''), 'mm/dd/yyyy'), to_timestamp(nullif(end_time, ''), 'mm/dd/yyyy')
FROM orders_tmp;

DROP TABLE IF EXISTS orders_tmp;

/*
 * ROUTINGS
 */
DROP TABLE IF EXISTS routings_tmp;
CREATE TABLE routings_tmp (
  material VARCHAR(20) NOT NULL,
  material_description TEXT NOT NULL,
  plan VARCHAR(10) NOT NULL,
  plan_counter VARCHAR(2),
  plan_status INTEGER NOT NULL,
  plan_description TEXT NOT NULL,
  operation VARCHAR(20) NOT NULL,
  workcenter VARCHAR(20) NOT NULL,
  operation_description TEXT NOT NULL,
  plan_setup_time NUMERIC(8,3) NOT NULL,
  plan_setup_time_unit VARCHAR(5),
  plan_processing_time NUMERIC(8,3),
  plan_processing_time_unit VARCHAR(5) NOT NULL,
  delivery_time VARCHAR(10) NOT NULL,
  net_price VARCHAR(10) NOT NULL,
  currency VARCHAR(5) NOT NULL,
  unit_price VARCHAR(10) NOT NULL
);

COPY routings_tmp FROM '/Users/m/Code/hw/b/routings.csv' DELIMITER ',' CSV HEADER;

DROP TABLE IF EXISTS routings;
CREATE TABLE routings (
  material VARCHAR(20) NOT NULL,
  material_description TEXT NOT NULL,
  plan VARCHAR(10) NOT NULL,
  plan_counter VARCHAR(2) NOT NULL,
  plan_status INTEGER NOT NULL,
  plan_description TEXT NOT NULL,
  operation VARCHAR(20) NOT NULL,
  workcenter VARCHAR(20) NOT NULL,
  operation_description TEXT DEFAULT NULL,
  plan_setup_time INTERVAL NOT NULL,
  plan_processing_time INTERVAL NOT NULL,
  delivery_time INTEGER DEFAULT NULL,
  net_price NUMERIC(10,3) DEFAULT NULL,
  currency VARCHAR(5) DEFAULT NULL,
  unit_price DECIMAL(10,3) DEFAULT NULL
);
CREATE INDEX routings_ix ON routings(material, plan, plan_counter,operation,workcenter);

INSERT INTO routings (material, material_description, plan, plan_counter, plan_status, plan_description, operation, workcenter, operation_description, plan_setup_time, plan_processing_time, delivery_time, net_price, currency, unit_price)
SELECT
  material, material_description, plan, plan_counter, plan_status, plan_description, nullif(operation, ''),
  nullif(workcenter, ''), nullif(operation_description, ''),
  case
    when plan_setup_time_unit = '' then '0 seconds'::INTERVAL
    when plan_setup_time_unit = 'H' then (plan_setup_time::TEXT || ' hours'::TEXT)::INTERVAL
    when plan_setup_time_unit = 'MIN' then (plan_setup_time::TEXT || ' minutes'::TEXT)::INTERVAL
  else
    '0 seconds'::INTERVAL
  end,
  case
    when plan_processing_time_unit = '' then '0 seconds'::INTERVAL
    when plan_processing_time_unit = 'H' then (plan_processing_time::TEXT || ' hours'::TEXT)::INTERVAL
    when plan_processing_time_unit = 'MIN' then (plan_processing_time::TEXT || ' minutes'::TEXT)::INTERVAL
  else
    '0 seconds'::INTERVAL
  end, nullif(delivery_time, '')::INTEGER, nullif(net_price, '')::NUMERIC, nullif(currency, ''), nullif(unit_price, '')::NUMERIC
FROM routings_tmp;

DROP TABLE IF EXISTS routings_tmp;

/*
 * OPERATIONS
 */
DROP TABLE IF EXISTS operations_tmp;
CREATE TABLE operations_tmp (
  order_id VARCHAR(10) NOT NULL,
  plan VARCHAR(10) NOT NULL,
  plan_counter VARCHAR(2),
  operation VARCHAR(20) NOT NULL,
  workcenter VARCHAR(20) NOT NULL,
  operation_description TEXT NOT NULL,
  start_time VARCHAR(10) NOT NULL,
  end_time VARCHAR(10) NOT NULL,
  delivery_time VARCHAR(10) NOT NULL,
  net_price VARCHAR(10) NOT NULL,
  currency VARCHAR(5) NOT NULL,
  unit_price VARCHAR(10) NOT NULL,
  plan_setup_time NUMERIC(8,3) NOT NULL,
  plan_setup_time_unit VARCHAR(5) NOT NULL,
  plan_processing_time NUMERIC(8,3) NOT NULL,
  plan_processing_time_unit VARCHAR(5) NOT NULL,
  setup_time VARCHAR(10) NOT NULL,
  setup_time_unit VARCHAR(5) NOT NULL,
  processing_time VARCHAR(10) NOT NULL,
  processing_time_unit VARCHAR(5) NOT NULL
);

COPY operations_tmp FROM '/Users/m/Code/hw/b/production_order_operations.csv' DELIMITER ',' CSV HEADER;

DROP TABLE IF EXISTS operations;
CREATE TABLE operations (
  order_id VARCHAR(10) NOT NULL,
  plan VARCHAR(10) DEFAULT NULL,
  plan_counter VARCHAR(2),
  operation VARCHAR(20) NOT NULL,
  workcenter VARCHAR(20) DEFAULT NULL,
  operation_description TEXT DEFAULT NULL,
  start_time TIMESTAMP DEFAULT NULL,
  end_time TIMESTAMP DEFAULT NULL,
  setup_time INTERVAL NOT NULL,
  processing_time INTERVAL NOT NULL,
  plan_setup_time INTERVAL NOT NULL,
  plan_processing_time INTERVAL NOT NULL,
  delivery_time INTEGER DEFAULT NULL,
  net_price NUMERIC(10,3) DEFAULT NULL,
  currency VARCHAR(5) DEFAULT NULL,
  unit_price NUMERIC(10,3) DEFAULT NULL
);
CREATE INDEX operations_ix ON operations(order_id, plan, plan_counter, operation, workcenter);

INSERT INTO operations (
  order_id, plan, plan_counter, operation, workcenter, operation_description, start_time, end_time,
  setup_time, processing_time, plan_setup_time, plan_processing_time, delivery_time, net_price, currency, unit_price
) SELECT
  nullif(order_id, ''), nullif(plan, ''), plan_counter, nullif(operation, ''), nullif(workcenter, ''), nullif(operation_description, ''),
  to_timestamp(nullif(start_time, ''), 'mm/dd/yyyy'), to_timestamp(nullif(end_time, ''), 'mm/dd/yyyy'),
  case
    when setup_time_unit = '' then '0 seconds'::INTERVAL
    when setup_time = '' then '0 seconds':: INTERVAL
    when setup_time_unit = 'H' then (regexp_replace(setup_time, ',', '') || ' hours'::TEXT)::INTERVAL
    when setup_time_unit = 'MIN' then (regexp_replace(setup_time, ',', '') || ' minutes'::TEXT)::INTERVAL
  else
    '0 seconds'::INTERVAL
  end,
  case
    when processing_time_unit = '' then '0 seconds'::INTERVAL
    when setup_time = '' then '0 seconds':: INTERVAL
    when processing_time_unit = 'H' then (regexp_replace(processing_time, ',', '') || ' hours'::TEXT)::INTERVAL
    when processing_time_unit = 'MIN' then (regexp_replace(processing_time, ',', '') || ' minutes'::TEXT)::INTERVAL
  else
    '0 seconds'::INTERVAL
  end,
  case
    when plan_setup_time_unit = '' then '0 seconds'::INTERVAL
    when plan_setup_time_unit = 'H' then (plan_setup_time::TEXT || ' hours'::TEXT)::INTERVAL
    when plan_setup_time_unit = 'MIN' then (plan_setup_time::TEXT || ' minutes'::TEXT)::INTERVAL
  else
    '0 seconds'::INTERVAL
  end,
  case
    when plan_processing_time_unit = '' then '0 seconds'::INTERVAL
    when plan_processing_time_unit = 'H' then (plan_processing_time::TEXT || ' hours'::TEXT)::INTERVAL
    when plan_processing_time_unit = 'MIN' then (plan_processing_time::TEXT || ' minutes'::TEXT)::INTERVAL
  else
    '0 seconds'::INTERVAL
  end, nullif(delivery_time, '')::INTEGER, regexp_replace(nullif(net_price, ''), ',', '')::NUMERIC, nullif(currency, ''), regexp_replace(nullif(unit_price, ''), ',', '')::NUMERIC
FROM operations_tmp;

DROP TABLE IF EXISTS operations_clean;
CREATE TABLE operations_clean (
  order_id VARCHAR(10) NOT NULL,
  plan VARCHAR(10) DEFAULT NULL,
  plan_counter VARCHAR(2),
  operation INTEGER NOT NULL,
  prev_operation INTEGER DEFAULT NULL,
  workcenter VARCHAR(5) NOT NULL,
  prev_workcenter VARCHAR(5) DEFAULT NULL,
  operation_description TEXT DEFAULT NULL,
  start_time TIMESTAMP NOT NULL,
  end_time TIMESTAMP  NOT NULL,
  setup_time INTERVAL NOT NULL,
  setup_time_seconds INTEGER NOT NULL,
  processing_time INTERVAL NOT NULL,
  processing_time_seconds INTEGER NOT NULL,
  plan_setup_time INTERVAL NOT NULL,
  plan_setup_time_seconds INTEGER NOT NULL,
  plan_processing_time INTERVAL NOT NULL,
  plan_processing_time_seconds INTEGER NOT NULL,
  setup_time_delay INTEGER NOT NULL,
  processing_time_delay INTEGER NOT NULL,
  delivery_time INTEGER DEFAULT NULL,
  net_price NUMERIC(10,3) DEFAULT NULL,
  currency VARCHAR(5) DEFAULT NULL,
  unit_price NUMERIC(10,3) DEFAULT NULL
);
SELECT create_hypertable('operations_clean', 'start_time');
INSERT INTO operations_clean (
  order_id, plan, plan_counter, operation, workcenter, operation_description, start_time, end_time,
  setup_time, setup_time_seconds, processing_time, processing_time_seconds,
  plan_setup_time, plan_setup_time_seconds, plan_processing_time, plan_processing_time_seconds,
  setup_time_delay, processing_time_delay, delivery_time, net_price, currency, unit_price
) SELECT order_id, plan, plan_counter, operation::INTEGER, workcenter, operation_description, start_time, end_time,
         setup_time, extract(epoch from setup_time), processing_time, extract(epoch from processing_time),
         plan_setup_time, extract(epoch from plan_setup_time), plan_processing_time, extract(epoch from plan_processing_time),
         extract(epoch from setup_time - plan_setup_time), extract(epoch from processing_time - plan_processing_time),
         delivery_time, net_price, currency, unit_price
FROM operations
WHERE start_time IS NOT NULL
  AND end_time IS NOT NULL
  AND setup_time > '0 seconds'::interval
  AND processing_time > '0 seconds'::interval
  AND plan_setup_time > '0 seconds'::interval
  AND plan_processing_time > '0 seconds'::interval;

CREATE INDEX operations_clean_order_id_ix ON operations_clean (order_id);
CREATE INDEX operations_clean_operation_ix ON operations_clean (operation DESC);
CREATE INDEX operations_clean_order_id_operation_ix ON operations_clean (order_id, operation DESC);

UPDATE operations_clean op2 SET prev_operation = op1.operation, prev_workcenter = op1.workcenter
FROM operations_clean op1
WHERE op2.order_id = op1.order_id
  AND op1.operation IN (
    SELECT MAX(operation)
    FROM operations_clean op0
    WHERE op0.order_id = op2.order_id
      AND op0.operation < op2.operation
    );

DROP TABLE IF EXISTS operations_tmp;





