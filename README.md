Jet is a super-flexible lightweight SQL interface ([GoDoc](http://godoc.org/github.com/eaigner/jet))

Features:

  - LRU query cache for superfast queries
  - Unpack query results to any format
  - Can expand queries for map and slice arguments (e.g. `$1` to `$1, $2, $3`, useful for hstore or set membership queries)
  - Serializes hstore columns to maps
  - Simple migration API
  - Customizable column name mapper

### Open

    db, err := jet.Open("postgres", "...")

### Insert Rows

    db.Query(`INSERT INTO "fruits" ( "name", "price" ) VALUES ( $1, $2 )`, "banana", 2.99).Run()

Run is Jet's `Exec` equivalent and is used instead of `Rows()` when no return values are expected

### Query Rows

    var rows []*struct{
      Name  string
      Price int
    }
    db.Query(`SELECT * FROM "fruits"`).Rows(&rows)

Jet's column mapper is very powerful. It tries to map the columns to any value you provide. You're not required to use a fixed output format. In this case `rows` could be anything e.g `struct`, `*struct`, `[]struct`, `[]*struct`, `Type`, `*Type`, `[]Type`, `[]*Type` even `map[string]interface{}` or just simple values like `int` or `*int`. You get the idea.

### Query Value

Jet provides a convenience method if the returned row is e.g. an aggregation result. You can use `Value()` to quickly get the value of the first column and row

    var count int
    db.Query(`SELECT COUNT(*) FROM "fruits"`).Value(&count)

### Hstore

Jet can also deserialize hstore columns for you. In this case the `header` column is a `hstore` value.

    var out struct{
      Header  map[string]interface{}
      Body    string
    }
    db.Query(`SELECT * FROM "emails"`).Rows(&out)

### Map and Slice Expansion

If you want to do e.g. hstore inserts or set membership queries, Jet can automatically expand the query and adjust the argument list for you.

Passing in a **map** argument

    db.Query(`INSERT INTO "emails" ( "header", "body" ) VALUES ( $1, $2 )`, aMap, aBody).Run()

will expanded the query to

    INSERT INTO "emails" ( "header", "body" ) VALUES ( hstore(ARRAY[ $1, $2, $3, $4 ... ]), $5 )

Passing in a **slice** argument

    db.Query(`SELECT * FROM "files" WHERE "files"."name" IN ( $1 )`, aSlice)

will expand the query to

    SELECT * FROM "files" WHERE "files"."name" IN ( $1, $2, $3, ... )
