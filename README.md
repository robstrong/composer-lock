# Composer Lock

A simple JSON API which takes in a composer.json file and responds with it's composer.lock file.

## Usage

Build the binary
```bash
go build
```

Start the server
```bash
./composer-lock
```

Send a request containing your composer.json in the body
```bash
curl -X POST --data "@/path/to/composer.json" localhost:5799
```

The server will place the composer.json contents into a temporary directory and run `composer install`. The resulting lock file will be in the body of the response. If an error occurs, it will return JSON in the following format:
```json
{
    "Status": "error",
    "Detail": "details about the error"
}
```
