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
	PrevItemIndex               string
}

// Test is the main struct containing the test data
type Test struct {
	PlanData     []byte
	LoopControl  LoopControl
	ResourcesSet ResourceSet
}

// CompDiff is a struct containing slice of CompDiffItem
type CompDiff struct {
	Items []CompDiffItem
}

// CompDiffItem is a struct containing got, key and want values for the diff
type CompDiffItem struct {
	Got, Key, Want string
}

type Resource struct {
	Address  string
	Metadata map[string]string
	Values   map[string]string
}

// ResourceSet is a map to store the Metadata and Values items to make easier to find resource items.
type ResourceSet struct {
	Items    map[string]map[string]map[string]gjson.Result
	CompDiff CompDiff
}

func ReadPlan(planPath string) (*Test, error) {
	tf := &Test{
		LoopControl: LoopControl{MaxDepth: 10},
		ResourcesSet: ResourceSet{
			Items:    map[string]map[string]map[string]gjson.Result{},
			CompDiff: CompDiff{},
		},
	}

	f, err := os.Open(planPath)
	if err != nil {
		return tf, fmt.Errorf("cannot open file: %s", planPath)
	}
	defer f.Close()

	plan, err := io.ReadAll(f)
	if err != nil {
		return tf, fmt.Errorf("cannot read data from IO Reader: %v", err)
	}

	tf.PlanData = plan
	tf.ResourceSet()

	return tf, nil
}

func (rs *ResourceSet) Contains(r Resource) bool {
	metadata, ok := rs.Items[r.Address]["Metadata"]
	if !ok {
		item := CompDiffItem{
			Got:  "does not exist",
			Key:  r.Address,
			Want: "exist",
		}
		rs.CompDiff.Items = append(rs.CompDiff.Items, item)

		return false
	}
	for k, v := range r.Metadata {
		valueFound, ok := metadata[k]
		if !ok {
			item := CompDiffItem{
				Got:  "",
				Key:  k,
				Want: v,
			}
			rs.CompDiff.Items = append(rs.CompDiff.Items, item)
			return false
		}
		if valueFound.String() != v {
			item := CompDiffItem{
				Got:  valueFound.String(),
				Key:  k,
				Want: v,
			}
			rs.CompDiff.Items = append(rs.CompDiff.Items, item)
			return false
		}
	}

	values, ok := rs.Items[r.Address]["Values"]
	if !ok {
		item := CompDiffItem{
			Got:  "does not exist",
			Key:  r.Address,
			Want: "exist",
		}
		rs.CompDiff.Items = append(rs.CompDiff.Items, item)

		return false
	}
	for k, v := range r.Values {
		valueFound, ok := values[k]
		if !ok {
			item := CompDiffItem{
				Got:  "",
				Key:  k,
				Want: v,
			}
			rs.CompDiff.Items = append(rs.CompDiff.Items, item)

			return false
		}
		if valueFound.String() != v {
			item := CompDiffItem{
				Got:  valueFound.String(),
				Key:  k,
				Want: v,
			}
			rs.CompDiff.Items = append(rs.CompDiff.Items, item)

			return false
		}
	}

	return true
}

func (resourceSet ResourceSet) Diff() string {
	var stringDiff string
	for _, diff := range resourceSet.CompDiff.Items {
		stringDiff += fmt.Sprintf("key %q: want %q, got %q\n", diff.Key, diff.Want, diff.Got)
	}
	return stringDiff
}

// ResourceSet transform the multi level json into one big object to make queries easier
func (tfPlan *Test) ResourceSet() {
	rootModule := gjson.GetBytes(tfPlan.PlanData, `planned_values.root_module|@pretty:{"sortKeys":true}`)
	rootModule.ForEach(tfPlan.resourceSet)
}

