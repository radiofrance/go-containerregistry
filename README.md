go-containerregistry
=====================

Small wrapper library for [google/go-containerregistry](https://github.com/google/go-containerregistry) 
designed to remove boilerplate code and make authentication to private registries transparent.

## Installation

Install the module:
```shell
go get github.com/radiofrance/go-containerregistry
```

## Basic usage

Create a new registry client, and get the SHA of an image:
```go
package main

import (
	"fmt"
	"log"

	registry "github.com/radiofrance/go-containerregistry"
)

func main() {
	gcr, err := registry.New("eu.gcr.io/project-id")
	if err != nil {
		log.Fatalf("Failed to create the registry client: %v", err)
	}

	head, err := gcr.Head("eu.gcr.io/project-id/nginx:latest")
	if err != nil {
		log.Fatalf("Failed to fetch remote: %v", err)
	}

	fmt.Printf("The SHA is %s\n", head.Digest.Hex)
}
```

## Authentication

If your registry is private, the client needs credentials to authenticate its requests to the remote API.

This is done pretty easily by:
- Creating a file containing a service account JSON key. The service account must have at least read access to the artifacts' storage bucket.
- Setting the `GCR_JSON_KEY_PATH` environment variable, it must contain the path to the file we created.

Example:
```shell
gcloud iam service-accounts keys create /secrets/credentials.json \
    --iam-account=<serviceaccount>@<project-id>.iam.gserviceaccount.com
export GCR_JSON_KEY_PATH=/secrets/credentials.json
```

Now the library will automatically detect the credentials and authenticate the requests.
