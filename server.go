package main

import (
       "github.com/go-martini/martini"
       "fmt"
       "net/http"
       "io/ioutil"
       "encoding/json"
       "time"
       "strconv"
       "github.com/garyburd/redigo/redis"
       "math"
       "os"
)

func perror(err error) {
    defer func() {
        if r := recover(); r != nil {
            fmt.Println("Recovered in perror", r)
        }
    }()
    if err != nil {
        panic(err)
    }
}

type Story_info struct {
   By       string      `json:"by"`
   Id       uint32      `json:"id"`
   Kids     []uint32    `json:"kids"`
   Score    uint32      `json:"score"`
   Text     string      `json:"text"`
   Time     int64      `json:"time"`
   Title    string      `json:"title"`
   Stype    string      `json:"type"`
   Url      string      `json:"url"`
}

type Comment_info struct {
   By       string     `json:"by"`
   Id       uint32     `json:"id"`
   Kids     []uint32   `json:"kids"`
   Parent   uint32     `json:"parent"`
   Text     string     `json:"text"`
   Time     int64     `json:"time"`
   Type     string     `json:"type"`
}

type algolia_item struct {
  Id           uint32          `json:"id"`
  Created_at   string          `json:"created_at"`
  Created_at_i int64          `json:"created_at_i"`
  Type         string          `json:"type"`
  Authur       string          `json:"author"`
  Title        string          `json:"title"`
  Url          string          `json:"url"`
  Text         string          `json:"text"`
  Points       uint32          `json:"points"`
  Children     []algolia_item  `json:"children"`
  Parent_id    uint32          `json:"parent_id"`
  Story_id     uint32          `json:"story_id"`
}


