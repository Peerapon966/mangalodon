package apperr

type ErrorCode string

const (
	// S3 error
	InvalidPutObjectMode          ErrorCode = "INVALID_PUT_OBJECT_MODE"
	IfMatchPreconditionFailed     ErrorCode = "IF_MATCH_PRECONDITION_FAILED"
	IfNoneMatchPreconditionFailed ErrorCode = "IF_NONE_MATCH_PRECONDITION_FAILED"
	PutObjectFailed               ErrorCode = "PUT_OBJECT_FAILED"

	NoSuchKey       ErrorCode = "NO_SUCH_KEY"
	GetObjectFailed ErrorCode = "GET_OBJECT_FAILED"

	HeadObjectFailed ErrorCode = "HEAD_OBJECT_FAILED"

	DeleteObjectFailed ErrorCode = "DELETE_OBJECT_FAILED"

	// Crawler errors
	DownloadSiteConfigsFailed    ErrorCode = "DOWNLOAD_SITE_CONFIGS_FAILED"
	TargetBlockedRequest         ErrorCode = "TARGET_BLOCKED_REQUEST"
	TargetTimeout                ErrorCode = "TARGET_TIMEOUT"
	TargetRequestFailed          ErrorCode = "TARGET_REQUEST_FAILED"
	UnsupportedURLFormat         ErrorCode = "UNSUPPORTED_URL_FORMAT"
	CreateRequestFailed          ErrorCode = "CREATE_REQUEST_FAILED"
	ReadResponseFailed           ErrorCode = "READ_RESPONSE_FAILED"
	NoMatchingCrawlerFounded     ErrorCode = "NO_MATCHING_CRAWLER_FOUNDED"
	NoSuchCrawler                ErrorCode = "NO_SUCH_CRAWLER"
	MissingExtractEpisodesMethod ErrorCode = "MISSING_EXTRACT_EPISODES_METHOD"
	TitleExtractFailed           ErrorCode = "TITLE_EXTRACT_FAILED"
	ImageExtractFailed           ErrorCode = "IMAGE_EXTRACT_FAILED"
	InfoExtractFailed            ErrorCode = "INFO_EXTRACT_FAILED"
	InvalidSiteConfig            ErrorCode = "INVALID_SITE_CONFIG"

	// Crypto errors
	DecodeFailed     ErrorCode = "DECODE_FAILED"
	DecryptionFailed ErrorCode = "DECRYPTION_FAILED"
	EncryptionFailed ErrorCode = "ENCRYPTION_FAILED"

	GetSecretValueFailed ErrorCode = "GET_SECRET_VALUE_FAILED"
	GetParameterFailed   ErrorCode = "GET_PARAMETER_FAILED"

	// Task manager errors
	TaskAlreadyExists ErrorCode = "TASK_ALREADY_EXISTS"

	// General errors
	ScrapeTaskFailed     ErrorCode = "SCRAPE_TASK_FAILED"
	InvalidRequestBody   ErrorCode = "INVALID_REQUEST_BODY"
	EnvVarNotSet         ErrorCode = "ENV_VAR_NOT_SET"
	MarshalFailed        ErrorCode = "MARSHAL_FAILED"
	UnmarshalFailed      ErrorCode = "UNMARSHAL_FAILED"
	InvalidRegexPattern  ErrorCode = "INVALID_REGEX_PATTERN"
	TypeConversionFailed ErrorCode = "TYPE_CONVERSION_FAILED"
)
