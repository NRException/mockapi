# MockAPI

A simple API service written in go, this API service is designed to get simple API's up and running without any external dependencies, such as runtimes or libraries.

## Description

MockAPI is intended to be simple to use and configure as it uses YAML as a configuration language. It currently boasts the following core features:

* ✅ Multiple web listeners, capable of listening on different ports similtaniously, either on HTTP or HTTPs
* ✅ Multiple content bindings, which are attached to a listener definition, which return headers, response codes, body data and datatypes independently of one-another.
* ✅ Inline content delivery such as simple text, whether this be HTML, JSON, CSV etc, it doesn't matter as it's treated as a simple string.
* ✅ File based content delivery from simple text based formats (currently; .html, .json, .xml, .txt and .csv data formats are supported).
* ✅ KVP based header support in server responses. Return whatever you want in your headers!

MockAPI also limits the amount of third party golang libaries used, this is intended to keep the contributors(s) to the codebase from extending the feature-set beyond the intended scope of this project, in a simple manner of speaking "to keep it simple, stupid". This also has the added benefit of limiting potential supply chain attacks.

As this program is written in go, it is platform agnostic. It is freely compabile with both BSD, linux and windows.

## Getting Started

### Dependencies

To compile:

* Golang 1.21.3 - This is the current build target.
* make - To run make.

### Installing

The make file will build this binary in portable format, you can run it from where-ever you like! Simply:

* Clone the Repo:

```bash
git clone https://github.com/NRException/MockAPI
cd MockAPI
```
* And build:
```bash
make build
```

* Finally, copy the relevent binary from the build/bsd,linux or windows folder to wherever you like!

### Executing program

* Help prompt:
```bash
./mockapi -h
```

* Run MockAPI from a single configuration file (non-verbose):
```bash
./mockapi -f <inputfile>
```

* Run MockAPI from a single configuration file (verbose):
```bash
./mockapi -f <inputfile> -v
```

## Help

<todo> </todo>

## Authors

* [Cameron Huggett](https://github.com/NRException)
* [Eric Wohltman](https://github.com/ewohltman)

## License

This project is licensed under the GNU GENERAL PUBLIC license - see the LICENSE.md file for details

## Acknowledgments

Inspiration, code snippets, etc.
* [simple-readme](https://gist.github.com/DomPizzie/7a5ff55ffa9081f2de27c315f5018afc#file-readme-template-md) - For the readme you're currently reading
* [go-yaml](https://github.com/go-yaml/yaml) - For yaml support
* [go uuid](https://github.com/google/uuid) - For UUID generation