Command-line tool to show posts from your favorite Reddit subreddits filtered by score

Default configuration file for `fire` located at $HOME/.fire.json

Installation:
-------------

```bash
go get github.com/FZambia/fire
```

Usage:

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

LICENCE
-------

MIT
