package model

// ModelPricing 对应 model_pricing 表
type ModelPricing struct {
    ID                     int64    `gorm:"primaryKey;autoIncrement" json:"id"`                       // 自增主键
    ModelName              string   `gorm:"type:varchar(100);not null" json:"model_name"`            // 模型名称
    InputPricePerMillion   *float64 `gorm:"type:decimal(12,6)" json:"input_price_per_million,omitempty"` // 输入价格/百万token
    OutputPricePerMillion  *float64 `gorm:"type:decimal(12,6)" json:"output_price_per_million,omitempty"`// 输出价格/百万token
    CacheTokenPricePerMillion *float64 `gorm:"type:decimal(12,6)" json:"cache_token_price_per_million,omitempty"` // 缓存token价格/百万token
    PricePerRequest        *float64 `gorm:"type:decimal(12,4)" json:"price_per_request,omitempty"`   // 按请求计费价格
    PricingType            string   `gorm:"type:varchar(50);default:'default'" json:"pricing_type"`  // 计费类型
    Remark                 *string  `gorm:"type:varchar(255)" json:"remark,omitempty"`               // 备注
    CurrencyUnit           string   `gorm:"type:varchar(10);not null;default:'USD'" json:"currency_unit"` // 计价单位
}

func (ModelPricing) TableName() string {
    return "model_pricing"
}