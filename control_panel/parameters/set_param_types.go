package parameters

type SetParamCommand struct {
	ParamName string      `json:"paramName"`
	Param     interface{} `json:"param"`
}