func get_hackernews_topstories() (string, string) {
  url := "https://hacker-news.firebaseio.com/v0/topstories.json"
  start := time.Now()

  res, err := http.Get(url)
  perror(err)
  defer res.Body.Close()
  body, err := ioutil.ReadAll(res.Body);
  length := len(body)
  perror(err)

  // bodyStr contains the list of top 100 hacknews stories
  bodyStr := string(body[:length])

  returnstring := ""
  
  // Go through the string of bodyStr and create an array of the top story IDs
  var story []string
  temp := ""
  for i:=1; i< length-1; i++ {
    if(bodyStr[i]==',') {
      story = append(story, temp)
      temp = ""
    } else {
      temp += string(bodyStr[i])
    }
  }
  story = append(story, temp)
  time.Sleep(15 * time.Millisecond)

  // Create a JSON stories file
  // Open File
  fo, err := os.Create("public/stories.json")
  if err != nil {
      panic(err)
  }
  // close fo on exit and check for its returned error
  defer func() {
      if err := fo.Close(); err != nil {
          panic(err)
      }
  }()

  var storynames map[string]string
  storynames = make(map[string]string)

  n3, err := fo.WriteString("{\n\t\"stories\":{\n")
  fmt.Printf("wrote %d bytes\n", n3)
  fo.Sync()


  // Go through the array of top story Ids and pull out their data from the HN API
  strLength := fmt.Sprintf("%d", len(story))
  for i:=0; i<len(story); i++ { // for the whole thing
  //for i:=0; i<50; i++ { // for just the top 30
    temp := story[i]
    currentIndex := fmt.Sprintf("%d", i+1)
    fmt.Println(currentIndex + " of " + strLength + ": " + string(temp))

    algoliaitem := &algolia_item{}
    u,_ := strconv.ParseUint(temp,0,32)
    algoliaitem = get_algolia_item(u)
    time.Sleep(15 * time.Millisecond)

    if(algoliaitem.Id == 0) {
      storyinfo := &Story_info{}
      storyinfo = getStoryData(u);
      tempname := fmt.Sprintf("%d", storyinfo.Id)
      storynames[tempname] = storyinfo.Title
      t := fmt.Sprint(storyinfo.Id)
      storyScore := fmt.Sprint(storyinfo.Score);
      tm := time.Unix(storyinfo.Time, 0)
      timesince := time.Since(tm)
      roundtm := Round(timesince.Hours(),.5,0)

      returnstring += "<tr><td colspan=\"8\"><div id=\"story_" + t + "\"></div></td></tr></tr><tr><td>" + fmt.Sprintf("%d", i+1) + "</td><td>" + fmt.Sprintf("%s",tm) + "</td><td>" + fmt.Sprintf("%d",int(roundtm)) + "</td><td>" + t + "</td><td>" + string(storyinfo.Stype) + "</td><td>"  + string(storyinfo.Title) + "</td><td>0</td><td>" + storyScore + "</td></tr>"
      time.Sleep(15 * time.Millisecond)
    } else {
      tempname := fmt.Sprintf("%d", algoliaitem.Id)
      storynames[tempname] = algoliaitem.Title

      t := fmt.Sprint(algoliaitem.Id)
       tm := time.Unix(algoliaitem.Created_at_i, 0)
      timesincehours := time.Since(tm).Hours()
      roundtm := Round(timesincehours,.5,0)
      timearr := make([]int,int(roundtm)+1) // plus 2hours, 1 hour for the index set at 0
      numofComments := fmt.Sprint(count_algolia_items(algoliaitem, timearr, tm))
        lastComma := ""
        if(i!=len(story)-1) {
          lastComma = ","
        }

      _, err = fo.WriteString("\t\t\"story_"+string(temp)+"\": [\n")
      for i,elem := range timearr {
        lastComma := ""
        if(i!=len(timearr)-1) {
          lastComma = ","
        }
        index := fmt.Sprintf("%d", i)
        amount := fmt.Sprintf("%d", elem)
        _, err = fo.WriteString("\t\t\t{\n" +
                                "\t\t\t\"hour\":" + index + ",\n" +
                                "\t\t\t\"comments\":" + amount  + "\n" +
                                "\t\t\t}" + lastComma + "\n")
      }
      _, err = fo.WriteString("\t\t]" + lastComma + "\n")
      fo.Sync()

      storyScore := fmt.Sprint(algoliaitem.Points);

      // time
      //timesince := time.Since(tm)

      returnstring += "<tr><td colspan=\"8\"><div id=\"story_" + t + "\"></div></td></tr></tr><tr><td>" + fmt.Sprintf("%d",i+1) + "</td><td>" + fmt.Sprintf("%s",tm) + "</td><td>" + fmt.Sprintf("%d",int(roundtm)) + "</td><td>" + t + "</td><td>" + string(algoliaitem.Type) + "</td><td><a href='" + string(algoliaitem.Url) + "'>"  + string(algoliaitem.Title) + "</a></td><td>" + numofComments + "</td><td>" + storyScore + "</td></tr>"
    }
  }
  _, err = fo.WriteString("\t},\n" +
                          "\t\"storynames\":{\n")

  for v,k := range storynames {
      _, err = fo.WriteString("\t\t\""+v+"\":\""+k+"\",\n")
  }
  _, err = fo.WriteString("\t\t\"\":\"\"\n")

  n4, err := fo.WriteString("\t}\n}\n")
  fmt.Printf("wrote %d bytes\n", n4)

  fo.Sync()

  //fo.Close()

  elapsed := time.Since(start)
  fmt.Printf("Time: %s\n", elapsed);

  returntime := fmt.Sprintf("%s", elapsed)

  return returnstring, returntime
}


func get_algolia_item(storyId uint64) (*algolia_item) {
  defer func() {
      if r := recover(); r != nil {
          fmt.Println("Recovered in get_algolia_item", r)
      }
  }()
  currentStory := fmt.Sprint(storyId)
  algolia_url := "https://hn.algolia.com/api/v1/items/"+currentStory
  //story_url := "https://hacker-news.firebaseio.com/v0/item/"+currentStory+".json"

  res, err := http.Get(algolia_url)
  perror(err)
  defer res.Body.Close()
  body, err := ioutil.ReadAll(res.Body);
  perror(err)
  
  algoliaitem := &algolia_item{}
  json.Unmarshal(body, &algoliaitem)

  return algoliaitem
}

