Command-line tool to view posts from your favorite Reddit subreddits filtered by score.

![console gif](https://raw.githubusercontent.com/FZambia/fire/master/console.gif)

Default configuration file for `fire` located at `$HOME/.fire.json`

Overview
--------

Every day I wake up in the morning and check my favorite Reddit subreddits for a good new posts.
Sometimes I have no enough time to browse the web - so I've written `fire` - command-line
utility that keeps a list of my favorite subreddits and pretty prints current hot posts in console 
based on minimal score for subreddit I previously set in configuration.

Installation
------------

```bash
go get github.com/FZambia/fire
```

Usage
-----

Show help information:

```bash
fire --help
```

Add (Update) subreddit with minimal score in configuration:

```bash
fire add python 50
fire add gifs 3000
```

Get entries for subbreddits from configuration:

```bash
fire
```

Get entries filtered by score for subreddit not currently listed in configuration:

```bash
fire get golang 20
```

Show all current configuration subreddits:

```bash
fire list
```

Delete subreddit from configuration

```bash
fire delete python
```

Use custom configuration file

```bash
fire -c custom_config.json
```

To get JSON output in console:

```bash
fire --json
```

And finally the sweetest thing - browser output!

```
fire --browser get gifs 3000
```

And your default browser will open a tab with something like this:

![console gif](https://raw.githubusercontent.com/FZambia/fire/master/browser.png)


LICENSE
-------

MIT
