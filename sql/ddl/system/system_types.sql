
CREATE TYPE system.http_method AS ENUM ('GET', 'HEAD', 'POST', 'PUT', 'PATCH', 'DELETE');

CREATE TYPE system.notification_type AS ENUM ('Info', 'Warning');

CREATE TYPE system.role AS ENUM ('Standard', 'Tech', 'Viewer');