func (tfPlan *Test) resourceSet(key, value gjson.Result) bool {
	if tfPlan.LoopControl.CurDepth > tfPlan.LoopControl.MaxDepth {
		fmt.Println("MaxDepth reached")
		return false
	}

	switch key.String() {
	case "resources":
		tfPlan.LoopControl.PrevItemIndex = "resources"
		tfPlan.LoopControl.CurDepth++
		for _, child := range value.Array() {
			child.ForEach(tfPlan.resourceSet)
		}
	case "child_modules":
		tfPlan.LoopControl.PrevItemIndex = "child_modules"
		tfPlan.LoopControl.CurDepth++
		for _, child := range value.Array() {
			child.ForEach(tfPlan.resourceSet)
		}
	case "values":
		tfPlan.LoopControl.CurItemSubKey = "Values"
		_, ok := tfPlan.ResourcesSet.Items[tfPlan.LoopControl.CurItemIndex]
		if !ok {
			tfPlan.ResourcesSet.Items[tfPlan.LoopControl.CurItemIndex] = map[string]map[string]gjson.Result{}
		}
		tfPlan.ResourcesSet.Items[tfPlan.LoopControl.CurItemIndex][tfPlan.LoopControl.CurItemSubKey] = map[string]gjson.Result{}
		value.ForEach(tfPlan.resourceSet)
	case "address":
		// We are only interested in resources address
		if tfPlan.LoopControl.PrevItemIndex != "resources" {
			break
		}
		tfPlan.LoopControl.CurItemSubKey = "Metadata"
		tfPlan.LoopControl.CurItemIndex = value.String()
		_, ok := tfPlan.ResourcesSet.Items[tfPlan.LoopControl.CurItemIndex]
		if !ok {
			tfPlan.ResourcesSet.Items[tfPlan.LoopControl.CurItemIndex] = map[string]map[string]gjson.Result{}
		}
		tfPlan.ResourcesSet.Items[tfPlan.LoopControl.CurItemIndex][tfPlan.LoopControl.CurItemSubKey] = map[string]gjson.Result{}

	default:
		tfPlan.ResourcesSet.Items[tfPlan.LoopControl.CurItemIndex][tfPlan.LoopControl.CurItemSubKey][key.String()] = value
		//fmt.Printf("Add key %v and value %v into %v into %v\n\n", key, value, tfPlan.LoopControl.CurItemIndex, tfPlan.LoopControl.CurItemSubKey)
	}

	return true
}

// Equal evaluate TFTest and TFResource and returns the diff and if it is equal
// or not.
// func Equal(tfTestResource TFTestResource, tfPlan TFTest) (TFDiff, bool) {
// 	tfDiff := TFDiff{}
// 	resource, ok := tfPlan.ResourcesSet["Metadata"][tfTestResource.Address]
// 	if !ok {
// 		item := TFCompDiffItem{
// 			Got:  "does not exist",
// 			Key:  tfTestResource.Address,
// 			Want: "exist",
// 		}
// 		tfDiff.Items = append(tfDiff.Items, item)

// 		return tfDiff, false
// 	}
// 	for k, v := range tfTestResource.Metadata {
// 		value, ok := resource[k]
// 		if !ok {
// 			item := TFDiffItem{
// 				Got:  "",
// 				Key:  k,
// 				Want: v,
// 			}
// 			tfDiff.Items = append(tfDiff.Items, item)

// 			return tfDiff, false
// 		}
// 		if value.String() != v {
// 			item := TFDiffItem{
// 				Got:  value.String(),
// 				Key:  k,
// 				Want: v,
// 			}
// 			tfDiff.Items = append(tfDiff.Items, item)

// 			return tfDiff, false
// 		}
// 	}

// 	return tfDiff, true
// }

// Diff returns all diffs in a string concanated by new line
func Diff(tfDiff CompDiff) string {
	var stringDiff string
	for _, diff := range tfDiff.Items {
		stringDiff += fmt.Sprintf("key %q: want %q, got %q\n", diff.Key, diff.Want, diff.Got)
	}
	return stringDiff
}
