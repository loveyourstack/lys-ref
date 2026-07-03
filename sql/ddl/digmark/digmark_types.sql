
CREATE TYPE digmark.launcher_status AS ENUM ('Unprepared', 'Preparing', 'Invalid', 'Ready', 'Queued', 'Processing', 'Failed', 'Completed');

CREATE TYPE digmark.manager AS ENUM ('All', 'Anna', 'Michael', 'Paul');

CREATE TYPE digmark.partner AS ENUM ('Facebook', 'Google Ads');
