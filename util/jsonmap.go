package util

type JsonMap map[string]interface{}

func JsonMapFromInterface(data interface{}) *JsonMap {
	ret := JsonMap(data.(map[string]interface{}))
	return &ret
}
func (this *JsonMap) Float(key string) float64       { return (*this)[key].(float64) }
func (this *JsonMap) Int(key string) int             { return QuickIntRound((*this)[key].(float64)) }
func (this *JsonMap) String(key string) string       { return (*this)[key].(string) }
func (this *JsonMap) Array(key string) []interface{} { return (*this)[key].([]interface{}) }
func (this *JsonMap) Map(key string) *JsonMap        { return JsonMapFromInterface((*this)[key]) }
