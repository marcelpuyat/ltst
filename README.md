# ltst
View the latest news of your choosing right in your terminal. All results are fetched in parallel, making for a blazingly fast and simple way of seeing what's up with your favorite sites!

## Installation
`go get github.com/marcelpuyat/ltst`

## Quick Start

Simply define a [config file](/examples/ltst-config-example.yaml) under `$HOME/.ltst.yaml` to pull markup (using [goquery](https://github.com/PuerkitoBio/goquery) syntax) from URLs of your choosing.

For example, to pull the latest article titles from [The Morning Paper](http://blog.acolyer.org/) blog, have an entry as such in your config file:
```json
-
  name: "The Morning Paper"
  command: "morningpaper"
  query: "[rel='bookmark']"
  url: http://blog.acolyer.org/
```

Then, when you type in `ltst`, you'll see the latest entry for your provided search:
```
The Morning Paper
	Chapar: Certified Causally Consistent Distributed Key-Value Stores
```

Use `ltst --help` to see other flags and features which are automatically generated as you add to your config file.

Other example commands:
```
> ltst morningpaper   # Outputs the latest 5 entries for The Morning Paper

The Morning Paper
	Chapar: Certified Causally Consistent Distributed Key-Value Stores
	Is Sound Gradual Typing Dead?
	Reducing Crash Recoverability to Reachability
	‘Cause I’m Strong Enough: Reasoning About Consistency Choices in Distributed Systems
	Modelling the ARM v8 Architecture, Operationally: Concurrency and ISA
```

```
> ltst morningpaper -o   # Opens blog.acolyer.org in your default browser
```