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
	CurItemIndex string
	Data         []byte
	Items        map[string]map[string]gjson.Result
	MaxDepth     int
}

// TFDiff is a struct containing slice of TFDiffItem
type TFDiff struct {
	Items []TFDiffItem
}

// TFDiffItem is a struct containing got, key and want values for the diff
type TFDiffItem struct {
	Got, Key, Want string
}

// TFResource is a struct with values to be checked and JSON query filter
type TFResource struct {
	Check  map[string]string
	Filter string
}

// Equal evaluate TFPlan and TFResource and returns the diff and if it is equal
// or not.
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

// OutputDiff prints out all diffs in a string concanated by new line
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
		Data:     tfplan,
		MaxDepth: 10,
	}
	return TFPlan, nil

}
