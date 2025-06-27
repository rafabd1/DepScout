# DepScout

![Go Version](https://img.shields.io/github/go-mod/go-version/rafabd1/DepScout)
![Release](https://img.shields.io/github/v/release/rafabd1/DepScout?include_prereleases)
![Build Status](https://github.com/rafabd1/DepScout/workflows/Release%20DepScout/badge.svg)
![License](https://img.shields.io/badge/license-MIT-blue.svg)
![GitHub stars](https://img.shields.io/github/stars/rafabd1/DepScout?style=social)
![Go Report Card](https://goreportcard.com/badge/github.com/rafabd1/DepScout)

<div align="center">
<pre>
  _____             _____                 _   
 |  __ \           / ____|               | |  
 | |  | | ___ _ __| (___   ___ ___  _   _| |_ 
 | |  | |/ _ \ '_ \\___ \ / __/ _ \| | | | __|
 | |__| |  __/ |_) |___) | (_| (_) | |_| | |_ 
 |_____/ \___| .__/_____/ \___\___/ \__,_|\__|
             | |                              
             |_|                              
</pre>
</div>

<p align="center">
    <b>A high-performance, concurrent scanner for detecting Dependency Confusion vulnerabilities.</b>
</p>

## Features

- **Dependency Confusion Scanning**: Identifies potential Dependency Confusion vulnerabilities by scanning JavaScript files for package names and checking their existence on the public npm registry.
- **Dual Parsing Engine**:
  - **Regex Mode**: Fast, high-performance scanning using fine-tuned regular expressions to find `require()` and `import` statements.
  - **Deep Scan Mode (`--deep-scan`)**: High-accuracy scanning using a full Abstract Syntax Tree (AST) parser to eliminate false positives from non-code contexts (e.g., comments). Includes an automatic fallback to regex if AST parsing fails.
- **Concurrent Architecture**: Leverages a powerful, concurrent worker model to perform high-speed fetching and analysis of hundreds or thousands of files simultaneously.
- **Adaptive Rate Limiting**: Features a smart, per-domain rate limiter that automatically adjusts request speeds based on server responses (e.g., `429 Too Many Requests`), maximizing speed without overwhelming targets.
- **Flexible Input**: Accepts targets from single URLs (`-u`), files (`-f`), local directories (`-d`), or piped via `stdin`.
- **Advanced Configuration**: Provides a rich set of CLI flags for fine-grained control over concurrency, timeouts, file sizes, and more.
- **Multiple Output Formats**: Delivers results in human-readable text or machine-readable `JSON` (`--json`).

## Installation

### From Source

```bash
# Clone the repository
git clone https://github.com/rafabd1/DepScout.git
cd DepScout

# Build the binary
go build -o depscout ./cmd/depscout

# Optional: Move to path (Linux/macOS)
sudo mv depscout /usr/local/bin/
```

### Using Go Install

```bash
go install github.com/rafabd1/DepScout/cmd/depscout@latest
```

### Binary Releases

You can download pre-built binaries for your platform from the [releases page](https://github.com/rafabd1/DepScout/releases).

## Quick Start

Scan a single JavaScript file:
```bash
depscout -u https://example.com/assets/app.js
```

Scan a local directory of JS files using the high-accuracy deep scan mode:
```bash
depscout -d /path/to/js/files --deep-scan
```

Scan a list of URLs from a file, increase concurrency, and save results to a JSON file:
```bash
cat urls.txt | depscout -c 100 -o results.json --json
```

## Command Line Options

| Flag | Description | Default |
|------|-------------|---------|
| `-u` | A single target URL or local file path. | - |
| `-f` | A file containing a list of targets. | - |
| `-d` | Path to a local directory to scan for `.js` and `.ts` files. | - |
| `-c` | Number of concurrent workers. | `25` |
| `-l` | Maximum requests per second per domain (in auto-adjustment mode). | `30` |
| `-H` | Custom header to include in all requests (can be specified multiple times). | - |
| `-o` | File to write output to. | stdout |
| `-p` | File containing a list of proxies (http/https/socks5). | - |
| `--deep-scan` | Enable deep scan using AST parsing (slower but more accurate). | `false` |
| `--json` | Enable JSON output format. | `false` |
| `--max-file-size` | Maximum file size to process in KB. | `10240` |
| `--no-limit` | Disable file size limit. | `false` |
| `--skip-verify` | Skip TLS certificate verification. | `false` |
| `-v` | Enable verbose output for debugging. | `false` |
| `--silent` | Suppress all output except for findings. | `false` |
| `--no-color` | Disable colorized output. | `false` |


## Disclaimer

**Usage Warning & Responsibility**

This tool is intended for security professionals and researchers for legitimate testing purposes only. Running DepScout against a target may generate a high volume of HTTP requests. You are responsible for your actions and must have explicit permission to test any target. The author of this tool is not responsible for any misuse or damage caused by this program.

## Documentation

- [Changelog](CHANGELOG.md) - Check the latest updates and version history.

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## License

This project is licensed under the MIT License.



<p align="center">
    <sub>Made with 🖤 by Rafael (github.com/rafabd1)</sub>
</p>

<p align="center">
    <a href="https://ko-fi.com/rafabd1" target="_blank"><img src="https://storage.ko-fi.com/cdn/kofi2.png?v=3" alt="Buy Me A Coffee" style="height: 60px !important;"></a>
</p>
