package utils

import (
	"fmt"
	"math"
)

const (
	EarthRadius = 6371000 // 地球半径（米）
)

// 坐标点结构
type Location struct {
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
}

// CalculateDistance 计算两点之间的距离（米）
func CalculateDistance(lat1, lon1, lat2, lon2 float64) float64 {
	// 转换为弧度
	lat1 = toRadians(lat1)
	lon1 = toRadians(lon1)
	lat2 = toRadians(lat2)
	lon2 = toRadians(lon2)

	// Haversine公式
	dLat := lat2 - lat1
	dLon := lon2 - lon1

	a := math.Sin(dLat/2)*math.Sin(dLat/2) +
		math.Cos(lat1)*math.Cos(lat2)*
			math.Sin(dLon/2)*math.Sin(dLon/2)

	c := 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))

	return EarthRadius * c
}

// GetBoundingBox 获取某个点指定半径范围内的边界框
func GetBoundingBox(lat, lon float64, radiusMeters float64) (minLat, minLon, maxLat, maxLon float64) {
	// 计算纬度范围
	latRadian := radiusMeters / EarthRadius
	latDelta := toDegrees(latRadian)

	// 计算经度范围
	lonRadian := math.Asin(math.Sin(radiusMeters/EarthRadius) / math.Cos(toRadians(lat)))
	lonDelta := toDegrees(lonRadian)

	return lat - latDelta, lon - lonDelta, lat + latDelta, lon + lonDelta
}

// IsPointInCircle 判断一个点是否在圆形范围内
func IsPointInCircle(centerLat, centerLon, pointLat, pointLon, radiusMeters float64) bool {
	distance := CalculateDistance(centerLat, centerLon, pointLat, pointLon)
	return distance <= radiusMeters
}

// FormatLocation 格式化位置信息
func FormatLocation(lat, lon float64) string {
	return fmt.Sprintf("%.6f,%.6f", lat, lon)
}

// ParseLocation 解析位置字符串
func ParseLocation(location string) (*Location, error) {
	var lat, lon float64
	_, err := fmt.Sscanf(location, "%f,%f", &lat, &lon)
	if err != nil {
		return nil, err
	}
	return &Location{Latitude: lat, Longitude: lon}, nil
}

// toRadians 将角度转换为弧度
func toRadians(degrees float64) float64 {
	return degrees * math.Pi / 180
}

// toDegrees 将弧度转换为角度
func toDegrees(radians float64) float64 {
	return radians * 180 / math.Pi
}
