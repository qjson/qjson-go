# QJSON to JSON converter

[JSON](https://www.json.org) is a very popular data encoding with a good support in many 
programming languages. It may thus seam a good idea to use it for manually managed 
data like configuration files. Unfortunately, JSON is not very convenient for such 
application because every string must be quoted and elements must be separated by commas. 
It is easy to parse by a program and to verify that it is correct, but it’s not connvenient
to write. 

For this reason, different alternatives to JSON have been proposed which are more human 
friendly like [YAML] or [HJSON](https://hjson.github.io/) for instance. 

QJSON is inspired by HJSON, and you should be able to parse HJSON text with QJSON. 
It differs from HJSON by extending its functionality and relaxing some rules.

Here is a list of QJSON text properties:

- comments of the form //...  #... or /*...*/
- commas between array values and object members are optional 
- double quote, single quote and quoteless strings
- non-breaking space is considered as a white space character
- newline is \n or \r\n, report \r alone or \n\r as an error
- numbers are decimal, floating point, hexadecimal, octal or binary
- numbers may contain underscore '_' as separator
- numbers may be simple mathematical expression
- member identifiers may be quoteless strings

Current limitation:

- the backspace or formfeed control characters are invalid
- multiline strings are not yet supported


## Usage 

You need to get the QJSON package if the go tools can’t find it by themselves.

`go got github.com/chmike/QJSON-go`


The API is a single function:

`QJSON.Decode(QJSONText []byte) (jsonText []byte, err error)` 

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

Positive feedbacks like that you are using QJSON are welcome in the issues. 

## Syntax 

**This package supports QJSON syntax v0.0.0"**

THe QJSON syntax is described in the 
[qjson-syntax project](http://github.com/qjson/qjson-syntax).

Take care to read the syntax specification supported
by this package as the syntax may evolve and be ahead.

## Reliability

qjson-go has been extensively tested with manualy defined tests (100% coverage), 
and go-fuzz running for longer than 10 hours and continuing. The number of bugs
left should be very small. 

For bug reports in this QJSON-go package, fill an issue in this project. 

The things you may still find are discordance between the syntax specification 
and the go implementation, or unclear of confusing error messages. 

## Contributing

QJSON is a recently developped package. It is thus a good time to 
suggest changes or extensions to the syntax since the user base is very
small. For syntax modification requests or problems fill an issue in the 
[QJSON-syntax project](http://github.com/chmike/QJSON-syntax).

Any contribution is welcome. 

## License

The licences is the 3-BSD clause. 
