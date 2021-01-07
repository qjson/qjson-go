# qjson to json converter

[json](https://www.json.org) is a very popular data encoding with a good support in many 
programming languages. It may thus seam a good idea to use it for manually managed 
data like configuration files. Unfortunately, json is not very convenient for such 
application because every string must be quoted and elements must be separated by commas. 
It is easy to parse by a program and to verify that it is correct, but it’s not connvenient
to write. 

For this reason, different alternatives to json have been proposed which are more human 
friendly like [yaml](https://yaml.org/), [toml](https://toml.io/en/) or 
[hjson](https://hjson.github.io/) for instance. 

qjson is inspired by hjson by being a human readable and extended json. The difference 
between qjson and hjson is that qjson extends its functionality and relax some rules.

Here is a list of qjson text properties:

- comments of the form //...  #... or /*...*/
- commas between array values and object members are optional 
- double quote, single quote and quoteless strings
- non-breaking space is considered as a white space character
- newline is \n or \r\n, report \r alone or \n\r as an error
- numbers are integer, floating point, hexadecimal, octal or binary
- numbers may contain underscore '_' as separator
- numbers may be simple mathematical expression with parenthesis
- member identifiers may be quoteless strings including spaces
- the newline type in multiline string is explicitely specified
- backspace and form feed controls are invalid characters except
  in /*...*/ comments or multiline strings
- time durations expressed with w, d, h, m, s suffix are converted to seconds
- time specified in ISO format is converted to UTC time is seconds

## Usage 

If Go can’t find the qjson package by itself, you may get it with
the following command.

`go get github.com/qjson/qjson-go`


The API is a single function that receive the qjson text as input and
returns the corresponding json string or an error message.

`qjson.Decode(qjsonText []byte) (jsonText []byte, err error)` 

Here is an example of usage:

```
package main

import (
    "github.com/qjson/qjson-go"
    "io/ioutil"
    "os"
)


func main() {

    qjsonText, err := ioutil.ReadFile("myFile.qjson")
    if err != nil {
        fmt.Println(err)
        os.Exit(1)
    }

    jsonText, err := qjson.Decode(qjsonText)
    if errr != nil {
        fmt.Println(err)
        os.Exit(1)
    }

    // process json text as usual
    ...
}
```

qjson-go imports only standard packages. There are no 
dependencies with other packages. 

## Syntax 

**This package supports qjson syntax v0.0.0"**

THe qjson syntax is described in the 
[qjson-syntax project](http://github.com/qjson/qjson-syntax).

Check the syntax version supported by the package since
the syntax may evolve. 

## Reliability

qjson-go has been extensively tested with manualy defined tests (100% coverage), 
and go-fuzz running for longer than a day and continuing. The number of bugs
left should be very small. 

For bug reports in this qjson-go package, fill an issue in this project. 

The things you may still find are discordance between the syntax specification 
and the go implementation, or unclear of confusing error messages. 

## Contributing

qjson is a recently developped package. It is thus a good time to 
suggest changes or extensions to the syntax since the user base is very
small. 

For suggestions or problems relative to syntax, fill an issue in the 
[qjson-syntax project](http://github.com/qjson/qjson-syntax).

Any contribution is welcome. 

## License

The licences is the 3-BSD clause. 
