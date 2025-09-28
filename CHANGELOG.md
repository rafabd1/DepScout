# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [1.0.0] - 2025-09-27

### Added

- **Intelligent Hybrid Analysis Engine**: Automatic combination of regex and AST-based JavaScript parsing with smart fallbacks
- **Advanced Pattern Recognition**: Highly optimized regex patterns for extracting API endpoints, parameters, headers, and domains
- **Contextual Parameter Validation**: Smart semantic analysis to drastically reduce false positives while maintaining high accuracy
- **Multi-threaded Architecture**: High-performance concurrent processing with configurable worker pools
- **Adaptive Rate Limiting**: Smart per-domain throttling that automatically adjusts based on server responses (429, timeouts)
- **Flexible Input Sources**: Support for URLs, local files, directories, or stdin for seamless tool integration
- **Professional Output Formats**: Clean terminal display plus structured JSON for automation and CI/CD pipelines
- **Comprehensive Asset Extraction**: 
  - API endpoints with methods and context
  - URL parameters, form fields, and API parameters
  - HTTP headers and authentication tokens
  - Internal domains, subdomains, and CDN endpoints
  - Hidden admin panels and debug endpoints

### Features

- **Command Line Options**:
  - `-u`: Single target URL or file path
  - `-f`: File containing list of targets (one per line)
  - `-d`: Directory scanning for JavaScript/TypeScript/HTML/JSON files
  - `-c`: Configurable concurrent workers (default: 25)
  - `-l`: Requests per second limit per domain (default: 30)
  - `-t`: Request timeout in seconds (default: 10)
  - `-H`: Custom headers support (multiple headers allowed)
  - `-p`: Proxy list file support
  - `--proxy`: Single proxy server configuration
  - `--max-file-size`: File size limit in KB (default: 10240)
  - `--no-limit`: Disable file size restrictions
  - `--json`: JSON output format for automation
  - `-v`: Verbose output for debugging
  - `--silent`: Suppress all output except findings
  - `--no-color`: Disable colored output
  - `--skip-verify`: Skip TLS certificate verification

### Performance Metrics

- **Concurrent Processing**: Up to 25 parallel workers by default
- **File Processing**: Handles files up to 10MB by default (configurable)
- **Memory Efficiency**: Optimized for processing thousands of JavaScript files
- **Speed**: Processes typical web application assets in seconds

### Use Cases

- **Bug Bounty Reconnaissance**: Comprehensive endpoint and parameter discovery
- **Security Assessment**: Hidden asset and admin panel detection
- **API Discovery**: REST endpoint and GraphQL schema extraction
- **DevSecOps Integration**: CI/CD pipeline integration with JSON output
- **Red Team Operations**: Intelligence gathering and attack surface mapping

[1.0.0]: https://github.com/rafabd1/Harpy/releases/tag/v1.0.0 