-- run as superuser
-- in production: use real passwords

CREATE USER lysref_owner WITH PASSWORD '123' CREATEROLE;
CREATE USER lysref_server WITH PASSWORD '456';
CREATE USER lysref_cli WITH PASSWORD '789';
CREATE USER lysref_lis WITH PASSWORD '567';
CREATE USER lysref_mcp WITH PASSWORD 'abc';


-- login for suppliers
CREATE USER lysref_supplier WITH PASSWORD '456';
