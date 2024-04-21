package public

const (
	ValidatorKey        = "ValidatorKey"
	TranslatorKey       = "TranslatorKey"
	AdminSessionInfoKey = "AdminSessionInfoKey"

	LoadTypeHTTP = 0
	LoadTypeTCP  = 1
	LoadTypeGRPC = 2

	RuleTypePrefixURL = 0
	RuleTypeDomin     = 1

	RedisFlowDayKey  = "flow_day_count"
	RedisFlowHourKey = "flow_hour_count"

	FlowTotal              = "flow_total"
	FlowCountServicePrefix = "flow_service_"
	FlowCountAppPrefix     = "flow_app_"
	FlowServicePrefix      = "flow_service_"
	FlowAppPrefix          = "flow_app_"

	JwtSignKey = "my_sign_key"
	JwtExpires = 60 * 60

	StaticConfig    = 0
	ZookeeperConfig = 1
	EtcdConfig      = 2
)

var (
	LoadTypeMap = map[int]string{
		LoadTypeHTTP: "HTTP",
		LoadTypeTCP:  "TCP",
		LoadTypeGRPC: "GRPC",
	}
)
