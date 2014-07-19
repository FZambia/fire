package main

import (
	"encoding/json"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/user"
	"path"
	"strconv"
	"time"

	"github.com/codegangsta/cli"
	"github.com/wsxiaoys/terminal"
)

// individual entry from subreddit
type Entry struct {
	Title     string
	Author    string
	URL       string
	Permalink string
	Score     int
}

// the feed is the full JSON data structure for subreddit
// this sets up the array of Entry types (defined above)
type Feed struct {
	Data struct {
		Children []struct {
			Data Entry
		}
	}
}

// subreddit representation
type Subreddit struct {
	Name         string
	Score        int
	Entries      []Entry `json:",omitempty"`
	Error        bool    `json:",omitempty"`
	ErrorMessage string  `json:",omitempty"`
}

// return new empty Subreddit instance
func NewSubreddit(name string, score int) *Subreddit {
	return &Subreddit{
		Name:  name,
		Score: score,
	}
}

// build JSON endpoint URL
func (subreddit *Subreddit) GetJsonUrl() string {
	url := "http://www.reddit.com/r/" + subreddit.Name + "/hot.json"
	return url
}

// build URL to browse
func (subreddit *Subreddit) GetUrl() string {
	url := "http://www.reddit.com/r/" + subreddit.Name + "/hot"
	return url
}

// configuration struct
type Configuration struct {
	Subreddits []*Subreddit
}

// fill Configuration based using JSON file
func (c *Configuration) LoadFromFile(fileName string) error {

	if _, err := os.Stat(fileName); os.IsNotExist(err) {
		os.Create(fileName)
	}

	file, err := os.Open(fileName)
	if err != nil {
		return err
	}
	defer file.Close()
	decoder := json.NewDecoder(file)
	err = decoder.Decode(c)
	if err == io.EOF {
		return nil
	}
	return err
}

// dump configuration into JSON file
func (c *Configuration) DumpIntoFile(fileName string) error {
	b, err := json.MarshalIndent(c, "", "    ")
	if err != nil {
		return err
	}
	err = ioutil.WriteFile(fileName, b, 0644)
	return err
}

// return Subreddit from configuration by name
func (c *Configuration) getSubredditByName(name string) *Subreddit {
	for _, subreddit := range c.Subreddits {
		if subreddit.Name == name {
			return subreddit
		}
	}
	return nil
}

// add(or update) Subreddit in configuration
func (c *Configuration) addSubreddit(name string, score int) {
	subreddit := c.getSubredditByName(name)
	if subreddit != nil {
		subreddit.Score = score
	} else {
		c.Subreddits = append(c.Subreddits, NewSubreddit(name, score))
	}
}

// delete Subreddit with given name from configuration
func (c *Configuration) deleteSubredditByName(name string) {
	subreddits := make([]*Subreddit, 0)

	for _, subreddit := range c.Subreddits {
		if subreddit.Name == name {
			continue
		} else {
			subreddits = append(subreddits, subreddit)
		}
	}

	c.Subreddits = subreddits

}

