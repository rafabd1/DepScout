# Harpy# DepScout



![Go Version](https://img.shields.io/github/go-mod/go-version/rafabd1/Harpy)![Go Version](https://img.shields.io/github/go-mod/go-version/rafabd1/DepScout)

![Release](https://img.shields.io/github/v/release/rafabd1/Harpy?include_prereleases)![Release](https://img.shields.io/github/v/release/rafabd1/DepScout?include_prereleases)

![Build Status](https://github.com/rafabd1/Harpy/workflows/Release%20Harpy/badge.svg)![Build Status](https://github.com/rafabd1/DepScout/workflows/Release%20DepScout/badge.svg)

![License](https://img.shields.io/badge/license-MIT-blue.svg)![License](https://img.shields.io/badge/license-MIT-blue.svg)

![GitHub stars](https://img.shields.io/github/stars/rafabd1/Harpy?style=social)![GitHub stars](https://img.shields.io/github/stars/rafabd1/DepScout?style=social)

![Go Report Card](https://goreportcard.com/badge/github.com/rafabd1/Harpy)![Go Report Card](https://goreportcard.com/badge/github.com/rafabd1/DepScout)



<div align="center"><div align="center">

<pre><pre>

    ██╗  ██╗ █████╗ ██████╗ ██████╗ ██╗   ██╗  _____             _____                 _   

    ██║  ██║██╔══██╗██╔══██╗██╔══██╗╚██╗ ██╔╝ |  __ \           / ____|               | |  

    ███████║███████║██████╔╝██████╔╝ ╚████╔╝  | |  | | ___ _ __| (___   ___ ___  _   _| |_ 

    ██╔══██║██╔══██║██╔══██╗██╔═══╝   ╚██╔╝   | |  | |/ _ \ '_ \\___ \ / __/ _ \| | | | __|

    ██║  ██║██║  ██║██║  ██║██║        ██║    | |__| |  __/ |_) |___) | (_| (_) | |_| | |_ 

    ╚═╝  ╚═╝╚═╝  ╚═╝╚═╝  ╚═╝╚═╝        ╚═╝    |_____/ \___| .__/_____/ \___\___/ \__,_|\__|

</pre>             | |                              

</div>             |_|                              

</pre>

<p align="center"></div>

    <b>High-performance web reconnaissance tool for extracting endpoints, parameters, and hidden assets from JavaScript files.</b>

</p><p align="center">

    <b>A high-performance, concurrent scanner for detecting unclaimed packages.</b>

## Features</p>



- **Advanced JavaScript Analysis**: Hybrid regex and AST-based parsing for accurate endpoint and parameter extraction from JavaScript files.## Features

- **Dual Analysis Modes**:

  - **Regex Mode**: Ultra-fast pattern matching for quick reconnaissance- **Dependency Confusion Scanning**: Identifies potential Dependency Confusion vulnerabilities by scanning JavaScript files for package names and checking their existence on the public npm registry.

  - **AST Mode**: Deep JavaScript parsing for comprehensive analysis with higher accuracy- **Dual Parsing Engine**:

  - **Hybrid Mode**: Intelligent combination of both approaches for optimal results  - **Regex Mode**: Fast, high-performance scanning using fine-tuned regular expressions to find `require()` and `import` statements.

- **Intelligent Extraction**: Discovers:  - **Deep Scan Mode (`--deep-scan`)**: High-accuracy scanning using a full Abstract Syntax Tree (AST) parser to eliminate false positives from non-code contexts (e.g., comments). Includes an automatic fallback to regex if AST parsing fails.

  - API endpoints and paths- **Concurrent Architecture**: Leverages a powerful, concurrent worker model to perform high-speed fetching and analysis of hundreds or thousands of files simultaneously.

  - URL parameters and form fields  - **Adaptive Rate Limiting**: Features a smart, per-domain rate limiter that automatically adjusts request speeds based on server responses (e.g., `429 Too Many Requests`), maximizing speed without overwhelming targets.

  - HTTP headers and authentication tokens- **Flexible Input**: Accepts targets from single URLs (`-u`), files (`-f`), local directories (`-d`), or piped via `stdin`.

  - Internal domains and subdomains- **Advanced Configuration**: Provides a rich set of CLI flags for fine-grained control over concurrency, timeouts, file sizes, and more.

  - Hidden admin panels and debug endpoints- **Multiple Output Formats**: Delivers results in human-readable text or machine-readable `JSON` (`--json`).

- **Concurrent Architecture**: High-speed processing with configurable worker pools for maximum efficiency

- **Adaptive Rate Limiting**: Smart per-domain throttling that automatically adjusts to server responses## Installation

- **Flexible Input Sources**: URLs, local files, directories, or stdin for seamless integration

- **Professional Output**: Clean terminal output plus structured JSON for automation and toolchain integration### From Source



## Installation```bash

# Clone the repository

### From Sourcegit clone https://github.com/rafabd1/DepScout.git

cd DepScout

```bash

# Clone the repository# Build the binary

git clone https://github.com/rafabd1/Harpy.gitgo build -o depscout ./cmd/depscout

cd Harpy

# Optional: Move to path (Linux/macOS)

# Build the binarysudo mv depscout /usr/local/bin/

go build -o harpy ./cmd/harpy```



# Optional: Move to path (Linux/macOS)### Using Go Install

sudo mv harpy /usr/local/bin/

``````bash

go install github.com/rafabd1/DepScout/cmd/depscout@latest

### Using Go Install```



```bash### Binary Releases

go install github.com/rafabd1/Harpy/cmd/harpy@latest

```You can download pre-built binaries for your platform from the [releases page](https://github.com/rafabd1/DepScout/releases).



### Binary Releases## Quick Start



Download pre-built binaries for your platform from the [releases page](https://github.com/rafabd1/Harpy/releases).Scan a single JavaScript file:

```bash

## Quick Startdepscout -u https://example.com/assets/app.js

```

Extract endpoints from a single JavaScript file:

```bashScan a local directory of JS files using the high-accuracy deep scan mode:

harpy -u https://example.com/assets/app.js```bash

```depscout -d /path/to/js/files --deep-scan

```

Scan local JavaScript files with AST analysis:

```bashScan a list of targets from a file:

harpy -d /path/to/js/files --enable-ast```bash

```depscout -f targets.txt

```

Scan from a list of targets and output as JSON:

```bash## Command Line Options

harpy -f targets.txt --json -o results.json

```| Flag | Description | Default |

|------|-------------|---------|

Extract parameters and headers in verbose mode:| `-u` | A single target URL or local file path. | - |

```bash| `-f` | A file containing a list of targets. | - |

harpy -u https://target.com/main.js --enable-regex --enable-ast -v| `-d` | Path to a local directory to scan for `.js` and `.ts` files. | - |

```| `-c` | Number of concurrent workers. | `25` |

| `-l` | Maximum requests per second per domain (in auto-adjustment mode). | `30` |

## Command Line Options| `-H` | Custom header to include in all requests (can be specified multiple times). | - |

| `-o` | File to write output to. | stdout |

### Input Options| `-p` | File containing a list of proxies (http/https/socks5). | - |

| Flag | Description | Default || `-proxy` | A single proxy server (e.g. http://127.0.0.1:8080). | - |

|------|-------------|---------|| `--deep-scan` | Enable deep scan using AST parsing (slower but more accurate). | `false` |

| `-u` | Single target URL or local file path | - || `-json` | Enable JSON output format. | `false` |

| `-f` | File containing list of targets (one per line) | - || `--max-file-size` | Maximum file size to process in KB. | `10240` |

| `-d` | Directory to scan for JavaScript/TypeScript files | - || `--no-limit` | Disable file size limit. | `false` |

| `--skip-verify` | Skip TLS certificate verification. | `false` |

### Analysis Options  | `-v` | Enable verbose output for debugging. | `false` |

| Flag | Description | Default || `--silent` | Suppress all output except for findings. | `false` |

|------|-------------|---------|| `--no-color` | Disable colorized output. | `false` |

| `--enable-regex` | Enable regex-based pattern matching | `true` |

| `--enable-ast` | Enable AST-based JavaScript parsing | `false` |

| `-c` | Number of concurrent workers | `25` |## Disclaimer

| `--max-file-size` | Maximum file size to process (KB) | `10240` |

| `--no-limit` | Disable file size restrictions | `false` |**Usage Warning & Responsibility**



### Network OptionsThis tool is intended for security professionals and researchers for legitimate testing purposes only. Running DepScout against a target may generate a high volume of HTTP requests. You are responsible for your actions and must have explicit permission to test any target. The author of this tool is not responsible for any misuse or damage caused by this program.

| Flag | Description | Default |

|------|-------------|---------|## Documentation

| `-l` | Requests per second limit per domain | `30` |

| `-H` | Custom headers (can be used multiple times) | - |- [Changelog](CHANGELOG.md) - Check the latest updates and version history.

| `-p` | Proxy list file | - |

| `--proxy` | Single proxy server | - |## Contributing

| `--skip-verify` | Skip TLS certificate verification | `false` |

| `--timeout` | Request timeout in seconds | `10` |Contributions are welcome! Please feel free to submit a Pull Request.



### Output Options## License

| Flag | Description | Default |

|------|-------------|---------|This project is licensed under the MIT License.

| `-o` | Output file path | stdout |

| `--json` | Output results in JSON format | `false` |

| `-v` | Enable verbose output | `false` |

| `--silent` | Suppress all output except findings | `false` |<p align="center">

| `--no-color` | Disable colored output | `false` |    <sub>Made with 🖤 by Rafael (github.com/rafabd1)</sub>

</p>

## Example Usage

<!-- <p align="center">

### Bug Bounty Reconnaissance    <a href="https://ko-fi.com/rafabd1" target="_blank"><img src="https://storage.ko-fi.com/cdn/kofi2.png?v=3" alt="Buy Me A Coffee" style="height: 60px !important;"></a>

```bash</p> -->

# Quick scan for endpoints and parameters
harpy -u https://target.com/app.js --enable-regex --enable-ast

# Comprehensive directory scan with JSON output
harpy -d ./js-files --enable-ast --json -o findings.json

# Scan with custom headers and proxy
harpy -f targets.txt -H "User-Agent: Harpy/1.0" --proxy http://127.0.0.1:8080
```

### Integration with Other Tools
```bash
# Pipe from subfinder and httpx
subfinder -d target.com | httpx -path /assets/app.js | harpy

# Extract endpoints and pass to ffuf
harpy -u https://target.com/main.js --json | jq -r '.results.findings[].endpoints[].path' | ffuf -u https://target.com/FUZZ -w -
```

## Output Formats

### Terminal Output
```
=== Extraction Results ===
✓ Sources processed: 1
✓ Domains found: 3  
✓ Endpoints found: 12
✓ Parameters found: 8
✓ Headers found: 4
```

### JSON Output
```json
{
  "metadata": {
    "tool": "Harpy",
    "version": "1.0.0", 
    "timestamp": "2024-01-15T10:30:00Z",
    "total_sources": 1
  },
  "results": {
    "findings": [
      {
        "source": "https://example.com/app.js",
        "domains": ["api.example.com", "admin.example.com"],
        "endpoints": [
          {
            "method": "GET",
            "path": "/api/v1/users",
            "context": "fetch('/api/v1/users')"
          }
        ],
        "parameters": [
          {
            "name": "userId", 
            "type": "url",
            "context": "?userId=123"
          }
        ],
        "headers": [
          {
            "name": "Authorization",
            "context": "headers: {'Authorization': token}"
          }
        ]
      }
    ]
  },
  "summary": {
    "total_domains": 2,
    "total_endpoints": 1, 
    "total_parameters": 1,
    "total_headers": 1
  }
}
```

## Disclaimer

**Usage Warning & Responsibility**

This tool is designed for security professionals, bug bounty hunters, and researchers for legitimate testing purposes only. Users must have explicit permission to test any target. The author is not responsible for any misuse or damage caused by this program.

## Documentation

- [Architecture Overview](dev/HARPY-ARCHITECTURE.md) - Technical details and design decisions
- [Changelog](CHANGELOG.md) - Version history and updates

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

---

<p align="center">
    <sub>Made with 🔍 for bug bounty hunters and security researchers</sub>
</p>

<p align="center">
    <sub>Created by Rafael (github.com/rafabd1)</sub>
</p>