func count_algolia_items(algoliaitem *algolia_item, tm []int, oritime time.Time) (int) {
    if (len(algoliaitem.Children) > 0) {
      total := 0
      for _,elem := range algoliaitem.Children {
        //roundtm := Round(timesincehours,.5,0)
        if(elem.Text != "") { // a 0 id means a deleted comment
          elemtm := time.Unix(elem.Created_at_i, 0)

          dur := elemtm.Sub(oritime)
          roundtm := int(Round(dur.Hours(),.5,0))

          if (roundtm >= 0) && roundtm <= len(tm) {
            tm[roundtm]++
          } else {
            fmt.Printf("%v\n", time.Since(oritime).Hours());
            fmt.Printf("%v\n", elem.Created_at_i)
            fmt.Printf("%v\n", time.Since(elemtm).Hours());
            fmt.Printf("Weird Number %v %v %v\n", tm, roundtm, elem) 
          }
          total += 1 + count_algolia_items(&elem, tm, oritime)
        }
      }
      return total
    } else {
      return 0
    }
}


func count_algolia_items_inner(items []algolia_item) (int) {
  if (len(items) == 0) {
    return 1
  } else {
    total := 0
    for _,elem := range items {
      total += count_algolia_items_inner(elem.Children)
    }
    return total
  }
}

func getStoryData(storyId uint64) (*Story_info) {
  defer func() {
      if r := recover(); r != nil {
          fmt.Println("Recovered in GetStoryData", r)
      }
  }()
  currentStory := fmt.Sprint(storyId)
  story_url := "https://hacker-news.firebaseio.com/v0/item/"+currentStory+".json"

  res, err := http.Get(story_url)
  perror(err)
  defer res.Body.Close()
  body, err := ioutil.ReadAll(res.Body);
  perror(err)
  
  storydata := &Story_info{}
  json.Unmarshal(body, &storydata)

  return storydata
}

func getCommentData(commentId uint64) (*Comment_info) {
  defer func() {
      if r := recover(); r != nil {
          fmt.Println("Recovered in GetCommentData", r)
      }
  }()
  currentComment := fmt.Sprint(commentId)
  comment_url := "https://hacker-news.firebaseio.com/v0/item/"+currentComment+".json"

  res, err := http.Get(comment_url)
  perror(err)
  defer res.Body.Close()
  body, err := ioutil.ReadAll(res.Body);
  perror(err)
  
  commentdata := &Comment_info{}
  json.Unmarshal(body, &commentdata)

  return commentdata
}

// count Comments recursively

func countComments(commentId uint64) (int) {
  id := fmt.Sprint(commentId)
  fmt.Println("counting comments of " + id);
  currentComment := &Comment_info{}
  currentComment = getCommentData(uint64(commentId))
  time.Sleep(10 * time.Millisecond)

  total := len(currentComment.Kids)
  culTotal := 0
  if total!=0 {
    culTotal = countCommentsInner(currentComment.Kids)
  }

  return total + culTotal
}

func countCommentsInner(kids []uint32) (int) {
  if (len(kids) == 0) {
    return 1
  } else {
    total := 0
    for _,elem := range kids {
      commentdata := &Comment_info{}
      commentdata = getCommentData(uint64(elem))
      time.Sleep(15 * time.Millisecond)
      total += countCommentsInner(commentdata.Kids)
    }
    return total
  }
}

