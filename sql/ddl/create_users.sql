-- run as superuser
-- production: use real passwords

CREATE USER lysref_owner WITH PASSWORD '123' CREATEROLE;
CREATE USER lysref_server WITH PASSWORD '456';
CREATE USER lysref_cli WITH PASSWORD '789';

-- login for suppliers
CREATE USER lysref_supplier WITH PASSWORD '456';