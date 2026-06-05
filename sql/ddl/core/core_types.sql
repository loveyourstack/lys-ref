
CREATE TYPE core.mandatory_enum AS ENUM ('A', 'B');
CREATE TYPE core.optional_enum AS ENUM ('', 'A', 'B');

CREATE TYPE core.performance_period AS ENUM (
  'Today',
  'Yesterday',
  'Last 3 days',
  'Last 7 days',
  'Last 14 days',
  'Last 30 days',
  'This month',
  'Last month',
  'Last 90 days'
);