func main() {
  hackerStr, time := get_hackernews_topstories()
  m := martini.Classic()
  m.Use(martini.Static("build"))
  m.Use(martini.Static("css"))

  c, err := redis.Dial("tcp", ":6379")
  if err != nil {
    panic(err);
  }
  defer c.Close()

  n, err := c.Do("APPEND", "hackernews", "")
  o, err := c.Do("SET", "hackernews", hackerStr)

  p, err := redis.Int(n, err)
  p2, err := redis.Int(o, err)

  fmt.Println(p)
  fmt.Println(p2)

  timePrint := fmt.Sprint(time)
  m.Get("/", func() string {
    return "<!DOCTYPE html>" +
           "<html>" +
           "<head>" +
           // Other
           "<link href='https://fonts.googleapis.com/css?family=Open+Sans:400,300,700' rel='stylesheet' type='text/css'>" +
           "<link href='https://fonts.googleapis.com/css?family=PT+Serif:400,700,400italic' rel='stylesheet' type='text/css'>" +
           "<link href='https://netdna.bootstrapcdn.com/font-awesome/4.2.0/css/font-awesome.css' rel='stylesheet' type='text/css'>" +
           // Jquery
           "<script src='https://ajax.googleapis.com/ajax/libs/jquery/1.11.1/jquery.min.js'></script>" +
           // Bootstrap 3
           "<link rel='stylesheet' href='https://maxcdn.bootstrapcdn.com/bootstrap/3.3.1/css/bootstrap.min.css'>" +
           "<link rel='stylesheet' href='https://maxcdn.bootstrapcdn.com/bootstrap/3.3.1/css/bootstrap-theme.min.css'>" +
           "<script src='https://maxcdn.bootstrapcdn.com/bootstrap/3.3.1/js/bootstrap.min.js'></script>" +
           // ReactJs
           "<script src='react-0.12.2.min.js'></script>" +
           "<script src='JSXTransformer-0.12.2.js'></script>" +
           // CSS
           "<link href='metricsgraphics.css' rel='stylesheet' type='text/css'>" +
           "<link href='metricsgraphics-demo.css' rel='stylesheet' type='text/css'>" +
           "<link href='highlightjs-default.css' rel='stylesheet' type='text/css'>" +

           // Jquery and D3 and other JS
           "<script src='highlight.pack.js'></script>" +
           "<script src='https://cdnjs.cloudflare.com/ajax/libs/d3/3.5.0/d3.min.js' charset='utf-8'></script>" +
           "<script src='metricsgraphics.min.js'></script>" +

           // Body
           "</head></body>" +
           "<div><h3>Build Time: " + timePrint + "</h3></div>" +
           "<div>" +
           "<table class='table'>" +
           "<tr>" +
            "<th></th>" +
            "<th>Date Created</th>" +
            "<th>Hours Since Created</th>" +
            "<th>ID</th>" +
            "<th>Type</th>" +
            "<th>Title</th>" +
            "<th>Comment Count</th>" +
            "<th>Score</th>" +
           "</tr>" +
           hackerStr +
           "</table>" +
           "</div>" +
           "<div id='ufo-sightings'></div>" +
          "<script>" +
          "hljs.initHighlightingOnLoad();" +
          "d3.json('stories.json', function(data) {" +
            "console.log(data.stories);" +
            "var stories = Object.keys(data.stories);" +
            "stories.forEach(function(i) {" +
              "var target = '#'+i;" +
              "console.log(i);" +
              "MG.data_graphic({" +
                  "title: data.storynames[i.substring(6)]," +
                  "data: data.stories[i]," +
                  "width: 650," +
                  "height: 150," +
                  "target: target," +
                  "x_accessor: 'hour'," +
                  "y_accessor: 'comments'" +
              "});" +
            "});" +
          "})" +
          "</script>" +
           "</body></html>"
  })
  m.Run()
}


func Round(val float64, roundOn float64, places int ) (newVal float64) {
  var round float64
  pow := math.Pow(10, float64(places))
  digit := pow * val
  _, div := math.Modf(digit)
  if div >= roundOn {
    round = math.Ceil(digit)
  } else {
    round = math.Floor(digit)
  }
  newVal = round / pow
  return
}
