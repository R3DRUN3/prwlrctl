# prwlrctl  

<img src="./media/logo.png" alt="logo" width="250"/>

A small, fast Go CLI for the [Prowler Server](https://docs.prowler.com/getting-started/products/prowler-app)'s API.  
Built for both interactive operators and automations.    

Prowler API docs can be found [here](https://api.prowler.com/api/v1/docs).  

## Install

    go install github.com/r3drun3/prwlrctl@latest  

or build locally:

    make build   # produces ./bin/prwlrctl

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

## Usage

    prwlrctl providers list
    prwlrctl providers get <provider-id>

    prwlrctl scans list --state executing
    prwlrctl scans launch --provider <provider-id> --name "weekly scan" --wait
    prwlrctl scans get <scan-id>

    prwlrctl findings list --scan <scan-id> --severity critical

Add `-o json` to any command for machine-readable output (pipe into `jq`).
Add `-q` to `scans launch` to print just the new scan ID, handy for scripts:

    scan_id=$(prwlrctl scans launch --provider "$PROVIDER_ID" -q)
    prwlrctl scans get "$scan_id" -o json | jq .

## Exit codes

Non-zero exit on any API or network error: safe to rely on in cron/CI
(`&&`/`||` chaining, `set -e`, etc.).