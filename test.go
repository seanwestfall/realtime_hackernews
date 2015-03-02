package main

import "fmt"
import "net/http"
import "io/ioutil"
import "encoding/json"

type Tracks struct {
    Toptracks Toptracks_info
}

type Toptracks_info struct {
    Track []Track_info
    Attr  Attr_info `json: "@attr"`
}

type Track_info struct {
    Name       string
    Duration   string
    Listeners  string
    Mbid       string
    Url        string
    Streamable Streamable_info
    Artist     Artist_info   
    Attr       Track_attr_info `json: "@attr"`
}

type Attr_info struct {
    Country    string
    Page       string
    PerPage    string
    TotalPages string
    Total      string
}

type Streamable_info struct {
    Text      string `json: "#text"`
    Fulltrack string
}

type Artist_info struct {
    Name string
    Mbid string
    Url  string
}

type Track_attr_info struct {
    Rank string
}

func perror(err error) {
    if err != nil {
        panic(err)
    }
}

func get_content() {
    url := "http://ws.audioscrobbler.com/2.0/?method=geo.gettoptracks&api_key=c1572082105bd40d247836b5c1819623&format=json&country=Netherlands"

    res, err := http.Get(url)
    perror(err)
    defer res.Body.Close()
    body, err := ioutil.ReadAll(res.Body)
    perror(err)

    var data Tracks
    err = json.Unmarshal(body, &data)
    if err != nil {
        fmt.Printf("%T\n%s\n%#v\n",err, err, err)
        switch v := err.(type){
            case *json.SyntaxError:
                fmt.Println(string(body[v.Offset-40:v.Offset]))
        }
    }
    for i, track := range data.Toptracks.Track{
        fmt.Printf("%d: %s %s\n", i, track.Artist.Name, track.Name)
    }
}

func main() {
    get_content()
  }
