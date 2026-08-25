# prwlrctl  

[![ci](https://github.com/R3DRUN3/prwlrctl/actions/workflows/ci.yml/badge.svg)](https://github.com/R3DRUN3/prwlrctl/actions/workflows/ci.yml) [![License: Unlicense](https://img.shields.io/badge/license-Unlicense-blue.svg)](http://unlicense.org/)  
[![Latest Release](https://img.shields.io/github/v/release/r3drun3/prwlrctl?logo=github)](https://github.com/r3drun3/prwlrctl/releases/latest)  [![Go](https://img.shields.io/github/go-mod/go-version/R3DRUN3/prwlrctl?logo=go)](https://github.com/R3DRUN3/prwlrctl/blob/main/go.mod)  

<img src="./media/logo_2.png" alt="logo" width="250"/>  
</br>


Unofficial CLI for the [Prowler Server](https://docs.prowler.com/getting-started/products/prowler-app) API.  
Built for both human and non-human operators.  

> [!TIP]
> This tool is mainly intended to streamline management and data retrieval for self-hosted instances of Prowler Server.  
> The API docs for your self-hosted Prowler Server can be found at `http://localhost:8080/api/v1/docs`.  



## Run  
You have many options to retrieve and run the cli:  

- retrieve the [release](https://github.com/R3DRUN3/prwlrctl/releases) you want. 

- download the [docker image](https://github.com/R3DRUN3/prwlrctl/pkgs/container/prwlrctl).  

- clone the repo and build locally:  
    ```console
    make build   # produces ./bin/prwlrctl
    ```  

### Run with Docker
If you want to launch the CLI via docker on the same machine as the Prowler server, first define a `.env` file:  
```console
PROWLER_BASE_URL=http://localhost:8080/api/v1
PROWLER_API_KEY=pk_YOURKEY
```  

Then launch the container:  
```console
docker run --rm \
--network host \
--env-file path-to-your-.env-file \
ghcr.io/r3drun3/prwlrctl:0.4.0 \
providers list -o json
```  



## Authentication

Two options, resolved with priority flags > env vars > config file:

- **API key** (recommended): create one in the Prowler UI, then:
    ```console
    export PROWLER_API_KEY="pk_xxx"
    export PROWLER_BASE_URL="https://api.prowler.com/api/v1"
    ```  

- **JWT login**:  
    ```console
    prwlrctl auth login --email you@example.com --password '...'
    prwlrctl auth refresh   # when the access token expires
    ```  

Or persist settings once:
```console
prwlrctl configure --base-url https://your-host/api/v1 --api-key pk_xxx
```  

## Usage Examples

```console
prwlrctl --version
prwlrctl help
prwlrctl health
prwlrctl providers list
prwlrctl providers get <provider-id>

prwlrctl scans list --state completed
prwlrctl scans launch --provider <provider-id> --name "weekly scan"
prwlrctl scans get <scan-id>
prwlrctl scans compliance-overview <scan-id>

prwlrctl findings list --scan <scan-id> --severity critical
prwlrctl findings get <finding-id>
prwlcrctl resources get <resource-id>
```  


Add `-o json` to any command for machine-readable output (pipe into `jq`).
Add `-q` to `scans launch` to print just the new scan ID, handy for scripts:  

```console
scan_id=$(prwlrctl scans launch --provider "$PROVIDER_ID" -q)
prwlrctl scans get "$scan_id" -o json | jq .  
```  


## Import package
Starting with `v0.4.0`, prwlrctl can also be imported as a Go package and used as a reusable Prowler API client in other Go projects.  

The reusable client is available under `github.com/r3drun3/prwlrctl/pkg/prowler`.  

Install it with:

```console
go get github.com/r3drun3/prwlrctl@v0.4.0 (or latest, or the version you want)
```  


You can then create a Prowler client and use the exported API methods directly, see the example below:  

```go
package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/joho/godotenv"

	"github.com/r3drun3/prwlrctl/pkg/prowler" // imported here
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Printf("warning: could not load .env: %v", err)
	}

	baseURL := os.Getenv("PROWLER_BASE_URL")
	apiKey := os.Getenv("PROWLER_API_KEY")
	accessToken := os.Getenv("PROWLER_ACCESS_TOKEN")

	if baseURL == "" {
		log.Fatal("PROWLER_BASE_URL is required")
	}

	if apiKey == "" && accessToken == "" {
		log.Fatal("either PROWLER_API_KEY or PROWLER_ACCESS_TOKEN is required")
	}

	// Create a reusable Prowler client.
	prowlerClient := prowler.New(
		baseURL,
		apiKey,
		accessToken,
		30*time.Second,
	)

	ctx := context.Background()

	// ------------------------------------------------------------
	// 1. Server health check
	// ------------------------------------------------------------

	fmt.Println("Checking Prowler server health...")

	health, err := prowlerClient.Health(ctx)
	if err != nil {
		log.Fatalf("health check failed: %v", err)
	}

	fmt.Printf("✓ Prowler is healthy\n")
	fmt.Printf("  Status:      %s\n", health.Status)
	fmt.Printf("  Version:     %s\n", health.Version)
	fmt.Printf("  Release ID:  %s\n", health.ReleaseID)
	fmt.Printf("  Service ID:  %s\n", health.ServiceID)
	fmt.Printf("  Description: %s\n", health.Description)

	// ------------------------------------------------------------
	// 2. Retrieve providers
	// ------------------------------------------------------------

	fmt.Println("\nRetrieving providers...")

	// No filters, first page, 100 items.
	providersDocument, err := prowlerClient.ListProviders(
		ctx,
		nil,
		1,
		100,
	)
	if err != nil {
		log.Fatalf("failed to retrieve providers: %v", err)
	}

	providers, err := providersDocument.Many()
	if err != nil {
		log.Fatalf("failed to decode providers: %v", err)
	}

	fmt.Printf("✓ Found %d providers\n\n", len(providers))

	for _, provider := range providers {
		fmt.Printf("Provider:\n")
		fmt.Printf("  ID:      %s\n", provider.ID)
		fmt.Printf("  Type:    %s\n", provider.Type)
		fmt.Printf("  Alias:   %s\n", provider.Str("alias"))
		fmt.Printf("  Provider: %s\n", provider.Str("provider"))

		if connected, ok := provider.NestedBool("connection", "connected"); ok {
			fmt.Printf("  Connected: %t\n", connected)
		}

		fmt.Println()
	}
}
```


## Local development
In order to develop and debug the code locally I suggest using vscode.  
Create the `.vscode/launch.json` file and add your debug configurations, like this:  
```json
{
  "version": "0.2.0",
  "configurations": [
    {
      "name": "Debug prwlrctl --version",
      "type": "go",
      "request": "launch",
      "mode": "debug",
      "program": "${workspaceFolder}/main.go",
      "args": ["--version"],
      "cwd": "${workspaceFolder}"
    },
    {
      "name": "Debug prwlrctl providers list",
      "type": "go",
      "request": "launch",
      "mode": "debug",
      "program": "${workspaceFolder}/main.go",
      "args": ["providers", "list", "--all"],
      "cwd": "${workspaceFolder}",
      "env": {
        "PROWLER_BASE_URL": "http://localhost:8080/api/v1",
        "PROWLER_API_KEY": "pk_YOUR-PROWLER-SERVER-API-KEY-HERE"
      }
    },
    {
      "name": "Debug prwlrctl scans list",
      "type": "go",
      "request": "launch",
      "mode": "debug",
      "program": "${workspaceFolder}/main.go",
      "args": ["scans", "list"],
      "cwd": "${workspaceFolder}",
      "env": {
        "PROWLER_BASE_URL": "http://localhost:8080/api/v1",
        "PROWLER_API_KEY": "pk_YOUR-PROWLER-SERVER-API-KEY-HERE"
      }
    }
  ]
}

```