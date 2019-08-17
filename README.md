<img src="https://img.shields.io/badge/lifecycle-experimental-orange.svg" alt="Life cycle: experimental"> [![GoDoc](https://godoc.org/github.com/JhuangLab/bquery?status.svg)](https://godoc.org/github.com/JhuangLab/bquery)

# bquery

- NCBI query

## Installation

```bash
go get -u github.com/JhuangLab/bquery
```

## Usage

```bash
=== main ===
Query bioinformatics website APIs. More see here https://github.com/JhuangLab/bquery.

Usage:
  bquery [flags]
  bquery [command]

Examples:
  bquery ncbi -d pubmed -q B-ALL -t XML -e your_email@domain.com

Available Commands:
  help        Help about any command
  ncbi        Query ncbi website APIs.

Flags:
  -h, --help      help for bquery
      --quiet     No log output.
      --version   version for bquery

Use "bquery [command] --help" for more information about a command.

=== bquery ncbi ===
Query ncbi website APIs. More see here https://github.com/JhuangLab/bquery.

Usage:
  bquery ncbi [flags]

Examples:
  bquery ncbi -d pubmed -q B-ALL -t XML -e your_email@domain.com

Flags:
  -d, --db string        Db specifies the database to search (default "pubmed")
  -e, --email string     Email specifies the email address to be sent to the server (required).
  -h, --help             help for ncbi
  -o, --outfn string     Out specifies destination of the returned data (default to stdout).
  -q, --query string     Query specifies the search query for record retrieval (required).
  -m, --retmax int       Retmax specifies the number of records to be retrieved per request. (default 500)
  -r, --retries int      Retry specifies the number of attempts to retrieve the data. (default 5)
  -t, --rettype string   Rettype specifies the format of the returned data. (default "XML")

Global Flags:
      --quiet   No log output.
```

## Maintainer

- [@Jianfeng](https://github.com/Miachol)

## License

Apache 2.0
