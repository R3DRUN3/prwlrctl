# prwlrctl  

[![ci](https://github.com/R3DRUN3/prwlrctl/actions/workflows/ci.yml/badge.svg)](https://github.com/R3DRUN3/prwlrctl/actions/workflows/ci.yml) [![License: Unlicense](https://img.shields.io/badge/license-Unlicense-blue.svg)](http://unlicense.org/)  
[![Latest Release](https://img.shields.io/github/v/release/r3drun3/prwlrctl?logo=github)](https://github.com/r3drun3/prwlrctl/releases/latest)  [![Go](https://img.shields.io/github/go-mod/go-version/R3DRUN3/prwlrctl?logo=go)](https://github.com/R3DRUN3/prwlrctl/blob/main/go.mod)  

<img src="./media/logo.png" alt="logo" width="250"/>

Unofficial CLI for the [Prowler Server](https://docs.prowler.com/getting-started/products/prowler-app)'s API.  
Built for both human and non human operators.     

Prowler API docs can be found [here](https://api.prowler.com/api/v1/docs).  

## Run  
You have many options to retrieve and run the cli:  

- clone the repo and build locally:  
    ```bash
    make build   # produces ./bin/prwlrctl
    ```  
- retrieve the [release](https://github.com/R3DRUN3/prwlrctl/releases) you want 

- download the [docker image](https://github.com/R3DRUN3/prwlrctl/pkgs/container/prwlrctl).  

### Run with Docker
For example, if you want to launch the CLI via docker on the same machine as the Prowler server, first define a `.env` file:  
```txt
PRWLRCTL_BASE_URL=http://localhost:8080/api/v1
PRWLRCTL_API_KEY=pk_YOURKEY
```  

Then launch the container:  
```bash
docker run --rm \
--network host \
--env-file path-to-your-.env-file \
ghcr.io/r3drun3/prwlrctl:0.1.0 \
providers list -o json
```  



## Authentication

Two options, resolved with priority flags > env vars > config file:

- **API key** (recommended for automation): create one in the Prowler UI
  under Profile → Account → API Keys, then:

      export PRWLRCTL_API_KEY="pk_xxx"
      export PRWLRCTL_BASE_URL="https://api.prowler.com/api/v1"

- **JWT login** (for interactive human use):

      prwlrctl auth login --email you@example.com --password '...'
      prwlrctl auth refresh   # when the access token expires

Or persist settings once:

    prwlrctl configure --base-url https://your-host/api/v1 --api-key pk_xxx

## Usage Examples

```bash
prwlrctl help
prwlrctl providers list
prwlrctl providers get <provider-id>

prwlrctl scans list --state completed
prwlrctl scans launch --provider <provider-id> --name "weekly scan"
prwlrctl scans get <scan-id>

prwlrctl findings list --scan <scan-id> --severity critical
```  


Add `-o json` to any command for machine-readable output (pipe into `jq`).
Add `-q` to `scans launch` to print just the new scan ID, handy for scripts:

    scan_id=$(prwlrctl scans launch --provider "$PROVIDER_ID" -q)
    prwlrctl scans get "$scan_id" -o json | jq .




## Exit codes

Non-zero exit on any API or network error: safe to rely on in cron/CI
(`&&`/`||` chaining, `set -e`, etc.).