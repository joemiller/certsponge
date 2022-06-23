# certsponge

Utility for splitting Vault's PKI output into one or more files containing the private key, certificate, and CA data.

```console
vault write pki -format=json pki/issue/rolename common_name=web.dom.tld | certsponge
```

## Install

Install latest using `go install`:

```console
go install github.com/joemiller/certsponge@latest
```

Pre-built binaries, packages, and docker images are available for various platforms on the [GitHub Releases](https://github.com/joemiller/certsponge/releases) page.


## Usage

`certsponge` expects to receive the JSON output from `vault write pki/issue/...`:

```console
vault write pki -format=json pki/issue/rolename common_name=web.dom.tld | certsponge
```

By default the output is saved into two files in the current directory:

- `tls.pem`: Contains `private_key`, `certificate`, and `ca_chain` (in that order).
- `ca.crt`: Contains `ca_chain`.

This behavior can be changed via flags. Run with `-h` for usage.


Files containing `private_key` are always created with mode `0600`.

Files containing only non-sensitive data (`certificate` and `ca_chain`) are created with mode `0644`.

Existing files will not be overwritten unless `-f` flag is specified.

## Motivation

I got tired of writing blocks this (and many other variations) in scripts:

```sh
out=$(vault write -format=json pki/issue/myrole common_name=foo)
key=$(jq -r '.data.private_key' <<<"$out")
cert=$(jq -r '.data.certificate' <<<"$out")
ca=$(jq -r '.data.ca_chain' <<<"$out")
{
  echo "$key"
  echo "$cert"
  echo "$ca"
} >tls.pem
```

## Similar Tools

- [vaultbot](https://gitlab.com/msvechla/vaultbot) is an excellent tool that implements the full
end-to-end process of requesting certs from Vault and writing them to files. It also handles
renewals. `certsponge` is not trying to do all of that, it's only goal is split the output
from the `vault` CLI into files.
