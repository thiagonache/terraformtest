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
	CurDepth, MaxDepth int
	CurItemIndex       string
	Data               []byte
	Items              map[string]map[string]gjson.Result
}

// TFDiff is a struct containing slice of TFDiffItem
type TFDiff struct {
	Items []TFDiffItem
}

// TFDiffItem is a struct containing got, key and want values for the diff
type TFDiffItem struct {
	Got, Key, Want string
}

// TFTestResource is a struct with values to be checked and JSON query filter
type TFTestResource struct {
	Check  map[string]string
	Filter string
}

// NewTerraformTest instantiate a new TFPlan object and returns a pointer to it.
func NewTerraformTest(planPath string) (*TFPlan, error) {
	tfp := &TFPlan{
		CurItemIndex: "",
		Data:         []byte{},
		Items:        map[string]map[string]gjson.Result{},
		MaxDepth:     10,
	}

	f, err := os.Open(planPath)
	if err != nil {
		return tfp, fmt.Errorf("cannot open file: %s", planPath)
	}
	reader := bufio.NewReader(f)
	plan, err := ioutil.ReadAll(reader)
	if err != nil {
		return tfp, fmt.Errorf("cannot read data from IO Reader: %v", err)
	}

	tfp.Data = plan

	return tfp, nil
}

// CoalescePlan transform the multi level json into one big object to make queries easier
func CoalescePlan(tfPlan *TFPlan, key string, value gjson.Result) bool {
	tfPlan.CurDepth++
	if tfPlan.CurDepth > tfPlan.MaxDepth {
		return false
	}

	switch key {
	case "resources":
		for _, child := range value.Array() {
			for k, v := range child.Map() {
				CoalescePlan(tfPlan, k, v)
			}
		}
	case "child_modules":
		for _, child := range value.Array() {
			for k, v := range child.Map() {
				CoalescePlan(tfPlan, k, v)
			}
		}
	default:
		if key == "address" {
			tfPlan.CurItemIndex = value.String()
			break
		}
		item := make(map[string]map[string]gjson.Result)
		item[tfPlan.CurItemIndex] = make(map[string]gjson.Result)
		item[tfPlan.CurItemIndex][key] = value
		tfPlan.Items = item
	}

	return true
}

// Equal evaluate TFPlan and TFTestResource and returns the diff and if it is equal
// or not.
func Equal(tfTestResource TFTestResource, tfPlan TFPlan) (TFDiff, bool) {
	returnValue := true
	tfDiff := TFDiff{}
	resources := gjson.GetBytes(tfPlan.Data, tfTestResource.Filter)
	for k, v := range tfTestResource.Check {
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

// Diff returns all diffs in a string concanated by new line
func Diff(tfDiff TFDiff) string {
	var stringDiff string
	for _, diff := range tfDiff.Items {
		stringDiff += fmt.Sprintf("key %q: want %q, got %q\n", diff.Key, diff.Want, diff.Got)
	}
	return stringDiff
}
