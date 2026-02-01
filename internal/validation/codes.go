package validation

// Flow Codes (FLOW_###)
const (
	ErrFlowCycle          = "FLOW_001"
	ErrFlowEmpty          = "FLOW_002"
	ErrFlowInvalidTimeout = "FLOW_003"
)

// Node Codes (NODE_###)
const (
	ErrNodeDefNotFound   = "NODE_001"
	ErrNodeMissingInput  = "NODE_002"
	ErrNodeMissingOutput = "NODE_003"
	ErrNodeInvalidTier   = "NODE_004"
	ErrNodeMissingFlag   = "NODE_005"
	ErrNodeUnknownFlag   = "NODE_006"
	ErrNodeDuplicateID   = "NODE_007"
	ErrNodeMissingID     = "NODE_008"
)

// Def Codes (DEF_###)

// Edge Codes (EDGE_###)
const (
	ErrEdgeNoProducer   = "EDGE_001"
	ErrEdgeNoConsumer   = "EDGE_002"
	ErrEdgeTypeMismatch = "EDGE_003"
	ErrEdgeNoDatawell   = "EDGE_004"
	ErrEdgeInvalidName  = "EDGE_005"
)

// DataWells Codes (WELL_###)
const (
	ErrDatawellInvalidSource       = "WELL_001"
	ErrDatawellEdgeNotFound        = "WELL_002"
	ErrDatawellDuplicateEdge       = "WELL_003"
	ErrDatawellStoreNotFound       = "WELL_004"
	ErrDatawellInvalidKeyTemplate  = "WELL_005"
	ErrDatawellStaticMissingKey    = "WELL_006"
	ErrDatawellStaticHasDynamicKey = "WELL_007"
	ErrDatawellDynamicHasStaticKey = "WELL_008"
)

// Store Codes (STORE_###)
const (
	ErrStoreNotFound      = "STORE_001"
	ErrStoreInvalidType   = "STORE_002"
	ErrStoreInvalidParams = "STORE_003"
	ErrStoreUnused        = "STORE_004" // warning
)

// Type Codes (TYPE_###)
const (
	ErrTypeInvalidMime = "TYPE_001"
	ErrTypeEmptySet    = "TYPE_002"
)

// General Warns (WARN_###
const (
	WarnDeprecatedField = "WARN_001"
)
