-- run as superuser
-- in production: use real passwords

CREATE USER lysref_owner WITH PASSWORD '123' CREATEROLE;
CREATE USER lysref_server WITH PASSWORD '456'; -- next project: "_srv" or "_http", not "_server"
CREATE USER lysref_cli WITH PASSWORD '789';
CREATE USER lysref_lis WITH PASSWORD '567';
CREATE USER lysref_mcp WITH PASSWORD 'abc';
CREATE USER lysref_rpc WITH PASSWORD 'def';


-- login for suppliers
CREATE USER lysref_supplier WITH PASSWORD '456';
