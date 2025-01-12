# sing-srs-converter

A Go-based tool to convert Clash rule provider files to sing-box rule-set format.

## Features

- Converts Clash rule provider YAML files to sing-box rule-set format
- Supports both source type (.json) and binary type (.srs) rule-sets
- Supports various rule types:
  - Domain rules (DOMAIN, DOMAIN-SUFFIX, DOMAIN-KEYWORD, DOMAIN-REGEX)
  - IP rules (IP-CIDR, IP-CIDR6, SRC-IP-CIDR)
  - Port rules (DST-PORT, SRC-PORT)
  - Process rules (PROCESS-NAME, PROCESS-PATH)
- Optional mix mode to combine different rule types into a single file

## Installation

```bash
go install github.com/puernya/sing-srs-converter@latest
```

## Usage

Basic usage:
```bash
sing-srs-converter [source-path]
```

Options:
```bash
Flags:
  -h, --help            help for sing-srs-converter
  -m, --mix            Enable mix mode to combine different rule types
  -o, --output string  Output file name (default "<file_name>.srs")
```
## Supported Formats

The converter supports two input formats:

1. YAML format (with payload):
```yaml
payload:
  - DOMAIN,example.com
  - IP-CIDR,192.168.1.0/24
```

2. LIST format (direct rules):
```
DOMAIN,example.com
IP-CIDR,192.168.1.0/24
```

Both formats can be used with either .yaml or .list file extensions.

## Output Files

Without mix mode (-m):
- `<output>-domain-v2.json`: Domain related rules in source format
- `<output>-domain.srs`: Domain related rules in binary format
- `<output>-ip-v2.json`: IP related rules in source format
- `<output>-ip.srs`: IP related rules in binary format
- `<output>-port-v2.json`: Port related rules in source format
- `<output>-port.srs`: Port related rules in binary format
- `<output>-process-v2.json`: Process related rules in source format
- `<output>-process.srs`: Process related rules in binary format

With mix mode (-m):
- `<output>-v2.json`: All rules combined in source format
- `<output>.srs`: All rules combined in binary format

## Example

Convert a Clash rule provider file:
```bash
sing-srs-converter rules.yaml -o converted
```

Convert and combine all rules into single files:
```bash
sing-srs-converter rules.yaml -o converted -m
```

## License

This project is licensed under the MIT License.

