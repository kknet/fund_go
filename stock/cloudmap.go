package stock

// GetCloudMap /* 全市场云图 */
func GetCloudMap() []map[string]interface{} {
	ids := GetIndustryData("sw")
	return ids
}
