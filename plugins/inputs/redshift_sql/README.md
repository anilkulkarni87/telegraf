# Redshift_SQL plugin

This redshift plugin enables you too publish metrics for your redshift database. It has been
designed to parse the sql queries in the plugin section of your telegraf.conf.

For now only two queries are specified and it's up to you to add more; some per
query parameters have been added :

* The SQl query itself
* The list of the column that have to be defined has tags

```
[[inputs.redshift_sql]]
  # specify address via a url matching:
  # postgres://[pqgotest[:password]]@localhost[/dbname]?sslmode=...
  #
  # All connection parameters are optional.  #
  # Without the dbname parameter, the driver will default to a database
  # with the same name as the user. This dbname is just for instantiating a
  # connection with the server and doesn't restrict the databases we are trying
  # to grab metrics for.
  #
  address = "postgres://user:pwd@host:port/dbname?sslmode=disable"
  #
  #
  # Structure :
  # [[inputs.redshift_sql.query]]
  #   sqlquery string
  #   tagvalue string (coma separated)
  [[inputs.redshift_sql.query]]
    sqlquery="SELECT * FROM tablename where datname"
    tagvalue=""
  [[inputs.redshift_sql.query]]
    sqlquery="SELECT * FROM tablename"
    tagvalue=""
```





 
