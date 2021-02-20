package terraformtest

import (
	"fmt"
	"io"
	"os"

	"github.com/tidwall/gjson"
)

// LoopControl is a struct containing items to control for loop to process json file
type LoopControl struct {
	CurDepth, MaxDepth          int
	CurItemIndex, CurItemSubKey string
}

// TFPlan is a struct containing the terraform plan data
type TFPlan struct {
	Data        []byte
	Items       map[string]TFResultResource
	LoopControl LoopControl
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
	Address  string
	Metadata map[string]string
	Values   map[string]string
}

// TFResultResource is a map to store the Metadata and Values items to make easier to find resource items.
type TFResultResource map[string]map[string]gjson.Result

// New instantiate a new TFPlan object and returns a pointer to it.
func New(planPath string) (*TFPlan, error) {
	tfp := &TFPlan{
		LoopControl: LoopControl{MaxDepth: 10},
		Items:       map[string]TFResultResource{},
	}

	f, err := os.Open(planPath)
	if err != nil {
		return tfp, fmt.Errorf("cannot open file: %s", planPath)
	}
	plan, err := io.ReadAll(f)
	if err != nil {
		return tfp, fmt.Errorf("cannot read data from IO Reader: %v", err)
	}

	tfp.Data = plan
	tfp.Coalesce()

	return tfp, nil
}

// Coalesce transform the multi level json into one big object to make queries easier
func (tfPlan *TFPlan) Coalesce() {
	rootModule := gjson.GetBytes(tfPlan.Data, `planned_values.root_module|@pretty:{"sortKeys":true}`)
	rootModule.ForEach(tfPlan.coalescePlan)
}

func (tfPlan *TFPlan) coalescePlan(key, value gjson.Result) bool {
	if tfPlan.LoopControl.CurDepth > tfPlan.LoopControl.MaxDepth {
		fmt.Println("MaxDepth reached")
		return false
	}

	switch key.String() {
	case "resources":
		tfPlan.LoopControl.CurDepth++
		for _, child := range value.Array() {
			child.ForEach(tfPlan.coalescePlan)
		}
	case "child_modules":
		tfPlan.LoopControl.CurDepth++
		for _, child := range value.Array() {
			child.ForEach(tfPlan.coalescePlan)
		}
	case "values":
		tfPlan.LoopControl.CurItemSubKey = "Values"
		_, ok := tfPlan.Items[tfPlan.LoopControl.CurItemSubKey]
		if !ok {
			tfPlan.Items[tfPlan.LoopControl.CurItemSubKey] = map[string]map[string]gjson.Result{}
		}
		tfPlan.Items[tfPlan.LoopControl.CurItemSubKey][tfPlan.LoopControl.CurItemIndex] = map[string]gjson.Result{}
		value.ForEach(tfPlan.coalescePlan)
	default:
		if key.String() == "address" {
			tfPlan.LoopControl.CurItemSubKey = "Metadata"
			tfPlan.LoopControl.CurItemIndex = value.String()
			_, ok := tfPlan.Items[tfPlan.LoopControl.CurItemSubKey]
			if !ok {
				tfPlan.Items[tfPlan.LoopControl.CurItemSubKey] = map[string]map[string]gjson.Result{}
			}
			tfPlan.Items[tfPlan.LoopControl.CurItemSubKey][tfPlan.LoopControl.CurItemIndex] = map[string]gjson.Result{}
			break
		}
		tfPlan.Items[tfPlan.LoopControl.CurItemSubKey][tfPlan.LoopControl.CurItemIndex][key.String()] = value
		//fmt.Printf("Add key %v and value %v into %v into %v\n\n", key, value, tfPlan.CurItemIndex, tfPlan.CurItemSubKey)
	}

	return true
}

// Equal evaluate TFPlan and TFTestResource and returns the diff and if it is equal
// or not.
func Equal(tfTestResource TFTestResource, tfPlan TFPlan) (TFDiff, bool) {
	tfDiff := TFDiff{}
	resource, ok := tfPlan.Items["Metadata"][tfTestResource.Address]
	if !ok {
		tfDiffItem := TFDiffItem{
			Got:  "does not exist",
			Key:  tfTestResource.Address,
			Want: "exist",
		}
		tfDiff.Items = append(tfDiff.Items, tfDiffItem)

		return tfDiff, false
	}
	for k, v := range tfTestResource.Metadata {
		value, ok := resource[k]
		if !ok {
			tfDiffItem := TFDiffItem{
				Got:  "",
				Key:  k,
				Want: v,
			}
			tfDiff.Items = append(tfDiff.Items, tfDiffItem)

			return tfDiff, false
		}
		if value.String() != v {
			tfDiffItem := TFDiffItem{
				Got:  value.String(),
				Key:  k,
				Want: v,
			}
			tfDiff.Items = append(tfDiff.Items, tfDiffItem)

			return tfDiff, false
		}
	}

	return tfDiff, true
}

// Diff returns all diffs in a string concanated by new line
func Diff(tfDiff TFDiff) string {
	var stringDiff string
	for _, diff := range tfDiff.Items {
		stringDiff += fmt.Sprintf("key %q: want %q, got %q\n", diff.Key, diff.Want, diff.Got)
	}
	return stringDiff
}
