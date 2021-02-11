package terraformtest

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"os"

	"github.com/tidwall/gjson"
)

type TFPlan struct {
	Data []byte
}

type TFResource struct {
	Metadata map[string]string
	Total    int
	Values   map[string]string
}

func Equal(tfResource TFResource, tfPlan TFPlan) bool {
	rootModuleResources := gjson.GetBytes(tfPlan.Data, "planned_values.root_module.resources")
	childModulesResources := gjson.GetBytes(tfPlan.Data, "planned_values.root_module.child_modules.#.resources")
	if tfResource.Total > 0 {
		if tfResource.Total != len(rootModuleResources.Array())+len(childModulesResources.Array()) {
			return false
		}
	}

	for k, v := range tfResource.Metadata {
		value := gjson.GetBytes([]byte(rootModuleResources.Raw), fmt.Sprintf("0.0.%s", k))
		if !value.Exists() {
			value = gjson.GetBytes([]byte(childModulesResources.Raw), fmt.Sprintf("0.0.%s", k))
			if !value.Exists() {
				return false
			}
		}
		if value.String() != v {
			return false
		}
	}

	for k, v := range tfResource.Values {
		value := gjson.GetBytes([]byte(rootModuleResources.Raw), fmt.Sprintf("0.0.values.%s", k))
		if !value.Exists() {
			value = gjson.GetBytes([]byte(childModulesResources.Raw), fmt.Sprintf("0.0.values.%s", k))
			if !value.Exists() {
				return false
			}
		}
		if value.String() != v {
			return false
		}
	}
	return true
}

func ReadTfPlan(path string) (TFPlan, error) {
	f, err := os.Open(path)
	if err != nil {
		return TFPlan{}, fmt.Errorf("cannot open file: %s", path)
	}
	reader := bufio.NewReader(f)
	tfplan, err := ioutil.ReadAll(reader)
	if err != nil {
		return TFPlan{}, fmt.Errorf("cannot read data from IO Reader: %v", err)
	}
	TFPlan := TFPlan{
		Data: tfplan,
	}
	return TFPlan, nil

}
