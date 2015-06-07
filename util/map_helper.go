package util

type MapHelper map[string]interface{}
type MapHelpers []MapHelper

// MapHelperFromInterface takes an interface{}, casts it to a map[string]interface{} and converts it to a MapHelper object.
func MapHelperFromInterface(data interface{}) MapHelper {
	return MapHelper(data.(map[string]interface{}))
}

// MapHelpersFromArray takes an interface{}, casts it to a []interface{} and converts it to a MapHelpers object.
func MapHelpersFromArray(data interface{}) MapHelpers {
	dataArr := data.([]interface{})
	ret := make(MapHelpers, len(dataArr))
	for idx, obj := range dataArr {
		ret[idx] = MapHelperFromInterface(obj)
	}
	return ret
}

// Float assumes and casts the value at key as a float64.
func (this MapHelper) Float(key string) float64 { return this[key].(float64) }

// Int assumes and casts the value at key as a int64.
func (this MapHelper) Int(key string) int { return QuickIntRound(this[key].(float64)) }

// String assumes and casts the value at key as a string.
func (this MapHelper) String(key string) string { return this[key].(string) }

// Array assumes and converts the value at key as a MapHelpers.
func (this MapHelper) Array(key string) MapHelpers { return MapHelpersFromArray(this[key]) }

// Map assumes and converts the value at key as a MapHelper.
func (this MapHelper) Map(key string) MapHelper { return MapHelperFromInterface(this[key]) }
