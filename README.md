# ltst

## Installation
`go get github.com/marcelpuyat/ltst`

## Description
View the latest news of your choosing right in your terminal.

Simply define a [config file](/examples/ltst-config-example.yaml) under `$HOME/.ltst.yaml` to pull markup (using [goquery](https://github.com/PuerkitoBio/goquery) syntax) from URLs of your choosing.

For example, to pull the latest article titles from [The Morning Paper](http://blog.acolyer.org/) blog, have an entry as such in your config file:
```json
-
  name: "The Morning Paper"
  command: "morningpaper"
  query: "[rel='bookmark']"
  url: http://blog.acolyer.org/
```

Then, when you type in `ltst`, you'll see the first entry for your provided search:
```
The Morning Paper
	Chapar: Certified Causally Consistent Distributed Key-Value Stores
```

All entries are fetched in parallel, making for a blazingly fast and simple way of seeing what's up with your favorite sites!

Use `ltst --help` to see other flags and features which are automatically generated as you add to your config file.