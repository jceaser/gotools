package main

/*
A quick and dirty program to take an XML RSS feed and convert it to Json.
Alternatively one of the fields in the Item Object can be returned instead of
the entire JSON.
*/

// GOOS=linux GOARCH=amd64 go build rss2json.go
// curl -s http://rss.cnn.com/rss/edition.rss | go run rss2json.go -field title

import (
	"encoding/json"
	"encoding/xml"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"reflect"
	"strings"
)

/******************************************************************************/
// MARK: - Structures

type RSS struct {
	XMLName xml.Name `xml:"rss"`
	Channel Channel  `xml:"channel"`
}

type Channel struct {
	Title       string `xml:"title"`
	Description string `xml:"description"`
	Link        string `xml:"link"`
	Items       []Item `xml:"item"`
}

type Item struct {
	Title       string `xml:"title"`
	Link        string `xml:"link"`
	Description string `xml:"description"`
	PubDate     string `xml:"pubDate"`
	GUID        string `xml:"guid"`
}

func (self Item) Field(field_name string) string {
    field_name = strings.Title(field_name)
    raw_value := reflect.ValueOf(self).FieldByName(field_name)
    if !raw_value.IsValid() {
        fmt.Fprintf(os.Stderr, "Error finding field [%s].\n", field_name)
        os.Exit(1)
    }
    return raw_value.String()
}

// Return a field from a file by specifying it by name
func (self Item) Field2(field string) string {
    ret := ""
    switch field {
        case "title", "Title":
            ret = self.Title
        case "description", "Description":
            ret = self.Description
        case "pubdate", "PubDate":
            ret = self.PubDate
        case "guid", "GUID":
            ret = self.GUID
        default:
            ret = "Unknown"
    }
    return ret
}

/******************************************************************************/
// MARK: - Generics

// Convert a raw XML document in a byte array to a structure
func XmlToStruct[Target any](bytes []byte) (Target, error) {
	var data Target
	err := xml.Unmarshal(bytes, &data)
	if err != nil {
		return data, err
	}
	return data, nil
}

// Convert a raw Json document in a byte array to a structure
func JsonToStruct[Target any](bytes []byte) (Target, error) {
	var data Target
	err := json.Unmarshal(bytes, &data)
	if err != nil {
		return data, err
	}
	return data, nil
}

// Convert a structure to a raw Json document in a byte array
func StructToJson[T any](data T, useIndent bool) ([]byte, error) {
	var bytes []byte
	var err error
	if useIndent {
		bytes, err = json.MarshalIndent(data, "", strings.Repeat(" ", 4))
	} else {
		bytes, err = json.Marshal(data)
	}
	if err != nil {
		return nil, err
	}
	return bytes, nil
}

/******************************************************************************/
// MARK: - Functions

func main() {

    var file, field string
	flag.StringVar(&file, "file", "-", "Path to the input file, or stdin.")
	flag.StringVar(&field, "field", "", "Field to output.")
	flag.Parse()

    // *********************************
    // Read input

    var data []byte
    var err error
    if file == "-"{
        // Read RSS XML from stream
        data, err = io.ReadAll(os.Stdin)
        if err != nil {
            fmt.Fprintf(os.Stderr, "Error reading stdin: %v\n", err)
            os.Exit(1)
        }
    } else {
        // Read RSS XML file
        data, err = ioutil.ReadFile(file)
        if err != nil {
            fmt.Fprintf(os.Stderr, "Error reading file: [%s] %v\n", file, err)
            os.Exit(1)
        }
    }

    // *********************************
	// Parse XML

	var rss RSS
	rss, err = XmlToStruct[RSS](data)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error parsing XML: %v\n", err)
		os.Exit(1)
	}

    // *********************************
	// Convert to JSON

    jsonData, err := StructToJson[RSS](rss, true)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error converting to JSON: %v\n", err)
		os.Exit(1)
	}

    // *********************************
	// Print JSON

	if field == "" {
    	fmt.Println(string(jsonData))
	} else {
	    for _, item := range rss.Channel.Items {
	        fmt.Printf("%s\n", item.Field(field))
	    }
	}
}
