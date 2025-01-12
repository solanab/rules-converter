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

## Version Control

Use `-v` flag to specify the rule set version:
```bash
sing-srs-converter rules.yaml -o output -v 2  # Generate version 2 rule set
```

Supported versions:
- Version 1: Legacy format
- Version 2: Current format
- Version 3: Latest format (default)

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

## Development

### GitHub Release Process

1. Create a new release in GitHub
   - Tag format: `v1.0.0` (following semantic versioning)
   - The release workflow will automatically build and attach binaries for all supported platforms

### Build From Source

```bash
# Clone the repository
git clone https://github.com/puernya/sing-srs-converter.git

# Change to project directory
cd sing-srs-converter

# Build the project
go build
```

### Supported Platforms

- Linux (x86_64, x86_64_v3, arm64)
- Windows (x86_64, x86_64_v3, arm64)
- macOS (x86_64, x86_64_v3, arm64)
