package coupon

import (
	// "database/sql"
	"encoding/json"
	// "fmt"
	"log"
)

// GetPropertiesString 获取属性（map[string]string）的字符串格式。
func (c *Coupon) GetPropertiesString() (*string, error) {
	jsonBytes, err := json.Marshal(c.Properties)
	if	nil != err {
		log.Println(err)  
		return nil, err
	} 
	var str = string(jsonBytes)
	return &str, nil
}

// SetPropertiesFromString 根据string来设置Coupon属性（map[string]string）
func (c *Coupon) SetPropertiesFromString(properties string) error{
	err := json.Unmarshal([]byte(properties), &c.Properties)
	if	nil != err {
		log.Println(err)  
		return err
	} 
	return nil
}

// GetRules 获取卡券的规则
func (c *Coupon) GetRules() map[string]interface{} {
	var rules map[string]interface{}
	var ok bool
	if rules, ok = c.Properties[KeyBindingRuleProperties].(map[string]interface{}); ok {
		return rules
	}
	// log.Println("============if rules, ok============" )  
	// log.Println(ok )  
	return nil
}