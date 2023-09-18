package util

import (
	"encoding/json"
	"fmt"
)

func LogJson(jsonData any) {
	str, _ := json.MarshalIndent(jsonData, "", "\t")
	fmt.Println(string(str))
}
