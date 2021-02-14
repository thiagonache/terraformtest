package terraformtest

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"os"

	"github.com/tidwall/gjson"
)

// TFPlan is a struct containing the terraform plan data
type TFPlan struct {
	Data []byte
}

type TFDiff struct {
	Items []TFDiffItem
}

type TFDiffItem struct {
	Got, Key, Want string
}

type TFResource struct {
	Check  map[string]string
	Filter string
}

func Equal(tfResource TFResource, tfPlan TFPlan) (TFDiff, bool) {
	returnValue := true
	tfDiff := TFDiff{}
	resources := gjson.GetBytes(tfPlan.Data, tfResource.Filter)
	for k, v := range tfResource.Check {
		value := gjson.GetBytes([]byte(resources.Raw), k)
		if !value.Exists() {
			tfDiffItem := TFDiffItem{
				Got:  "",
				Key:  k,
				Want: v,
			}
			tfDiff.Items = append(tfDiff.Items, tfDiffItem)
			returnValue = false
			continue
		}
		if value.String() != v {
			tfDiffItem := TFDiffItem{
				Got:  value.String(),
				Key:  k,
				Want: v,
			}
			tfDiff.Items = append(tfDiff.Items, tfDiffItem)
			returnValue = false
		}
	}

	return tfDiff, returnValue
}

func OutputDiff(tfDiff TFDiff) string {
	var stringDiff string
	for _, diff := range tfDiff.Items {
		stringDiff += fmt.Sprintf("key %q: want %q, got %q\n", diff.Key, diff.Want, diff.Got)
	}
	return stringDiff
}

// ReadPlanFile reads the terraform plan file
func ReadPlanFile(path string) (TFPlan, error) {
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
