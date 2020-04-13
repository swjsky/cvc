config: Go configuration
========================

Go configuration from configuration file and then command line arguments.

Example
-------

```go
type Config struct {
    // ...
}

func main() {
	var cfg Config
	if err := config.Parse(&cfg); err != nil {
		if _, ok := err.(*config.HelpError); ok {
			fmt.Println(err)
			return
		}
		log.Fatal(err)
	}
    // ...
}
```