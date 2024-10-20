# Key Value Store

## Overview
A basic key value store implementation in Go that writes data into disk.

## Installation
1. Clone this repo locally & cd into the cloned folder
2. Run the server (default PORT - 8000)
```bash
 go run main.go
```
3. You can send requests for CRUD operations
4. It will create 2 files - data.json for storing data & logs.log for storing logs

## Features
- Supports TTL(Time To Live)
- Can handle multiple requests concurrently
- Implements basic CRUD (Create, Read, Update, Delete) operations for key-value pairs.
- Has simple persistence to disk (Writes to disk after every specified count or time)
- Implements basic logging for operations


## License

[MIT](https://choosealicense.com/licenses/mit/)