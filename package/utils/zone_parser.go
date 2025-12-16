package utils

import "encoding/json"

func ParseZoneGeoJSON(
	raw json.RawMessage,
) (string, [][]float64, float64, float64, float64) {

	var polygon [][]float64
	var lat, lng, radius float64

	if len(raw) == 0 {
		return "", polygon, lat, lng, radius
	}

	var fc map[string]interface{}
	if err := json.Unmarshal(raw, &fc); err != nil {
		return "", polygon, lat, lng, radius
	}

	features, ok := fc["features"].([]interface{})
	if !ok || len(features) == 0 {
		return "", polygon, lat, lng, radius
	}

	feature := features[0].(map[string]interface{})
	geom := feature["geometry"].(map[string]interface{})
	geomType := geom["type"].(string)

	switch geomType {

	case "Polygon":
		ring := geom["coordinates"].([]interface{})[0].([]interface{})
		for _, p := range ring {
			point := p.([]interface{})
			polygon = append(polygon, []float64{
				point[0].(float64),
				point[1].(float64),
			})
		}
		return "polygon", polygon, 0, 0, 0

	case "Point":
		point := geom["coordinates"].([]interface{})
		lng = point[0].(float64)
		lat = point[1].(float64)

		if props, ok := feature["properties"].(map[string]interface{}); ok {
			if r, ok := props["radius"].(float64); ok {
				radius = r
			}
		}
		return "circle", nil, lat, lng, radius
	}

	return "", polygon, lat, lng, radius
}
