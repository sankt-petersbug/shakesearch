# ShakeSearch

ShakeSearch is simple application for searching William Shakespeare's works.
See [example app](https://peter-shakesearch.herokuapp.com/)

NOTE: the example application is hosted on Heroku free tier and [goes to sleep](https://devcenter.heroku.com/articles/free-dyno-hours#dyno-sleeping) if it receives no web traffic in a 30-minute period. If that happens, it may not be able to return a full result so try it again later(takes about 130 seconds to index texts).

## Start Server

Using make:

```sh
$ make start
```
or you can also do:

```sh
$ go run main.go
```

then open `localhost:3000`

## GET /search

QueryParams:

- q (str): query string
- page[number] (int): page number to return
- page[size] (int): number of record in a page
- fuzziness (int): fuzzy search (default: 0)
- workId (str): search from a specific work
- sortBy (str): a comma-delimited list(prefix - to desc. -Title). available fields: Title, LineNumber, WorkID, _score 


```sh
$ curl 'localhost:3000/search?q=sonnet&fuzziness=1&page[size]=10&sortBy=Title,LineNumber'
```

Example Response:

```json
{
    "data": [
        {
            "line": "And deep-brain’d <mark>sonnets</mark> that did amplify",
            "lineNumber": 481,
            "score": 0.9557341597600069,
            "title": "A LOVER’S COMPLAINT",
            "workId": "ALOVERSCOMPLAINT"
        },
        {
            "line": "Good Captain, will you give me a copy of the <mark>sonnet</mark> you writ to Diana",
            "lineNumber": 7787,
            "score": 0.6758060938119992,
            "title": "ALL’S WELL THAT ENDS WELL",
            "workId": "ALLSWELLTHATENDSWELL"
        }
    ],
    "meta": {
        "highlight": {
            "postTag": "</mark>",
            "preTag": "<mark>"
        },
        "pageNumber": 1,
        "pageSize": 2,
        "totalResults": 39
    }
}
```

## GET /titles

```sh
$ curl localhost:3000/titles
```

Example Response:

```json
[
    {
        "title": "A LOVER’S COMPLAINT",
        "workId": "ALOVERSCOMPLAINT"
    },
    {
        "title": "A MIDSUMMER NIGHT’S DREAM",
        "workId": "AMIDSUMMERNIGHTSDREAM"
    },
]
```

## GET /works/:id

```sh
$ curl localhost:3000/works/ALOVERSCOMPLAINT
```

Example Response:

```json
{
    "content": "\n\n\n\n\n\nFrom off a hill whose concave womb reworded\n\nA plaintful story from a sist’ring vale,\n\nMy spirits t’attend this double voice accorded,\n\nAnd down I laid to list the sad-tun’d tale;\n\nEre long espied a fickle maid full pale,\n\nTearing of papers, breaking rings a-twain,\n\nStorming her world with sorrow’s wind and rain.\n\n\n\nUpon her head a platted hive of straw,\n\nWhich fortified her visage from the sun,\n\nWhereon the thought might think sometime it saw\n\nThe carcass of a beauty spent and done;\n\nTime had not scythed all that youth begun,\n\n...",
    "id": "ALOVERSCOMPLAINT",
    "title": "A LOVER’S COMPLAINT"
}
```

## TODO

- Divide work into sections/chapters (indexing each line is expensive and returning too many results for a user to parse)
- Performance tuning and benchmarks
- Improve search results
