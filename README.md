<img src="https://img.shields.io/badge/lifecycle-experimental-orange.svg" alt="Life cycle: experimental"> [![GoDoc](https://godoc.org/github.com/Miachol/bapi?status.svg)](https://godoc.org/github.com/Miachol/bapi)

# bapi

- NCBI query

## Installation

```bash
go get -u github.com/Miachol/bapi
```

## Usage

**Main interface:**

```bash
Query bioinformatics website APIs. More see here https://github.com/Miachol/bapi.

Usage:
  bapi [flags]
  bapi [command]

Available Commands:
  gdc         Query GDC portal website APIs.
  help        Help about any command
  ncbi        Query ncbi website APIs.

Flags:
  -e, --email string    Email specifies the email address to be sent to the server (NCBI website is required). (default "your_email@domain.com")
      --format string   Rettype specifies the format of the returned data (CSV, TSV, JSON for gdc; XML/TEXT for ncbi).
      --from int        Parameters of API control the start item of retrived data. (default -1)
  -h, --help            help for bapi
  -o, --outfn string    Out specifies destination of the returned data (default to stdout).
  -q, --query string    Query specifies the search query for record retrieval (required).
      --quiet           No log output.
  -r, --retries int     Retry specifies the number of attempts to retrieve the data. (default 5)
      --size int        Parameters of API control the lenth of retrived data. Default is auto determined. (default -1)
      --version         version for bapi

Use "bapi [command] --help" for more information about a command.
```

**GDC query:**

```bash
Query GDC portal APIs. More see here https://github.com/Miachol/bapi.

Usage:
  bapi gdc [flags]

Examples:
  bapi gdc -p
  bapi gdc -p --json-pretty
  bapi gdc -p -q TARGET-NBL --json-pretty
  bapi gdc -p --format TSV > tcga_projects.tsv
  bapi gdc -p --format CSV > tcga_projects.csv
  bapi gdc -p --from 1 --szie 2
  bapi gdc -s
  bapi gdc -c
  bapi gdc -f
  bapi gdc -a

  // Download manifest for gdc-client
  bapi gdc -m -q "5b2974ad-f932-499b-90a3-93577a9f0573,556e5e3f-0ab9-4b6c-aa62-c42f6a6cf20c" -o my_manifest.txt
  bapi gdc -m -q "5b2974ad-f932-499b-90a3-93577a9f0573,556e5e3f-0ab9-4b6c-aa62-c42f6a6cf20c" > my_manifest.txt
  bapi gdc -m -q "5b2974ad-f932-499b-90a3-93577a9f0573,556e5e3f-0ab9-4b6c-aa62-c42f6a6cf20c" -n

  // Download data
  bapi gdc -d -q "5b2974ad-f932-499b-90a3-93577a9f0573" -n

Flags:
  -a, --annotations     Retrive annotations info from GDC portal.
  -c, --cases           Retrive cases info from GDC portal.
  -d, --data            Retrive /data from GDC portal.
  -f, --files           Retrive files info from GDC portal.
      --filter string   Retrive data with GDC filter.
  -h, --help            help for gdc
      --json            Retrive JSON data.
      --json-pretty     Retrive pretty JSON data.
  -l, --legacy          Use legacy API of GDC portal.
  -m, --manifest        Retrive /manifest data from GDC portal.
  -p, --projects        Retrive projects meta info from GDC portal.
  -n, --remote-name     Use remote defined filename.
      --slicing         Retrive BAM slicing from GDC portal.
  -s, --status          Check GDC portal status (https://portal.gdc.cancer.gov/).
      --token string    Token to access GDC.

Global Flags:
  -e, --email string    Email specifies the email address to be sent to the server (NCBI website is required). (default "your_email@domain.com")
      --format string   Rettype specifies the format of the returned data (CSV, TSV, JSON for gdc; XML/TEXT for ncbi).
      --from int        Parameters of API control the start item of retrived data. (default -1)
  -o, --outfn string    Out specifies destination of the returned data (default to stdout).
  -q, --query string    Query specifies the search query for record retrieval (required).
      --quiet           No log output.
  -r, --retries int     Retry specifies the number of attempts to retrieve the data. (default 5)
      --size int        Parameters of API control the lenth of retrived data. Default is auto determined. (default -1)

```

**Query NCBI:**

```bash
Query ncbi website APIs. More see here https://github.com/Miachol/bapi.

Usage:
  bapi ncbi [flags]

Examples:
  bapi ncbi -d pubmed -q B-ALL --format XML -e your_email@domain.com
  bapi ncbi -q "RNA-seq and bioinformatics[journal]" -e "your_email@domain.com" -m 100 | awk '/<[?]xml version="1.0" [?]>/{close(f); f="abstract.http.XML.tmp" ++c;next} {print>f;}'

  k="algorithm, tool, model, pipleline, method, database, workflow, dataset, bioinformatics, sequencing, http, github.com, gitlab.com, bitbucket.org, RNA-Seq, DNA, profile, landscape"
  echo "[" > final.json
  bapi ncbi --xml2json pubmed abstract.http.XML.tmp* -k "${k}"| sed 's/}{/},{/g' >> final.json
  echo "]" >> final.json

Flags:
  -d, --db string         Db specifies the database to search (default "pubmed")
  -h, --help              help for ncbi
  -k, --keywords string   Keywords to extracted from abstract. (default "algorithm, tool, model, pipleline, method, database, workflow, dataset, bioinformatics, sequencing, http, github.com, gitlab.com, bitbucket.org")
  -m, --per-size int      Retmax specifies the number of records to be retrieved per request. (default 100)
  -t, --thread int        Thread to parse XML from local files. (default 2)
      --xml2json string   Convert XML files to json [e.g. pubmed].

Global Flags:
  -e, --email string    Email specifies the email address to be sent to the server (NCBI website is required). (default "your_email@domain.com")
      --format string   Rettype specifies the format of the returned data (CSV, TSV, JSON for gdc; XML/TEXT for ncbi).
      --from int        Parameters of API control the start item of retrived data. (default -1)
  -o, --outfn string    Out specifies destination of the returned data (default to stdout).
  -q, --query string    Query specifies the search query for record retrieval (required).
      --quiet           No log output.
  -r, --retries int     Retry specifies the number of attempts to retrieve the data. (default 5)
      --size int        Parameters of API control the lenth of retrived data. Default is auto determined. (default -1)

```

## Maintainer

- [@Jianfeng](https://github.com/Miachol)

## License

Apache 2.0