// load JSON from Reddit, fill Subreddit Entries and then send Subreddit into results channel
func fetch(subreddit *Subreddit, results chan *Subreddit) {
	resp, err := http.Get(subreddit.GetJsonUrl())
	if err != nil {
		log.Fatalln("Error fetching:", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		log.Fatalln("Error Status not OK:", resp.StatusCode)
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatalln("Error reading body:", err)
	}

	var entries Feed
	if err := json.Unmarshal(body, &entries); err != nil {
		log.Fatalln("Error decoing JSON", err)
	}

	for _, ed := range entries.Data.Children {
		entry := ed.Data
		if entry.Score >= subreddit.Score {
			subreddit.Entries = append(subreddit.Entries, entry)
		}
	}

	results <- subreddit
}

// pretty print collected data into stdout
func prettyOutput(subreddits []*Subreddit) {
	for _, subreddit := range subreddits {
		terminal.Stdout.Color("y").Print(subreddit.Name).Print(" ").Color("c").Print(subreddit.GetUrl())
		for _, entry := range subreddit.Entries {
			terminal.Stdout.Nl().Color("g").Print(entry.Score, " ").Reset().Print(entry.Title, " ").Color("m").Print(entry.URL)
		}
		terminal.Stdout.Nl().Nl()
	}
	terminal.Stdout.Reset()
}

// print JSON output into stdout
func jsonOutput(subreddits []*Subreddit) {
	enc := json.NewEncoder(os.Stdout)
	enc.Encode(subreddits)
}

// gather filled Subreddits from results channel, return slice with them
func collect(results chan *Subreddit, configuration *Configuration, timeout time.Duration) {
	entries := make([]*Subreddit, 0)
	for {
		select {
		case subreddit := <-results:
			entries = append(entries, subreddit)
			if len(entries) == len(configuration.Subreddits) {
				return
			}
		case <-time.After(timeout * time.Second):
			log.Println("timeout")
			return
		}
	}
	return
}

// fetch data for subreddits based on configuration provided
func load(configuration *Configuration, context *cli.Context) {

	timeout := time.Duration(context.GlobalInt("timeout"))
	jsonOut := context.GlobalBool("json")

	if len(configuration.Subreddits) == 0 {
		log.Fatalln("No subreddits found")
	}

	results := make(chan *Subreddit, len(configuration.Subreddits))

	for _, subreddit := range configuration.Subreddits {
		go fetch(subreddit, results)
	}

	collect(results, configuration, timeout)

	if jsonOut {
		jsonOutput(configuration.Subreddits)
	} else {
		prettyOutput(configuration.Subreddits)
	}

}

func main() {

	usr, err := user.Current()
	if err != nil {
		log.Fatalln(err)
	}

	app := cli.NewApp()
	app.Name = "fire"
	app.Usage = "show posts from your favorite Reddit subreddits filtered by score"

	configFlag := cli.StringFlag{Name: "config, c", Value: path.Join(usr.HomeDir, ".fire.json"), Usage: "path to JSON configuration file"}
	timeoutFlag := cli.IntFlag{Name: "timeout, t", Value: 3, Usage: "timeout"}
	jsonFlag := cli.BoolFlag{Name: "json, j", Usage: "JSON output"}

	app.Flags = []cli.Flag{
		configFlag,
		timeoutFlag,
		jsonFlag,
	}

	app.Commands = []cli.Command{
		{
			Name:  "add",
			Usage: "add or replace subreddit with score in configuration",
			Action: func(c *cli.Context) {
				name := c.Args().First()
				score, err := strconv.Atoi(c.Args().Get(1))
				if err != nil {
					log.Fatalln(err)
				}

				configuration := &Configuration{}
				if err = configuration.LoadFromFile(c.GlobalString("config")); err != nil {
					log.Fatalln(err)
				}
				configuration.addSubreddit(name, score)

				err = configuration.DumpIntoFile(c.GlobalString("config"))
				if err != nil {
					log.Fatalln(err)
				}
			},
		},
		{
			Name:  "delete",
			Usage: "remove subreddit from configuration",
			Action: func(c *cli.Context) {
				configuration := &Configuration{}
				names := c.Args()
				if err := configuration.LoadFromFile(c.GlobalString("config")); err != nil {
					log.Fatalln(err)
				}
				for _, name := range names {
					configuration.deleteSubredditByName(name)
				}

				err = configuration.DumpIntoFile(c.GlobalString("config"))
				if err != nil {
					log.Fatalln(err)
				}
			},
		},
		{
			Name:  "list",
			Usage: "list subreddits from configuration",
			Action: func(c *cli.Context) {
				configuration := &Configuration{}
				if err := configuration.LoadFromFile(c.GlobalString("config")); err != nil {
					log.Fatalln("configuration file error:", err)
				}
				for _, subreddit := range configuration.Subreddits {
					terminal.Stdout.Color("y").Print(subreddit.Name).Print(" ").Color("g").Print(subreddit.Score).Nl().Reset()
				}
			},
		},
		{
			Name:  "get",
			Usage: "filter single subreddit by score",
			Action: func(c *cli.Context) {
				configuration := &Configuration{}
				name := c.Args().First()
				score, err := strconv.Atoi(c.Args().Get(1))
				if err != nil {
					log.Fatalln(err)
				}
				configuration.Subreddits = []*Subreddit{NewSubreddit(name, score)}
				load(configuration, c)
			},
		},
	}

	app.Action = func(c *cli.Context) {
		configuration := &Configuration{}
		if err := configuration.LoadFromFile(c.GlobalString("config")); err != nil {
			log.Fatalln("configuration file error:", err)
		}
		load(configuration, c)
	}

	app.Run(os.Args)
}
