package terraformtest

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"

	"github.com/savaki/jq"
)

type TFPlan struct {
	data     []byte
	jqFilter string
}

type TFResource []map[string]interface{}

// Should I use strings.TrimFunc?
func removeQuotesString(s string) string {
	if len(s) > 0 && s[0] == '"' {
		s = s[1:]
	}
	if len(s) > 0 && s[len(s)-1] == '"' {
		s = s[:len(s)-1]
	}

	return s
}

func Equal(tfResourceKey, tfResourceValue string, tfResources []byte) bool {
	data, err := ExtractPlanData(tfResources, tfResourceKey)
	if err != nil {
		fmt.Printf("cannot extract %s from JSON with resources", tfResourceKey)
		return false
	}
	value := removeQuotesString(string(data))
	return value == tfResourceValue
	// for k, v := range tfResource {
	// 	jsonValue := tfResourcesJSON[k].(string)
	// 	if v != jsonValue {
	// 		return false
	// 	}
	// }
	// for _, i := range tfResourcesJSON {
	// 	for k, v := range i {
	// 		fmt.Println("Key:", k, "Value:", v)
	// 	}
	// }
	return true
}

func CountNumberResources(tfPlan []byte, jqPath string) (int, error) {
	tfPlan, err := ExtractPlanData(tfPlan, jqPath)
	if err != nil {
		return 0, fmt.Errorf("cannot extract data from plan using %q query path", jqPath)
	}

	tfResources := make([]map[string]interface{}, 0)
	err = json.Unmarshal(tfPlan, &tfResources)
	if err != nil {
		return 0, fmt.Errorf("cannot unmarshal extracted plan data: %v", err)
	}

	return len(tfResources), nil
}

func ExtractPlanData(tfPlan []byte, jqPath string) ([]byte, error) {
	jqOp, err := jq.Parse(jqPath)
	if err != nil {
		return nil, fmt.Errorf("invalid jq query: %q", jqPath)
	}

	value, err := jqOp.Apply(tfPlan)
	if err != nil {
		return nil, fmt.Errorf("cannot apply query %q on tfPlan data", jqPath)
	}

	return value, nil
}

func ParseJSON(path string) ([]byte, error) {
	f, err := os.Open(path)
	if err != nil {
		log.Fatalf("cannot open file: %s", path)
	}
	reader := bufio.NewReader(f)
	tfPlan, err := ioutil.ReadAll(reader)
	if err != nil {
		return nil, fmt.Errorf("cannot read data from IO Reader: %v", err)
	}
	return tfPlan, nil

}
