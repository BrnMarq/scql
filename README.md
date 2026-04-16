# SCQL: The Scraping Query Language

SCQL (Scraping Query Language) is a declarative, SQL-like language specifically designed to make web scraping and browser automation simple, intuitive, and readable. By treating websites as databases, SCQL allows you to extract data, authenticate, and configure crawling behavior using familiar syntax.

## Getting Started

You can interact with SCQL either through its interactive REPL (Read-Eval-Print Loop) or by executing script files.

**Start the interactive REPL:**

```bash
go run main.go
```

**Run an SCQL script file:**

```bash
go run main.go script.scql
```

## Language Reference

### SELECT Statement

The `SELECT` statement is the core of SCQL, used for extracting specific data points from target URLs. It supports filtering and ordering of the scraped data.

**Syntax:**

```sql
SELECT field1, field2
FROM "https://example.com"
WHERE condition
ORDER BY field1 ASC;
```

### AUTHENTICATE Statement

The `AUTHENTICATE` statement handles login flows and form submissions required to access protected pages before scraping.

**Syntax:**

```sql
AUTHENTICATE AT "https://example.com/login"
SUBMIT FORM WITH (
    username = "myuser",
    password = "mypassword"
);
```

### SET Statement

The `SET` statement allows you to configure global scraping behaviors and engine parameters, such as timeouts, retries, or toggling features.

**Syntax:**

```sql
SET
    "#search-input" = "web scraping tools",
    ".category-dropdown" = "software",
    "input[name='agree_terms']" = TRUE,
    "#promo-code" = NULL;
```

## Data Types & Operators

### Supported Data Types

- **Strings**: `"Hello, World"` or `'Text'`
- **Integers**: `42`, `-10`
- **Floats**: `3.14`, `-0.001`
- **Booleans**: `TRUE`, `FALSE`
- **Null**: `NULL`

### Supported Operators

- **Arithmetic**: `+` (add), `-` (subtract), `*` (multiply), `/` (divide)
- **Comparison**: `=`, `!=`, `>`, `<`, `>=`, `<=`
- **Logical**: `AND`, `OR`

### Comments

SCQL supports both single-line and multi-line comments:

```sql
-- This is a single-line comment

/*
   This is a
   multi-line comment
*/
```

## Full Example

Below is a complete, real-world scenario demonstrating how to configure the engine, log in to a portal, and extract the latest articles:

```sql
-- 1. Configure scraping behavior
SET scraping_enabled = TRUE,
    retries = 3,
    user_agent = "SCQL-Bot/1.0";

-- 2. Authenticate to access protected content
AUTHENTICATE AT "https://news.ycombinator.com/login"
SUBMIT FORM WITH (
    acct = "scql_user",
    pw = "supersecret123"
);

-- 3. Extract data from the target page
SELECT title, points, author
FROM "https://news.ycombinator.com/newest"
WHERE points > 50 AND author != "admin"
ORDER BY points DESC;
```
