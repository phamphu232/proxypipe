# proxypipe

# Usage

```
# Create file proxies.txt
cp proxies.txt.example proxies.txt
sudo nano proxies.txt

# Start container
docker compose up -d

# Get pipe
docker logs proxypipe
```

# Build

```
docker run --rm -v ./:/app -w /app golang:1.23.1-alpine3.19 go build -o proxypipe proxypipe.go

## Manual build
docker run -it -v ./:/app -w /app --rm golang:1.23.1-alpine3.19
go mod init proxypipe
go mod tidy
go build -o proxypipe proxypipe.go
```