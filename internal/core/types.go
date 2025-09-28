package core

/**
 * @description Core data structures for Harpy's finding system
 */

// FileType represents the type of file being analyzed
type FileType int

const (
	Unknown FileType = iota
	JavaScript
	TypeScript
	HTML
	JSON
	XML
	CSS
)

func (ft FileType) String() string {
	return [...]string{"unknown", "javascript", "typescript", "html", "json", "xml", "css"}[ft]
}

// ParamType represents the type/location of a parameter
type ParamType int

const (
	URLParam ParamType = iota
	PostParam
	HeaderParam
	CookieParam
	BodyParam
)

func (pt ParamType) String() string {
	return [...]string{"url", "post", "header", "cookie", "body"}[pt]
}

/**
 * @description Main finding structure containing all extracted data with source context
 */
type Finding struct {
	Source      SourceInfo    `json:"source"`
	Domains     []string      `json:"domains"`
	Endpoints   []Endpoint    `json:"endpoints"`
	Parameters  []Parameter   `json:"parameters"`
	Headers     []Header      `json:"headers"`
	Bodies      []RequestBody `json:"bodies"`
}

/**
 * @description Source information for traceability and context
 */
type SourceInfo struct {
	FilePath string `json:"file_path"`
	FileType string `json:"file_type"`
}

/**
 * @description Endpoint representation with method and context
 */
type Endpoint struct {
	Method        string   `json:"method"`
	Path          string   `json:"path"`
	Context       string   `json:"context"`
	RelatedParams []string `json:"related_params,omitempty"`
}

/**
 * @description Parameter with type classification and context
 */
type Parameter struct {
	Name           string    `json:"name"`
	Type           ParamType `json:"type"`
	Context        string    `json:"context"`
	DataType       string    `json:"data_type"`
	PossibleValues []string  `json:"possible_values,omitempty"`
}

/**
 * @description Header information with pattern matching
 */
type Header struct {
	Name             string   `json:"name"`
	ValuePattern     string   `json:"value_pattern"`
	Context          string   `json:"context"`
	RelatedEndpoints []string `json:"related_endpoints,omitempty"`
}

/**
 * @description Request body structure information
 */
type RequestBody struct {
	ContentType string      `json:"content_type"`
	Structure   interface{} `json:"structure"`
	Context     string      `json:"context"`
}

/**
 * @description Raw finding structure used during extraction process
 */
type RawFinding struct {
	Domains     []string
	Endpoints   []Endpoint
	Parameters  []Parameter
	Headers     []Header
	Bodies      []RequestBody
	Context     string
	Confidence  float64
}

/**
 * @description Enhanced finding with additional context from AST analysis
 */
type EnhancedFinding struct {
	RawFinding
	ASTContext    interface{} // Additional AST-derived context
	Relationships []Relationship
}

/**
 * @description Relationship between different extracted elements
 */
type Relationship struct {
	Type        string      `json:"type"`
	From        interface{} `json:"from"`
	To          interface{} `json:"to"`
	Description string      `json:"description"`
}

/**
 * @description Context information for proximity-based analysis
 */
type ExtractionContext struct {
	SourceURL     string
	Content       []byte
	WindowStart   int
	WindowEnd     int
	FileType      FileType
	IsMinified    bool
}

/**
 * @description Pattern match result with confidence scoring
 */
type PatternMatch struct {
	Pattern    string
	Match      string
	Groups     []string
	Position   int
	Context    string
	Confidence float64
}

/**
 * @description Extraction strategy interface for different processing approaches
 */
type ExtractionStrategy interface {
	Extract(content []byte, context ExtractionContext) (*RawFinding, error)
	CanHandle(context ExtractionContext) bool
	GetName() string
}

/**
 * @description Result statistics for analysis reporting
 */
type AnalysisStats struct {
	TotalFiles       int            `json:"total_files"`
	ProcessedFiles   int            `json:"processed_files"`
	FailedFiles      int            `json:"failed_files"`
	TotalEndpoints   int            `json:"total_endpoints"`
	TotalParameters  int            `json:"total_parameters"`
	TotalDomains     int            `json:"total_domains"`
	FileTypeBreakdown map[string]int `json:"file_type_breakdown"`
	ProcessingTime   string         `json:"processing_time"`
}

/**
 * @description Configuration for pattern matching behavior
 */
type PatternConfig struct {
	EnableRegex        bool
	EnableAST          bool
	EnableContextMap   bool
	MinConfidence      float64
	MaxMatches         int
	WindowSize         int
	IgnoreMinified     bool
	CustomPatterns     []string
}

/**
 * @description Domain classification for intelligent processing
 */
type DomainInfo struct {
	Name        string
	IsAPI       bool
	IsInternal  bool
	IsAdmin     bool
	Technology  string
	Confidence  float64
}