# Commander

Commander makes it easier to build command-line tools in Go.

Commander is tiny library on top of the excellent [kingpin library](https://github.com/alecthomas/kingpin).

## Install

```sh
go get -u github.com/matthewmueller/commander
```

## Usage

Here's a real-world example of using Commander to build a migration tool:

```go
func main() {
	log := log.Log
	migrate := commander.New("migrate", "Postgres migration CLI")

	{ // create a new migration
		new := migrate.Command("new", "create a new migration")
		name := new.Arg("name", "create a new migration by name").Required().String()
		dir := new.Flag("dir", "migrations directory").Default("./migrations").String()
		new.Run(func() error {
			return migrator.New(*dir, *name)
		})
	}

	{ // migrate up
		up := migrate.Command("up", "migrate up")
		db := up.Flag("db", "database url (e.g. postgres://localhost:5432)").Required().String()
		name := up.Arg("name", "name of the migration to migrate up to").String()
		dir := up.Flag("dir", "migrations directory").Default("./migrations").String()
		up.Run(func() error {
			conn, err := connect(*db)
			if err != nil {
				return err
			}
			defer conn.Close()
			var n string
			if name != nil {
				n = *name
			}
			return migrator.UpTo(conn, *dir, n)
		})
	}

	{ // migrate down
		down := migrate.Command("down", "migrate down")
		db := down.Flag("db", "database url (e.g. postgres://localhost:5432)").Required().String()
		name := down.Arg("name", "name of the migration to migrate down to").String()
		dir := down.Flag("dir", "migrations directory").Default("./migrations").String()
		down.Run(func() error {
			conn, err := connect(*db)
			if err != nil {
				return err
			}
			defer conn.Close()
			var n string
			if name != nil {
				n = *name
			}
			return migrator.DownTo(conn, *dir, n)
		})
	}

	{ // get info on the current migration
		info := migrate.Command("info", "get the current migration number")
		db := info.Flag("db", "database url (e.g. postgres://localhost:5432)").Required().String()
		info.Run(func() error {
			conn, err := connect(*db)
			if err != nil {
				return err
			}
			defer conn.Close()
			v, err := migrate.Version(conn)
			if err != nil {
				return err
			}
			log.Infof("currently at: %d", v)
			return nil
		})
	}

	cli.MustParse(os.Args[1:])
}
```

## Authors

- Matt Mueller [https://twitter.com/mattmueller](https://twitter.com/mattmueller)

## License

MIT
