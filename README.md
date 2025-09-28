# Harpy# Harpy# DepScout



![Go Version](https://img.shields.io/github/go-mod/go-version/rafabd1/Harpy)

![Release](https://img.shields.io/github/v/release/rafabd1/Harpy?include_prereleases)

![Build Status](https://github.com/rafabd1/Harpy/workflows/Release%20Harpy/badge.svg)![Go Version](https://img.shields.io/github/go-mod/go-version/rafabd1/Harpy)![Go Version](https://img.shields.io/github/go-mod/go-version/rafabd1/DepScout)

![License](https://img.shields.io/badge/license-MIT-blue.svg)

![GitHub stars](https://img.shields.io/github/stars/rafabd1/Harpy?style=social)![Release](https://img.shields.io/github/v/release/rafabd1/Harpy?include_prereleases)![Release](https://img.shields.io/github/v/release/rafabd1/DepScout?include_prereleases)

![Go Report Card](https://goreportcard.com/badge/github.com/rafabd1/Harpy)

![Build Status](https://github.com/rafabd1/Harpy/workflows/Release%20Harpy/badge.svg)![Build Status](https://github.com/rafabd1/DepScout/workflows/Release%20DepScout/badge.svg)

<div align="center">

<pre>![License](https://img.shields.io/badge/license-MIT-blue.svg)![License](https://img.shields.io/badge/license-MIT-blue.svg)

    ██╗  ██╗ █████╗ ██████╗ ██████╗ ██╗   ██╗

    ██║  ██║██╔══██╗██╔══██╗██╔══██╗╚██╗ ██╔╝![GitHub stars](https://img.shields.io/github/stars/rafabd1/Harpy?style=social)![GitHub stars](https://img.shields.io/github/stars/rafabd1/DepScout?style=social)

    ███████║███████║██████╔╝██████╔╝ ╚████╔╝ 

    ██╔══██║██╔══██║██╔══██╗██╔═══╝   ╚██╔╝  ![Go Report Card](https://goreportcard.com/badge/github.com/rafabd1/Harpy)![Go Report Card](https://goreportcard.com/badge/github.com/rafabd1/DepScout)

    ██║  ██║██║  ██║██║  ██║██║        ██║   

    ╚═╝  ╚═╝╚═╝  ╚═╝╚═╝  ╚═╝╚═╝        ╚═╝   

</pre>

</div><div align="center"><div align="center">



<p align="center"><pre><pre>

    <b>High-performance web reconnaissance tool for extracting endpoints, parameters, and hidden assets from JavaScript files.</b>

</p>    ██╗  ██╗ █████╗ ██████╗ ██████╗ ██╗   ██╗  _____             _____                 _   



## Features    ██║  ██║██╔══██╗██╔══██╗██╔══██╗╚██╗ ██╔╝ |  __ \           / ____|               | |  



- **Intelligent Hybrid Analysis**: Automatically combines regex and AST-based parsing for maximum accuracy    ███████║███████║██████╔╝██████╔╝ ╚████╔╝  | |  | | ___ _ __| (___   ___ ___  _   _| |_ 

- **Smart Fallbacks**: Seamlessly switches between analysis methods based on file type and content

- **Comprehensive Extraction**: Discovers:    ██╔══██║██╔══██║██╔══██╗██╔═══╝   ╚██╔╝   | |  | |/ _ \ '_ \\___ \ / __/ _ \| | | | __|

  - API endpoints and paths

  - URL parameters and form fields      ██║  ██║██║  ██║██║  ██║██║        ██║    | |__| |  __/ |_) |___) | (_| (_) | |_| | |_ 

  - HTTP headers and authentication tokens

  - Internal domains and subdomains    ╚═╝  ╚═╝╚═╝  ╚═╝╚═╝  ╚═╝╚═╝        ╚═╝    |_____/ \___| .__/_____/ \___\___/ \__,_|\__|

  - Hidden admin panels and debug endpoints

- **High Performance**: Multi-threaded processing with configurable worker pools</pre>             | |                              

- **Adaptive Rate Limiting**: Smart per-domain throttling that respects server responses

- **Flexible Input**: URLs, local files, directories, or stdin for seamless tool integration</div>             |_|                              

- **Professional Output**: Clean terminal display plus structured JSON for automation

</pre>

## Installation

<p align="center"></div>

### From Source

    <b>High-performance web reconnaissance tool for extracting endpoints, parameters, and hidden assets from JavaScript files.</b>

```bash

# Clone the repository</p><p align="center">

git clone https://github.com/rafabd1/Harpy.git

cd Harpy    <b>A high-performance, concurrent scanner for detecting unclaimed packages.</b>



# Build the binary## Features</p>

go build -o harpy ./cmd/harpy



# Optional: Move to path (Linux/macOS)

sudo mv harpy /usr/local/bin/- **Advanced JavaScript Analysis**: Hybrid regex and AST-based parsing for accurate endpoint and parameter extraction from JavaScript files.## Features

```

- **Dual Analysis Modes**:

### Using Go Install

  - **Regex Mode**: Ultra-fast pattern matching for quick reconnaissance- **Dependency Confusion Scanning**: Identifies potential Dependency Confusion vulnerabilities by scanning JavaScript files for package names and checking their existence on the public npm registry.

```bash

go install github.com/rafabd1/Harpy/cmd/harpy@latest  - **AST Mode**: Deep JavaScript parsing for comprehensive analysis with higher accuracy- **Dual Parsing Engine**:

```

  - **Hybrid Mode**: Intelligent combination of both approaches for optimal results  - **Regex Mode**: Fast, high-performance scanning using fine-tuned regular expressions to find `require()` and `import` statements.

### Binary Releases

- **Intelligent Extraction**: Discovers:  - **Deep Scan Mode (`--deep-scan`)**: High-accuracy scanning using a full Abstract Syntax Tree (AST) parser to eliminate false positives from non-code contexts (e.g., comments). Includes an automatic fallback to regex if AST parsing fails.

Download pre-built binaries for your platform from the [releases page](https://github.com/rafabd1/Harpy/releases).

  - API endpoints and paths- **Concurrent Architecture**: Leverages a powerful, concurrent worker model to perform high-speed fetching and analysis of hundreds or thousands of files simultaneously.

## Quick Start

  - URL parameters and form fields  - **Adaptive Rate Limiting**: Features a smart, per-domain rate limiter that automatically adjusts request speeds based on server responses (e.g., `429 Too Many Requests`), maximizing speed without overwhelming targets.

Extract endpoints from a single JavaScript file:

```bash  - HTTP headers and authentication tokens- **Flexible Input**: Accepts targets from single URLs (`-u`), files (`-f`), local directories (`-d`), or piped via `stdin`.

harpy -u https://example.com/assets/app.js

```  - Internal domains and subdomains- **Advanced Configuration**: Provides a rich set of CLI flags for fine-grained control over concurrency, timeouts, file sizes, and more.



Scan local JavaScript files:  - Hidden admin panels and debug endpoints- **Multiple Output Formats**: Delivers results in human-readable text or machine-readable `JSON` (`--json`).

```bash

harpy -d /path/to/js/files- **Concurrent Architecture**: High-speed processing with configurable worker pools for maximum efficiency

```

- **Adaptive Rate Limiting**: Smart per-domain throttling that automatically adjusts to server responses## Installation

Scan from a list of targets and output as JSON:

```bash- **Flexible Input Sources**: URLs, local files, directories, or stdin for seamless integration

harpy -f targets.txt --json -o results.json

```- **Professional Output**: Clean terminal output plus structured JSON for automation and toolchain integration### From Source



Extract with verbose output:

```bash

harpy -u https://target.com/main.js -v## Installation```bash

```

# Clone the repository

## Command Line Options

### From Sourcegit clone https://github.com/rafabd1/DepScout.git

### Input Options

| Flag | Description | Default |cd DepScout

|------|-------------|---------|

| `-u` | Single target URL or file path | - |```bash

| `-f` | File containing list of targets (one per line) | - |

| `-d` | Directory to scan for JavaScript/TypeScript/HTML/JSON files | - |# Clone the repository# Build the binary



### Analysis Options  git clone https://github.com/rafabd1/Harpy.gitgo build -o depscout ./cmd/depscout

| Flag | Description | Default |

|------|-------------|---------|cd Harpy

| `-c` | Number of concurrent workers | `25` |

| `--max-file-size` | Maximum file size to process (KB) | `10240` |# Optional: Move to path (Linux/macOS)

| `--no-limit` | Disable file size restrictions | `false` |

# Build the binarysudo mv depscout /usr/local/bin/

**Note:** Harpy automatically uses intelligent hybrid analysis (Regex + AST) with smart fallbacks. No manual mode configuration needed.

go build -o harpy ./cmd/harpy```

### Network Options

| Flag | Description | Default |

|------|-------------|---------|

| `-l` | Requests per second limit per domain | `30` |# Optional: Move to path (Linux/macOS)### Using Go Install

| `-t` | Request timeout in seconds | `10` |

| `-H` | Custom headers (can be used multiple times) | - |sudo mv harpy /usr/local/bin/

| `-p` | Proxy list file | - |

| `--proxy` | Single proxy server | - |``````bash

| `--skip-verify` | Skip TLS certificate verification | `false` |

go install github.com/rafabd1/DepScout/cmd/depscout@latest

### Output Options

| Flag | Description | Default |### Using Go Install```

|------|-------------|---------|

| `-o` | Output file path | stdout |

| `--json` | Output results in JSON format | `false` |

| `-v` | Enable verbose output | `false` |```bash### Binary Releases

| `--silent` | Suppress all output except findings | `false` |

| `--no-color` | Disable colored output | `false` |go install github.com/rafabd1/Harpy/cmd/harpy@latest



## Example Usage```You can download pre-built binaries for your platform from the [releases page](https://github.com/rafabd1/DepScout/releases).



### Bug Bounty Reconnaissance

```bash

# Quick scan for endpoints and parameters### Binary Releases## Quick Start

harpy -u https://target.com/app.js



# Comprehensive directory scan with JSON output

harpy -d ./js-files --json -o findings.jsonDownload pre-built binaries for your platform from the [releases page](https://github.com/rafabd1/Harpy/releases).Scan a single JavaScript file:



# Scan with custom headers and proxy```bash

harpy -f targets.txt -H "User-Agent: Harpy/1.0" --proxy http://127.0.0.1:8080

```## Quick Startdepscout -u https://example.com/assets/app.js



### Integration with Other Tools```

```bash

# Pipe from subfinder and httpxExtract endpoints from a single JavaScript file:

subfinder -d target.com | httpx -path /assets/app.js | harpy

```bashScan a local directory of JS files using the high-accuracy deep scan mode:

# Extract endpoints and pass to ffuf

harpy -u https://target.com/main.js --json | jq -r '.results.findings[].endpoints[].path' | ffuf -u https://target.com/FUZZ -w -harpy -u https://example.com/assets/app.js```bash

```

```depscout -d /path/to/js/files --deep-scan

## Output Formats

```

### Terminal Output

```Scan local JavaScript files with AST analysis:

=== Extraction Results ===

✓ Sources processed: 1```bashScan a list of targets from a file:

✓ Domains found: 3  

✓ Endpoints found: 12harpy -d /path/to/js/files --enable-ast```bash

✓ Parameters found: 8

✓ Headers found: 4```depscout -f targets.txt

```

```

### JSON Output

```jsonScan from a list of targets and output as JSON:

{

  "metadata": {```bash## Command Line Options

    "tool": "Harpy",

    "version": "1.0.0", harpy -f targets.txt --json -o results.json

    "timestamp": "2024-01-15T10:30:00Z",

    "total_sources": 1```| Flag | Description | Default |

  },

  "results": {|------|-------------|---------|

    "findings": [

      {Extract parameters and headers in verbose mode:| `-u` | A single target URL or local file path. | - |

        "source": "https://example.com/app.js",

        "domains": ["api.example.com", "admin.example.com"],```bash| `-f` | A file containing a list of targets. | - |

        "endpoints": [

          {harpy -u https://target.com/main.js --enable-regex --enable-ast -v| `-d` | Path to a local directory to scan for `.js` and `.ts` files. | - |

            "method": "GET",

            "path": "/api/v1/users",```| `-c` | Number of concurrent workers. | `25` |

            "context": "fetch('/api/v1/users')"

          }| `-l` | Maximum requests per second per domain (in auto-adjustment mode). | `30` |

        ],

        "parameters": [## Command Line Options| `-H` | Custom header to include in all requests (can be specified multiple times). | - |

          {

            "name": "userId", | `-o` | File to write output to. | stdout |

            "type": "url",

            "context": "?userId=123"### Input Options| `-p` | File containing a list of proxies (http/https/socks5). | - |

          }

        ],| Flag | Description | Default || `-proxy` | A single proxy server (e.g. http://127.0.0.1:8080). | - |

        "headers": [

          {|------|-------------|---------|| `--deep-scan` | Enable deep scan using AST parsing (slower but more accurate). | `false` |

            "name": "Authorization",

            "context": "headers: {'Authorization': token}"| `-u` | Single target URL or local file path | - || `-json` | Enable JSON output format. | `false` |

          }

        ]| `-f` | File containing list of targets (one per line) | - || `--max-file-size` | Maximum file size to process in KB. | `10240` |

      }

    ]| `-d` | Directory to scan for JavaScript/TypeScript files | - || `--no-limit` | Disable file size limit. | `false` |

  },

  "summary": {| `--skip-verify` | Skip TLS certificate verification. | `false` |

    "total_domains": 2,

    "total_endpoints": 1, ### Analysis Options  | `-v` | Enable verbose output for debugging. | `false` |

    "total_parameters": 1,

    "total_headers": 1| Flag | Description | Default || `--silent` | Suppress all output except for findings. | `false` |

  }

}|------|-------------|---------|| `--no-color` | Disable colorized output. | `false` |

```

| `--enable-regex` | Enable regex-based pattern matching | `true` |

## Disclaimer

| `--enable-ast` | Enable AST-based JavaScript parsing | `false` |

**Usage Warning & Responsibility**

| `-c` | Number of concurrent workers | `25` |## Disclaimer

This tool is designed for security professionals, bug bounty hunters, and researchers for legitimate testing purposes only. Users must have explicit permission to test any target. The author is not responsible for any misuse or damage caused by this program.

| `--max-file-size` | Maximum file size to process (KB) | `10240` |

## Documentation

| `--no-limit` | Disable file size restrictions | `false` |**Usage Warning & Responsibility**

- [Architecture Overview](dev/HARPY-ARCHITECTURE.md) - Technical details and design decisions

- [Changelog](CHANGELOG.md) - Version history and updates



## Contributing### Network OptionsThis tool is intended for security professionals and researchers for legitimate testing purposes only. Running DepScout against a target may generate a high volume of HTTP requests. You are responsible for your actions and must have explicit permission to test any target. The author of this tool is not responsible for any misuse or damage caused by this program.



Contributions are welcome! Please feel free to submit a Pull Request.| Flag | Description | Default |



## License|------|-------------|---------|## Documentation



This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.| `-l` | Requests per second limit per domain | `30` |



---| `-H` | Custom headers (can be used multiple times) | - |- [Changelog](CHANGELOG.md) - Check the latest updates and version history.



<p align="center">| `-p` | Proxy list file | - |

    <sub>Made with 🔍 for bug bounty hunters and security researchers</sub>

</p>| `--proxy` | Single proxy server | - |## Contributing



<p align="center">| `--skip-verify` | Skip TLS certificate verification | `false` |

    <sub>Created by Rafael (github.com/rafabd1)</sub>

</p>| `--timeout` | Request timeout in seconds | `10` |Contributions are welcome! Please feel free to submit a Pull Request.



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